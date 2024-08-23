package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func (db *DB) GetUnreceivedMessages(username string) ([]Message, error) {
	var messages []Message
	if data, err := db.cache.LRange(context.Background(), fmt.Sprintf("msg:%s", username), 0, -1).Result(); err == nil {
		for _, msgString := range data {
			var message Message
			if err := json.Unmarshal([]byte(msgString), &message); err != nil {
				return nil, err
			}
			messages = append(messages, message)
		}
		return messages, nil
	}
	result := db.database.Where("sender = ? AND received = ?", username, false).Find(&messages)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return messages, nil
}

func (db *DB) AddMsg(username string, msg Message) error {
	if err := db.database.Create(msg).Error; err != nil {
		return err
	}
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if !msg.Received {
		db.cache.LPush(context.Background(), fmt.Sprintf("msg:%s", username), msgJSON)
	}
	return nil
}
