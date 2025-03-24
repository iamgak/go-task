package models

import (
	"github.com/gin-gonic/gin"
	"github.com/iamgak/go-task/pkg"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Init struct {
	// Task TaskModel
	// Users        UserModel
	UsersORM     UserModelORM
	Redis        RedisStruct
	TaskModelORM TaskModelORM
}

func Constructor(dbORM *gorm.DB, redis *redis.Client, Logger *logrus.Logger) *Init {
	RedisClient := RedisStruct{client: redis, logger: Logger}
	return &Init{
		// Task: TaskModel{db: db, redis: RedisClient, logger: Logger},
		// Users:        UserModel{db: db, redis: redis, logger: Logger},
		UsersORM:     UserModelORM{db: dbORM, redis: redis, logger: Logger},
		TaskModelORM: TaskModelORM{db: dbORM, redis: RedisClient, logger: Logger},
		// Review: ReviewModel{db: db, redis: rd},
	}
}

func NewFilters(c *gin.Context) *Filters {
	var validator *pkg.Validator
	return &Filters{
		PageSize:  validator.ReadInt(c.Query("limit"), 10),
		CurrPage:  validator.ReadInt(c.Query("page"), 1),
		Status:    validator.ReadString(c.Query("status"), ""),
		SortOrder: validator.ReadString(c.Query("sort_order"), "desc"),
		SortBy:    validator.ReadString(c.Query("sort_by"), "id"),
		DueAfter:  validator.GetValidDate(c.Query("due_date_after")),
		DueBefore: validator.GetValidDate(c.Query("due_date_before")),
	}
}
