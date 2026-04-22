package database_test

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"keyraccoon/internal/database"
	"keyraccoon/internal/models"
	"keyraccoon/internal/utils"
)

func TestSeedCreatesSuperadminOnce(t *testing.T) {
	db := openTestDB(t)

	if err := database.Seed(db, "admin@example.com", "super-secret"); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}
	if err := database.Seed(db, "admin@example.com", "super-secret"); err != nil {
		t.Fatalf("Seed() second call error = %v", err)
	}

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("user count = %d, want 1", len(users))
	}
	if users[0].Role != "superadmin" {
		t.Fatalf("role = %q, want superadmin", users[0].Role)
	}
	if !utils.CheckPasswordHash("super-secret", users[0].Password) {
		t.Fatal("stored password is not hashed as expected")
	}
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "seed.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(
		&models.User{},
		&models.Channel{},
		&models.ChannelAPIKey{},
		&models.Model{},
		&models.APIKey{},
		&models.Proxy{},
	); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return db
}
