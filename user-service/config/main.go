package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	host := os.Getenv("PG_HOST")
	if host == "" {
		host = "localhost"
	}
	user := os.Getenv("PG_USER")
	if user == "" {
		user = "postgres"
	}
	pass := os.Getenv("PG_PASSWORD")
	if pass == "" {
		pass = "postgres"
	}
	dbname := os.Getenv("PG_DATABASE")
	if dbname == "" {
		dbname = "internhub"
	}
	port := os.Getenv("PG_PORT")
	if port == "" {
		port = "5432"
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, pass, dbname, port)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(25)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}
	log.Println("user-service: DB connected (PostgreSQL)")
}
