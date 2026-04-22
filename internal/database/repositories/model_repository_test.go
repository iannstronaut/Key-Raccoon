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

func TestModelRepositoryCRUD(t *testing.T) {
	db := openModelRepoDB(t)
	channel := &models.Channel{Name: "OpenAI", Type: "openai", IsActive: true}
	if err := db.Create(channel).Error; err != nil {
		t.Fatalf("Create channel error = %v", err)
	}

	repo := repositories.NewModelRepository(db)
	model := &models.Model{
		ChannelID:    channel.ID,
		Name:         "gpt-4",
		DisplayName:  "GPT-4",
		TokenPrice:   0.03,
		SystemPrompt: "Hello",
		IsActive:     true,
	}
	if err := repo.Create(model); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.GetByID(model.ID)
	if err != nil || found.Name != "gpt-4" {
		t.Fatalf("GetByID() = %+v, %v", found, err)
	}

	byName, err := repo.GetByNameAndChannelID("GPT-4", channel.ID)
	if err != nil || byName.ID != model.ID {
		t.Fatalf("GetByNameAndChannelID() = %+v, %v", byName, err)
	}

	modelsList, err := repo.GetByChannelID(channel.ID)
	if err != nil || len(modelsList) != 1 {
		t.Fatalf("GetByChannelID() len=%d err=%v", len(modelsList), err)
	}

	if err := repo.UpdateFields(model.ID, map[string]any{"display_name": "GPT-4 Turbo", "is_active": false}); err != nil {
		t.Fatalf("UpdateFields() error = %v", err)
	}
	active, err := repo.GetActiveByChannelID(channel.ID)
	if err != nil || len(active) != 0 {
		t.Fatalf("GetActiveByChannelID() len=%d err=%v", len(active), err)
	}

	if err := repo.Delete(model.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repo.GetByID(model.ID); !errors.Is(err, repositories.ErrModelNotFound) {
		t.Fatalf("GetByID() after delete error = %v, want ErrModelNotFound", err)
	}
}

func openModelRepoDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "model_repository.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.Channel{}, &models.Model{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return db
}
