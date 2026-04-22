package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"keyraccoon/internal/config"
	"keyraccoon/internal/models"
)

func TestRootCommandIncludesPhase2Commands(t *testing.T) {
	root := NewRootCommand()
	names := make([]string, 0, len(root.Commands()))
	for _, command := range root.Commands() {
		names = append(names, command.Name())
	}

	assertContains(t, names, "config")
	assertContains(t, names, "setup")
	assertContains(t, names, "create-user")
}

func TestConfigCommandPrintsSummary(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)

	tempDir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	t.Setenv("ENVIRONMENT", "test")
	t.Setenv("SERVER_HOST", "127.0.0.1")
	t.Setenv("SERVER_PORT", "4010")
	t.Setenv("DB_HOST", "db.test")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "keyraccoon_test")
	t.Setenv("REDIS_HOST", "redis.test")
	t.Setenv("REDIS_PORT", "6379")

	root := NewRootCommand()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"config"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := normalizeLine(buf.String())
	if !strings.Contains(output, "environment=test") {
		t.Fatalf("output = %q, want environment summary", output)
	}
}

func TestSetupAndCreateUserCommands(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	originalInitConfig := cliInitConfig
	originalInitDatabase := cliInitDatabase
	originalGetDB := cliGetDB
	t.Cleanup(func() {
		cliInitConfig = originalInitConfig
		cliInitDatabase = originalInitDatabase
		cliGetDB = originalGetDB
	})

	testCfg := &config.Config{
		AdminEmail:    "admin@example.com",
		AdminPassword: "AdminPassword123",
		JWTSecret:     "test-secret",
		JWTExpire:     60,
	}
	cliInitConfig = func() (*config.Config, error) { return testCfg, nil }
	cliInitDatabase = func(cfg *config.Config) error { return nil }
	cliGetDB = func() *gorm.DB { return db }

	root := NewRootCommand()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"setup"})
	if err := root.Execute(); err != nil {
		t.Fatalf("setup Execute() error = %v", err)
	}
	if !strings.Contains(normalizeLine(buf.String()), "Superadmin created successfully") {
		t.Fatalf("setup output = %q", buf.String())
	}

	root = NewRootCommand()
	buf = new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{
		"create-user",
		"--email", "user@example.com",
		"--password", "SecurePass123",
		"--name", "John Doe",
		"--role", "user",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("create-user Execute() error = %v", err)
	}
	if !strings.Contains(normalizeLine(buf.String()), "User created successfully") {
		t.Fatalf("create-user output = %q", buf.String())
	}
}

func assertContains(t *testing.T, items []string, want string) {
	t.Helper()
	for _, item := range items {
		if item == want {
			return
		}
	}
	t.Fatalf("items %v do not contain %q", items, want)
}
