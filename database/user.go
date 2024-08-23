package database

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

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
