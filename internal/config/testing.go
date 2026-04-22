package config

import (
	"sync"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ResetForTesting clears package globals so tests can run independently.
func ResetForTesting() {
	cfg = nil
	db = nil
	redisClient = nil
	cfgOnce = sync.Once{}
}

// SetConfigForTesting injects config state for tests.
func SetConfigForTesting(testCfg *Config) {
	cfg = testCfg
}

// SetDBForTesting injects database state for tests.
func SetDBForTesting(testDB *gorm.DB) {
	db = testDB
}

// SetRedisForTesting injects redis state for tests.
func SetRedisForTesting(testRedis *redis.Client) {
	redisClient = testRedis
}
