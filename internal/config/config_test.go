package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"keyraccoon/internal/config"
)

func TestInitLoadsEnvironmentValues(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)

	tempDir := t.TempDir()
	writeEnvFile(t, filepath.Join(tempDir, "config", ".env"), "SERVER_PORT=9999\n")

	t.Setenv("SERVER_HOST", "127.0.0.1")
	t.Setenv("ENVIRONMENT", "test")
	t.Setenv("DB_HOST", "db.local")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "tester")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "keyraccoon_test")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("REDIS_HOST", "redis.local")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("REDIS_PASS", "redis-secret")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("JWT_EXPIRE", "90")
	t.Setenv("ADMIN_EMAIL", "admin@test.local")
	t.Setenv("ADMIN_PASSWORD", "admin-pass")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	cfg, err := config.Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if cfg.ServerPort != "9999" {
		t.Fatalf("ServerPort = %q, want %q", cfg.ServerPort, "9999")
	}
	if cfg.ServerHost != "127.0.0.1" || cfg.Environment != "test" {
		t.Fatalf("unexpected server config: %+v", cfg)
	}
	if cfg.RedisDB != 2 || cfg.JWTExpire != 90 {
		t.Fatalf("unexpected int config: redis=%d jwt=%d", cfg.RedisDB, cfg.JWTExpire)
	}
	if config.Get() != cfg {
		t.Fatal("Get() did not return initialized config")
	}
}

func TestInitFallsBackToDefaultsForInvalidInts(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)

	tempDir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	t.Setenv("REDIS_DB", "invalid")
	t.Setenv("JWT_EXPIRE", "not-a-number")

	cfg, err := config.Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if cfg.RedisDB != 0 {
		t.Fatalf("RedisDB = %d, want 0", cfg.RedisDB)
	}
	if cfg.JWTExpire != 60 {
		t.Fatalf("JWTExpire = %d, want 60", cfg.JWTExpire)
	}
}

func writeEnvFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writefile: %v", err)
	}
}
