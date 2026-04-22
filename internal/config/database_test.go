package config_test

import (
	"strings"
	"testing"

	"keyraccoon/internal/config"
)

func TestInitDatabaseError(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)

	cfg := &config.Config{
		DBHost:        "127.0.0.1",
		DBPort:        "1",
		DBUser:        "postgres",
		DBPassword:    "password",
		DBName:        "keyraccoon",
		DBSSLMode:     "disable",
		AdminEmail:    "admin@example.com",
		AdminPassword: "secret",
	}

	err := config.InitDatabase(cfg)
	if err == nil {
		t.Fatal("InitDatabase() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "database") {
		t.Fatalf("InitDatabase() error = %v, want database context", err)
	}
}
