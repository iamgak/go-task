package main

import (
	"fmt"
	"log"
	"os"

	"github.com/iamgak/go-task/models"
	"github.com/iamgak/go-task/pkg"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func openDBORM() (*gorm.DB, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, pkg.ErrNoEnvFileFound
	}

	dbUser := os.Getenv("DB_USERNAME")
	dbName := os.Getenv("DB_DATABASE")
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbUser == "" || dbName == "" || dbPassword == "" {
		return nil, pkg.ErrNoEnvFileFound
	}

	dsn := fmt.Sprintf("%s:%s@/%s?parseTime=true", dbUser, dbPassword, dbName)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

func InitRedis() *redis.Client {
	name := "localhost"
	passw := ""
	redis_port := 6379
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", name, redis_port),
		Password: passw,
		DB:       0,
	})

	return client
}

func MigrateDB(DB *gorm.DB) {
	err := DB.AutoMigrate(&models.Task{})
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	err = DB.AutoMigrate(&models.UsersSession{})
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	err = DB.AutoMigrate(&models.UserActivityLog{})
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
}
