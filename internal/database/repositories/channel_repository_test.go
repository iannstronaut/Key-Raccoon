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

func TestChannelRepositoryCRUDAndBinding(t *testing.T) {
	db := openChannelRepoDB(t)
	userRepo := repositories.NewUserRepository(db)
	channelRepo := repositories.NewChannelRepository(db)

	user := &models.User{
		Email:    "user@example.com",
		Password: "hashed",
		Name:     "User",
		Role:     "user",
		IsActive: true,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("userRepo.Create() error = %v", err)
	}

	channel := &models.Channel{
		Name:        "OpenAI Prod",
		Type:        "openai",
		Description: "Primary channel",
		IsActive:    true,
	}
	if err := channelRepo.Create(channel); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := channelRepo.GetByID(channel.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found.Name != channel.Name {
		t.Fatalf("GetByID() name = %q, want %q", found.Name, channel.Name)
	}

	byName, err := channelRepo.GetByName("openai prod")
	if err != nil || byName.ID != channel.ID {
		t.Fatalf("GetByName() = %+v, %v", byName, err)
	}

	channels, total, err := channelRepo.GetAll(10, 0)
	if err != nil || len(channels) != 1 || total != 1 {
		t.Fatalf("GetAll() got len=%d total=%d err=%v", len(channels), total, err)
	}

	if err := channelRepo.UpdateFields(channel.ID, map[string]any{"description": "Updated"}); err != nil {
		t.Fatalf("UpdateFields() error = %v", err)
	}
	found, _ = channelRepo.GetByID(channel.ID)
	if found.Description != "Updated" {
		t.Fatalf("description = %q, want Updated", found.Description)
	}

	active, totalActive, err := channelRepo.GetActive(10, 0)
	if err != nil || len(active) != 1 || totalActive != 1 {
		t.Fatalf("GetActive() got len=%d total=%d err=%v", len(active), totalActive, err)
	}

	if err := channelRepo.BindUserToChannel(user.ID, channel.ID); err != nil {
		t.Fatalf("BindUserToChannel() error = %v", err)
	}
	userChannels, err := channelRepo.GetByUserID(user.ID)
	if err != nil || len(userChannels) != 1 {
		t.Fatalf("GetByUserID() len=%d err=%v", len(userChannels), err)
	}

	if err := channelRepo.UnbindUserFromChannel(user.ID, channel.ID); err != nil {
		t.Fatalf("UnbindUserFromChannel() error = %v", err)
	}
	userChannels, err = channelRepo.GetByUserID(user.ID)
	if err != nil || len(userChannels) != 0 {
		t.Fatalf("GetByUserID() after unbind len=%d err=%v", len(userChannels), err)
	}

	if err := channelRepo.Delete(channel.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := channelRepo.GetByID(channel.ID); !errors.Is(err, repositories.ErrChannelNotFound) {
		t.Fatalf("GetByID() after delete error = %v, want ErrChannelNotFound", err)
	}
}

func openChannelRepoDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "channel_repository.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.User{}, &models.Channel{}, &models.ChannelAPIKey{}, &models.Model{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return db
}
