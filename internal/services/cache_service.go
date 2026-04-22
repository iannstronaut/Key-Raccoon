package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"keyraccoon/internal/config"
)

type CacheService struct {
	redis *redis.Client
}

func NewCacheService() *CacheService {
	return &CacheService{
		redis: config.GetRedis(),
	}
}

const (
	CacheChannelList = "cache:channels:list"
	CacheModelList   = "cache:models:%d"
	CacheUserData    = "cache:user:%d"
	CacheTTL         = 5 * time.Minute
)

func (c *CacheService) Get(ctx context.Context, key string) ([]byte, error) {
	if c.redis == nil {
		return nil, nil
	}
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

func (c *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c.redis == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache value: %w", err)
	}
	return c.redis.Set(ctx, key, data, ttl).Err()
}

func (c *CacheService) Delete(ctx context.Context, key string) error {
	if c.redis == nil {
		return nil
	}
	return c.redis.Del(ctx, key).Err()
}

func (c *CacheService) InvalidateChannelCache(ctx context.Context) {
	c.Delete(ctx, CacheChannelList)
}

func (c *CacheService) InvalidateModelCache(ctx context.Context, channelID uint) {
	c.Delete(ctx, fmt.Sprintf(CacheModelList, channelID))
}

func (c *CacheService) InvalidateUserCache(ctx context.Context, userID uint) {
	c.Delete(ctx, fmt.Sprintf(CacheUserData, userID))
}
