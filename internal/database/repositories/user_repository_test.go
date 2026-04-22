package repositories_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
)

func TestUserRepositoryCRUDAndUsage(t *testing.T) {
	repo := openUserRepository(t)

	user := &models.User{
		Email:       "repo@example.com",
		Password:    "hashed",
		Name:        "Repo User",
		Role:        "user",
		IsActive:    true,
		TokenLimit:  100,
		CreditLimit: 50,
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	foundByID, err := repo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if foundByID.Email != user.Email {
		t.Fatalf("GetByID() email = %q, want %q", foundByID.Email, user.Email)
	}

	foundByEmail, err := repo.GetByEmail("REPO@example.com")
	if err != nil {
		t.Fatalf("GetByEmail() error = %v", err)
	}
	if foundByEmail.ID != user.ID {
		t.Fatalf("GetByEmail() id = %d, want %d", foundByEmail.ID, user.ID)
	}

	users, total, err := repo.GetAll(10, 0)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(users) != 1 || total != 1 {
		t.Fatalf("GetAll() got len=%d total=%d, want 1 and 1", len(users), total)
	}

	if err := repo.UpdateFields(user.ID, map[string]any{"name": "Updated Repo"}); err != nil {
		t.Fatalf("UpdateFields() error = %v", err)
	}
	foundByID, _ = repo.GetByID(user.ID)
	if foundByID.Name != "Updated Repo" {
		t.Fatalf("name = %q, want Updated Repo", foundByID.Name)
	}

	if err := repo.UpdateTokenUsage(user.ID, 10); err != nil {
		t.Fatalf("UpdateTokenUsage() error = %v", err)
	}
	if err := repo.UpdateCreditUsage(user.ID, 12.5); err != nil {
		t.Fatalf("UpdateCreditUsage() error = %v", err)
	}

	ok, err := repo.HasTokenAvailable(user.ID, 91)
	if err != nil || ok {
		t.Fatalf("HasTokenAvailable() = %v, %v, want false, nil", ok, err)
	}
	ok, err = repo.HasCreditAvailable(user.ID, 38)
	if err != nil || ok {
		t.Fatalf("HasCreditAvailable() = %v, %v, want false, nil", ok, err)
	}

	if err := repo.Delete(user.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repo.GetByID(user.ID); !errors.Is(err, repositories.ErrUserNotFound) {
		t.Fatalf("GetByID() after delete error = %v, want ErrUserNotFound", err)
	}
}

func openUserRepository(t *testing.T) *repositories.UserRepository {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "user_repository.db")
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
	return repositories.NewUserRepository(db)
}
