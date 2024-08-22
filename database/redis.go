package database

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

func connectRedis() (*redis.Client, error) {
	cacheHost := os.Getenv("REDIS_HOST")
	cachePort := os.Getenv("REDIS_PORT")
	redisdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cacheHost, cachePort),
		Password: "",
		DB:       0,
	})
	_, err := redisdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return redisdb, nil
}
