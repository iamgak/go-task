package models

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type RedisStruct struct {
	client *redis.Client
	logger *logrus.Logger
}

func (c *RedisStruct) getRedis(ctx context.Context, key string) (interface{}, error) {
	val, err := c.client.Get(ctx, key).Result()
	return val, err
}

func (c *RedisStruct) setRedis(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := c.client.Set(ctx, key, value, expiration).Err()
	return err
}

func (c *RedisStruct) deleteRedis(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	return err
}
func (c *RedisStruct) Publish(ctx context.Context, msg []byte) error {
	// msg := []byte("New to-do item added")
	err := c.client.Publish(ctx, "todo.notifications", msg).Err()
	return err
}

func (c *RedisStruct) Subscribe(ctx context.Context) {
	// msg := []byte("New to-do item added")
	sub := c.client.Subscribe(ctx, "todo.notifications")
	ch := sub.Channel()
	for msg := range ch {
		c.logger.Info("Received message:", msg)
	}
}

func (m *RedisStruct) FlushCache(ctx context.Context) error {
	recordType := "tasks:listing:"
	pattern := fmt.Sprintf("%s:*", recordType)
	keys, err := m.client.Keys(ctx, pattern).Result()
	if err != nil {
		m.logger.Errorf("Error fetching keys:%T", err)
		return err
	}

	if len(keys) > 0 {
		_, err := m.client.Del(ctx, keys...).Result()
		if err != nil {
			m.logger.Errorf("Error deleting cache keys:%T", err)
			return err
		}
	}

	return err
}
