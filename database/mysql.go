package database

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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
