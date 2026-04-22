package config_test

import (
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"

	"keyraccoon/internal/config"
)

func TestInitRedisSuccess(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	t.Cleanup(mini.Close)

	cfg := &config.Config{
		RedisHost: mini.Host(),
		RedisPort: mini.Port(),
	}

	if err := config.InitRedis(cfg); err != nil {
		t.Fatalf("InitRedis() error = %v", err)
	}

	if config.GetRedis() == nil {
		t.Fatal("GetRedis() returned nil after successful init")
	}
}

func TestInitRedisError(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)

	cfg := &config.Config{
		RedisHost: "127.0.0.1",
		RedisPort: "1",
	}

	err := config.InitRedis(cfg)
	if err == nil {
		t.Fatal("InitRedis() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "ping redis") {
		t.Fatalf("InitRedis() error = %v, want ping redis context", err)
	}
}
