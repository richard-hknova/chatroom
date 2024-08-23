package database

import (
	"context"
	"fmt"
	"strconv"
)

func (db *DB) SetFriend(username string, avatar int, target User) error {
	if err := db.database.Model(&Friend{}).Where("user_one = ? AND user_two = ?", username, target.Username).Update("accepted", true).Error; err != nil {
		fmt.Println("Error updating friends data:", err)
		return err
	}
	pipe := db.cache.Pipeline()
	pipe.HDel(context.Background(), fmt.Sprintf("request:%s", username), target.Username)
	pipe.HSet(context.Background(), fmt.Sprintf("friend:%s", username), target.Username, target.Avatar)
	pipe.HSet(context.Background(), fmt.Sprintf("friend:%s", target.Username), username, avatar)
	_, err := pipe.Exec(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) SetRequest(username string, avatar int, target string) error {
	friend := Friend{
		UserOne:  username,
		UserTwo:  target,
		Accepted: false,
	}
	if err := db.database.Create(friend).Error; err != nil {
		return err
	}
	if err := db.cache.HSet(context.Background(), fmt.Sprintf("request:%s", target), username, avatar).Err(); err != nil {
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
				return nil, err
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
		return nil, err
	}
	return friends, nil
}

func (db *DB) GetRequests(username string) ([]User, error) {
	var friends []User
	if users, err := db.cache.HGetAll(context.Background(), fmt.Sprintf("request:%s", username)).Result(); err == nil {
		for key, friend := range users {
			avatar, err := strconv.Atoi(friend)
			if err != nil {
				return nil, err
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
		return nil, err
	}
	return friends, nil
}

func (db *DB) DeleteFriendOrRequest(username string, target string) error {
	if err := db.database.Where("user_one = ? AND user_two = ? OR user_one = ? AND user_two = ?", username, target, target, username).Delete(&Friend{}).Error; err != nil {
		fmt.Println("Error updating friends data:", err)
		return err
	}
	pipe := db.cache.Pipeline()
	pipe.HDel(context.Background(), fmt.Sprintf("request:%s", username), target)
	pipe.HDel(context.Background(), fmt.Sprintf("friend:%s", username), target)
	pipe.HDel(context.Background(), fmt.Sprintf("friend:%s", target), username)
	_, err := pipe.Exec(context.Background())
	if err != nil {
		return err
	}
	return nil
}
