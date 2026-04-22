package repositories_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
)

func TestChannelAPIKeyRepositoryLifecycle(t *testing.T) {
	db := openAPIKeyRepoDB(t)
	channel := &models.Channel{Name: "OpenAI", Type: "openai", IsActive: true}
	if err := db.Create(channel).Error; err != nil {
		t.Fatalf("Create channel error = %v", err)
	}

	repo := repositories.NewChannelAPIKeyRepository(db)
	key, err := repo.GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey() error = %v", err)
	}
	if !strings.HasPrefix(key, "sk-") {
		t.Fatalf("GenerateAPIKey() = %q, want sk- prefix", key)
	}

	apiKey := &models.ChannelAPIKey{ChannelID: channel.ID, APIKey: key, IsActive: true}
	if err := repo.Create(apiKey); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.GetByKey(key)
	if err != nil || found.ID != apiKey.ID {
		t.Fatalf("GetByKey() = %+v, %v", found, err)
	}

	byChannel, err := repo.GetByChannelID(channel.ID)
	if err != nil || len(byChannel) != 1 {
		t.Fatalf("GetByChannelID() len=%d err=%v", len(byChannel), err)
	}
	active, err := repo.GetActiveByChannelID(channel.ID)
	if err != nil || len(active) != 1 {
		t.Fatalf("GetActiveByChannelID() len=%d err=%v", len(active), err)
	}

	before := apiKey.UpdatedAt
	time.Sleep(10 * time.Millisecond)
	if err := repo.UpdateLastUsed(apiKey.ID); err != nil {
		t.Fatalf("UpdateLastUsed() error = %v", err)
	}
	updated, err := repo.GetByID(apiKey.ID)
	if err != nil || !updated.UpdatedAt.After(before) {
		t.Fatalf("updated_at did not advance: before=%v after=%v err=%v", before, updated.UpdatedAt, err)
	}

	if err := repo.Disable(apiKey.ID); err != nil {
		t.Fatalf("Disable() error = %v", err)
	}
	active, err = repo.GetActiveByChannelID(channel.ID)
	if err != nil || len(active) != 0 {
		t.Fatalf("GetActiveByChannelID() after disable len=%d err=%v", len(active), err)
	}

	if err := repo.Delete(apiKey.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repo.GetByID(apiKey.ID); !errors.Is(err, repositories.ErrChannelAPIKeyNotFound) {
		t.Fatalf("GetByID() after delete error = %v, want ErrChannelAPIKeyNotFound", err)
	}
}

func openAPIKeyRepoDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "channel_apikey_repository.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.Channel{}, &models.ChannelAPIKey{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return db
}
