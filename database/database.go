package database

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DB struct {
	cache    *redis.Client
	database *gorm.DB
}

type User struct {
	Username string
	Avatar   int
	Hash     string
}

type Message struct {
	Received bool
	Sender   string
	Receiver string
	Content  string
}

type Friend struct {
	UserOne  string
	UserTwo  string
	Accepted bool
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

func connectMysql() (*gorm.DB, error) {
	dbUser := os.Getenv("MYSQL_USER")
	dbPass := os.Getenv("MYSQL_PASSWORD")
	dbHost := os.Getenv("MYSQL_HOST")
	dbPort := os.Getenv("MYSQL_PORT")
	dbName := os.Getenv("MYSQL_DATABASE")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPass, dbHost, dbPort, dbName)
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	mysqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}
	err = mysqlDB.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Database Connected!")
	gormDB.AutoMigrate(&User{})
	gormDB.AutoMigrate(&Friend{})
	gormDB.AutoMigrate(&Message{})
	return gormDB, nil
}
