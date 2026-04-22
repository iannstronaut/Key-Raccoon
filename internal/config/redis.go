package config

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"keyraccoon/pkg/logger"
)

var redisClient *redis.Client

func InitRedis(cfg *Config) error {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:     cfg.RedisPass,
		DB:           cfg.RedisDB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}

	redisClient = client
	logger.Info("redis connected")
	return nil
}

func GetRedis() *redis.Client {
	return redisClient
}
