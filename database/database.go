package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type DB struct {
	cache    *redis.Client
	database *gorm.DB
}

func ConnectDB() (*DB, error) {
	cache, err := connectRedis()
	if err != nil {
		return nil, err
	}
	database, err := connectMysql()
	if err != nil {
		return nil, err
	}
	return &DB{cache: cache, database: database}, nil
}

func (db *DB) SetUser(username string, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	user := &User{
		Username: username,
		Avatar:   1,
		Hash:     string(hash),
	}
	if err := db.database.Create(user).Error; err != nil {
		return err
	}
	if err = db.cache.HSet(context.Background(), fmt.Sprintf("user:%s", username), "username", user.Username, "avatar", user.Avatar, "hash", user.Hash).Err(); err != nil {
		return err
	}
	return nil
}

func (db *DB) GetUser(username string) (*User, error) {
	var user User
	if err := db.cache.HGetAll(context.Background(), fmt.Sprintf("user:%s", username)).Scan(&user); err != nil {
		return &User{Username: username, Avatar: user.Avatar, Hash: user.Hash}, nil
	}
	if err := db.database.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	if err := db.cache.HSet(context.Background(), fmt.Sprintf("user:%s", username), "avatar", user.Avatar, "hash", user.Hash).Err(); err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) AuthUser(username string, password string) (*User, error) {
	user, err := db.GetUser(username)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}
	return user, nil
}

func (db *DB) SetFriend(user User, target User) error {
	if err := db.database.Model(&Friend{}).Where("user_one = ? AND user_two = ?", user.Username, target.Username).Update("accepted", true).Error; err != nil {
		fmt.Println("Error updating friends data:", err)
		return err
	}
	pipe := db.cache.Pipeline()
	pipe.HDel(context.Background(), fmt.Sprintf("request:%s", user.Username), target.Username)
	pipe.HSet(context.Background(), fmt.Sprintf("friend:%s", user.Username), target.Username, target.Avatar)
	pipe.HSet(context.Background(), fmt.Sprintf("friend:%s", target.Username), user.Username, user.Avatar)
	_, err := pipe.Exec(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetRequest(user User, target string) error {
	friend := Friend{
		UserOne:  user.Username,
		UserTwo:  target,
		Accepted: false,
	}
	if err := db.database.Create(friend).Error; err != nil {
		return err
	}
	if err := db.cache.HSet(context.Background(), fmt.Sprintf("request:%s", target), user.Username, user.Avatar).Err(); err != nil {
		return err
	}
	return nil
}

func (db *DB) FriendOrRequestExist(user string, target string) (bool, error) {
	if exist, err := db.cache.HExists(context.Background(), fmt.Sprintf("friend:%s", user), target).Result(); err == nil {
		if exist {
			return true, nil
		}
	}
	if exist, err := db.cache.HExists(context.Background(), fmt.Sprintf("request:%s", user), target).Result(); err == nil {
		if exist {
			return true, nil
		}
	}
	var count int64
	if err := db.database.Where("use_one = ? AND user_two = ? OR user_one = ? AND user_two = ?", user, target, target, user).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (db *DB) GetFriends(username string) ([]User, error) {
	var friends []User
	if users, err := db.cache.HGetAll(context.Background(), fmt.Sprintf("friend:%s", username)).Result(); err == nil {
		for key, friend := range users {
			avatar, err := strconv.Atoi(friend)
			if err != nil {
				return make([]User, 0), err
			}
			friends = append(friends, User{Username: key, Avatar: avatar})
		}
		return friends, nil
	}
	db.database.Table("users AS U").
		Select("U.username, U.avatar").
		Joins("JOIN friends AS F ON CASE WHEN F.user_one = ? THEN F.user_two WHEN F.user_two = ? THEN F.user_one END = U.username", username, username).
		Where("F.accepted = ?", true).
		Group("U.username").
		Scan(&friends)
	pipe := db.cache.Pipeline()
	for _, friend := range friends {
		pipe.HSet(context.Background(), fmt.Sprintf("friend:%s", username), &friend.Avatar)
	}
	_, err := pipe.Exec(context.Background())
	if err != nil {
		return make([]User, 0), err
	}
	return friends, nil
}

func (db *DB) GetRequests(username string) ([]User, error) {
	var friends []User
	if users, err := db.cache.HGetAll(context.Background(), fmt.Sprintf("request:%s", username)).Result(); err == nil {
		for key, friend := range users {
			avatar, err := strconv.Atoi(friend)
			if err != nil {
				return make([]User, 0), err
			}
			friends = append(friends, User{Username: key, Avatar: avatar})
		}
		return friends, nil
	}
	db.database.Table("users AS U").
		Select("U.username, U.avatar").
		Joins("JOIN friends AS F ON CASE WHEN F.user_one = ? THEN F.user_two WHEN F.user_two = ? THEN F.user_one END = U.username", username, username).
		Where("F.accepted = ?", false).
		Group("U.username").
		Scan(&friends)
	pipe := db.cache.Pipeline()
	for _, friend := range friends {
		pipe.HSet(context.Background(), fmt.Sprintf("request:%s", username), &friend.Avatar)
	}
	_, err := pipe.Exec(context.Background())
	if err != nil {
		return make([]User, 0), err
	}
	return friends, nil
}

func (db *DB) DeleteFriendOrRequest(user string, target string) error {
	if err := db.database.Where("user_one = ? AND user_two = ? OR user_one = ? AND user_two = ?", user, target, target, user).Delete(&Friend{}).Error; err != nil {
		fmt.Println("Error updating friends data:", err)
		return err
	}
	pipe := db.cache.Pipeline()
	pipe.HDel(context.Background(), fmt.Sprintf("request:%s", user), target)
	pipe.HDel(context.Background(), fmt.Sprintf("friend:%s", user), target)
	pipe.HDel(context.Background(), fmt.Sprintf("friend:%s", target), user)
	_, err := pipe.Exec(context.Background())
	if err != nil {
		return err
	}
	return nil
}
