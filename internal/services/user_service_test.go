package services_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
	"keyraccoon/internal/services"
	"keyraccoon/internal/utils"
)

func TestUserServiceCreateLoginUpdateUsageAndDelete(t *testing.T) {
	service, repo := openUserService(t)

	user, err := service.CreateUser("user@example.com", "super-secret", "Test User", "user", 500, 25)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if user.Email != "user@example.com" || user.Role != "user" {
		t.Fatalf("unexpected created user: %+v", user)
	}
	if !utils.CheckPasswordHash("super-secret", user.Password) {
		t.Fatal("password was not hashed")
	}

	if _, err := service.CreateUser("user@example.com", "another-pass", "Dup", "user", 0, 0); err == nil {
		t.Fatal("CreateUser() duplicate email error = nil, want error")
	}

	loggedIn, err := service.Login("user@example.com", "super-secret")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if loggedIn.LastLogin == nil {
		t.Fatal("Login() did not update last login")
	}

	updated, err := service.UpdateUser(user.ID, map[string]any{
		"name":         "Updated User",
		"role":         "admin",
		"token_limit":  int64(800),
		"credit_limit": 55.0,
	})
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}
	if updated.Name != "Updated User" || updated.Role != "admin" {
		t.Fatalf("unexpected updated user: %+v", updated)
	}

	if err := repo.UpdateTokenUsage(user.ID, 120); err != nil {
		t.Fatalf("UpdateTokenUsage() error = %v", err)
	}
	if err := repo.UpdateCreditUsage(user.ID, 12.5); err != nil {
		t.Fatalf("UpdateCreditUsage() error = %v", err)
	}

	usage, err := service.GetUserUsage(user.ID)
	if err != nil {
		t.Fatalf("GetUserUsage() error = %v", err)
	}
	if usage["token_remaining"].(int64) != 680 {
		t.Fatalf("token_remaining = %v, want 680", usage["token_remaining"])
	}

	if err := service.SetTokenLimit(user.ID, 300); err != nil {
		t.Fatalf("SetTokenLimit() error = %v", err)
	}
	if err := service.SetCreditLimit(user.ID, 10); err != nil {
		t.Fatalf("SetCreditLimit() error = %v", err)
	}

	usage, err = service.GetUserUsage(user.ID)
	if err != nil {
		t.Fatalf("GetUserUsage() second error = %v", err)
	}
	if usage["token_used"].(int64) != 0 || usage["credit_used"].(float64) != 0 {
		t.Fatalf("usage should reset after limit update: %+v", usage)
	}

	if err := service.DisableUser(user.ID); err != nil {
		t.Fatalf("DisableUser() error = %v", err)
	}
	if _, err := service.Login("user@example.com", "super-secret"); err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("Login() after disable error = %v, want disabled error", err)
	}

	if err := service.DeleteUser(user.ID); err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
	if _, err := service.GetUser(user.ID); !errors.Is(err, repositories.ErrUserNotFound) {
		t.Fatalf("GetUser() after delete error = %v, want ErrUserNotFound", err)
	}
}

func TestUserServiceValidationAndNilRepo(t *testing.T) {
	nilService := services.NewUserService(nil)
	if _, err := nilService.CreateUser("a@b.com", "pass", "Name", "user", 0, 0); err == nil {
		t.Fatal("CreateUser() with nil repo error = nil, want error")
	}

	service, _ := openUserService(t)
	if _, err := service.CreateUser("", "", "", "user", 0, 0); err == nil {
		t.Fatal("CreateUser() with missing fields error = nil, want error")
	}
	if _, err := service.CreateUser("bad@example.com", "secret", "Bad Role", "owner", 0, 0); err == nil {
		t.Fatal("CreateUser() with invalid role error = nil, want error")
	}

	user, err := service.CreateUser("another@example.com", "secret", "Another", "user", 0, 0)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if _, err := service.UpdateUser(user.ID, map[string]any{"password": "nope"}); err == nil {
		t.Fatal("UpdateUser() with forbidden field error = nil, want error")
	}
}

func openUserService(t *testing.T) (*services.UserService, *repositories.UserRepository) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "user_service.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	repo := repositories.NewUserRepository(db)
	return services.NewUserService(repo), repo
}
