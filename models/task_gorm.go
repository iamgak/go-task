package models

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/iamgak/go-task/pkg"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Task struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint       `gorm:"index;not null" json:"-" binding:"-"`
	Title       string     `gorm:"not null" json:"title,omitempty"` // Optional
	Description string     `gorm:"not null" json:"description"`
	Status      string     `gorm:"type:enum('pending','in progress','completed');not null" json:"status"`
	IsDeleted   bool       `gorm:"default:0" json:"-"`                   // Hidden from JSON (soft delete)
	DueAt       *time.Time `gorm:"default:null" json:"due_at,omitempty"` // Optional
	Version     uint       `gorm:"default:1" json:"version"`
	CreatedAt   *time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP()" json:"created_at,omitempty" binding:"-"`
	UpdatedAt   *time.Time `gorm:"default:null" json:"updated_at,omitempty" binding:"-"` // Optional
	DeletedAt   *time.Time `gorm:"default:null" json:"-" binding:"-"`                    // Hidden from JSON (soft delete)
}

type TaskModelORM struct {
	db     *gorm.DB
	logger *logrus.Logger
	mute   sync.RWMutex
	redis  RedisStruct
}

func (c *TaskModelORM) TaskById(ctx context.Context, taskID int) (*Task, error) {
	var task *Task
	cacheKey := fmt.Sprintf("tasks:id:%d", taskID)
	cachedData, err := c.redis.getRedis(ctx, cacheKey)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if err != redis.Nil {
		byteVal, ok := cachedData.([]byte)
		if !ok {
			strVal, ok := cachedData.(string)
			if !ok {
				return nil, fmt.Errorf("expected %T to be a []byte or string", byteVal)
			}
			byteVal = []byte(strVal)
		}

		err = json.Unmarshal(byteVal, &task)
		return task, err
	}

	result := c.db.WithContext(ctx).Where("id = ? AND is_deleted = 0", taskID).First(&task)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return task, pkg.ErrNoRecord
		}

		c.logger.Error("Query Execution Failed: ", result.Error)
		return task, result.Error
	}

	tasksJSON, _ := json.Marshal(task)
	return task, c.redis.setRedis(ctx, cacheKey, tasksJSON, 10*time.Minute)
}

func (c *TaskModelORM) TaskListing(ctx context.Context, f *Filters) ([]*Task, error) {
	var task []*Task
	cacheKey := fmt.Sprintf("tasks:listing:%s", strings.TrimSpace(f.otherConditions()))
	cachedData, err := c.redis.getRedis(ctx, cacheKey)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if err != redis.Nil {
		byteVal, ok := cachedData.([]byte)
		if !ok {
			strVal, ok := cachedData.(string)
			if !ok {
				return nil, fmt.Errorf("expected %T to be a []byte or string", byteVal)
			}
			byteVal = []byte(strVal)
		}

		err = json.Unmarshal(byteVal, &task)
		return task, err
	}

	result := c.db.WithContext(ctx).Where("is_deleted = 0 " + f.otherConditions()).Find(&task)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return task, pkg.ErrNoRecord
		}
		return task, result.Error
	}

	tasksJSON, _ := json.Marshal(task)
	return task, c.redis.setRedis(ctx, cacheKey, tasksJSON, 10*time.Minute)
}

func (c *TaskModelORM) CreateTask(ctx context.Context, task *Task) error {
	c.mute.Lock()
	defer c.mute.Unlock()
	result := c.db.WithContext(ctx).Model(&Task{}).Create(task)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return pkg.ErrNoRecord
	}

	return c.redis.FlushCache(ctx)
}
func (c *TaskModelORM) UpdateTask(ctx context.Context, id int, task *Task) error {
	c.mute.Lock()
	defer c.mute.Unlock()

	updates := map[string]interface{}{
		"title":       task.Title,
		"description": task.Description,
		"status":      task.Status,
		"updated_at":  time.Now(),
		"version":     gorm.Expr("version + 1"),
	}

	if !task.DueAt.IsZero() {
		updates["due_at"] = task.DueAt
	}

	// Perform the update with conditional check
	var tasks Task
	result := c.db.
		WithContext(ctx).
		Model(tasks).
		// Clauses(clause.Returning{Columns: []clause.Column{{Name: "title"}, {Name: "description"}}}).
		Where("id = ? AND user_id = ? AND is_deleted = 0", id, task.UserID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return pkg.ErrInvalidUserFound
	}

	_ = c.redis.deleteRedis(ctx, fmt.Sprintf("task:id:%d", id))
	return c.redis.FlushCache(ctx)
}

func (c *TaskModelORM) SoftDelete(ctx context.Context, userID, taskID uint) error {
	c.mute.Lock()
	defer c.mute.Unlock()
	updates := map[string]interface{}{
		"deleted_at": time.Now(),
		"is_deleted": true,
	}
	result := c.db.WithContext(ctx).
		Model(&Task{}).
		Where("id = ? AND user_id = ? AND is_deleted = 0", taskID, userID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return pkg.ErrInvalidUserFound
	}

	_ = c.redis.deleteRedis(ctx, fmt.Sprintf("task:id:%d", taskID))
	return c.redis.FlushCache(ctx)
}

func (m *TaskModelORM) ValidateTaskData(task *Task, updated bool) *pkg.Validator {
	validator := &pkg.Validator{
		Errors: make(map[string]string),
	}

	validator.CheckField(validator.NotBlank(task.Title), "title", "Please, fill the title field")
	validator.CheckField(validator.NotBlank(task.Description), "description", "Please, fill the description field")
	validator.CheckField(validator.NotBlank(task.Status), "status", "Please, fill the status field")
	if validator.Errors["status"] == "" {
		validator.CheckField(validator.ValidStatus(task.Status), "status", "Invalid Status Input")
	}

	if task.DueAt != nil && !task.DueAt.IsZero() {
		parsedDate, err := time.Parse("2006-01-02", task.DueAt.Format("2006-01-02"))
		if err != nil {
			validator.CheckField(false, "due_at", "Incorrect format. Expected format: YYYY-MM-DD")
		} else if !updated && parsedDate.Before(time.Now()) {
			validator.CheckField(false, "due_at", "Due date must be a future date")
		}
	}

	return validator
}
