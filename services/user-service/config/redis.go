package config

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

func (c *Config) RedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:        c.Redis.Host + ":" + c.Redis.Port,
		Password:    c.Redis.Password,
		DB:          c.Redis.DB,
		DialTimeout: 5 * time.Second,
		ReadTimeout: 3 * time.Second,
	})
}

// TestRedisConnection tests the Redis connection
func (c *Config) TestRedisConnection() error {
	client := c.RedisClient()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return client.Ping(ctx).Err()
}
