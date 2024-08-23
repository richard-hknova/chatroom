package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func (db *DB) GetMessages(user string) ([]Message, error) {
	var messages []Message
	result := db.database.Where("sender = ? AND received = ?", user, false).Find(&messages)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return messages, nil
}

func (db *DB) AddMsg(user string, received bool, sender string, receiver string, content string) error {
	msg := Message{
		received,
		sender,
		receiver,
		content,
	}
	if err := db.database.Create(msg).Error; err != nil {
		return err
	}
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if !received {
		db.cache.LPush(context.Background(), fmt.Sprintf("msg:%s", user), msgJSON)
	}
	return nil
}
