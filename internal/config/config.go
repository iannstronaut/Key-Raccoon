package config

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort  string
	ServerHost  string
	Environment string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisHost string
	RedisPort string
	RedisPass string
	RedisDB   int

	JWTSecret string
	JWTExpire int

	AdminEmail    string
	AdminPassword string
}

var (
	cfg     *Config
	cfgOnce sync.Once
)

func Init() (*Config, error) {
	var initErr error

	cfgOnce.Do(func() {
		loadEnvFiles()

		cfg = &Config{
			ServerPort:    getEnv("SERVER_PORT", "3000"),
			ServerHost:    getEnv("SERVER_HOST", "0.0.0.0"),
			Environment:   getEnv("ENVIRONMENT", "development"),
			DBHost:        getEnv("DB_HOST", "localhost"),
			DBPort:        getEnv("DB_PORT", "5432"),
			DBUser:        getEnv("DB_USER", "postgres"),
			DBPassword:    getEnv("DB_PASSWORD", "password"),
			DBName:        getEnv("DB_NAME", "keyraccoon"),
			DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
			RedisHost:     getEnv("REDIS_HOST", "localhost"),
			RedisPort:     getEnv("REDIS_PORT", "6379"),
			RedisPass:     getEnv("REDIS_PASS", ""),
			RedisDB:       getEnvInt("REDIS_DB", 0),
			JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
			JWTExpire:     getEnvInt("JWT_EXPIRE", 60),
			AdminEmail:    getEnv("ADMIN_EMAIL", "admin@keyraccoon.com"),
			AdminPassword: getEnv("ADMIN_PASSWORD", "AdminPassword123"),
		}
	})

	return cfg, initErr
}

func Get() *Config {
	return cfg
}

func loadEnvFiles() {
	paths := []string{
		".env",
		".env.local",
		filepath.Join("config", ".env"),
		filepath.Join("config", ".env.local"),
	}

	for _, path := range paths {
		_ = godotenv.Overload(path)
	}
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	raw := getEnv(key, "")
	if raw == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}

	return value
}
