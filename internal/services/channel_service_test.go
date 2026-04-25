package services_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
	"keyraccoon/internal/services"
)

func TestChannelServiceLifecycle(t *testing.T) {
	service, userRepo, channelRepo, apiKeyRepo, modelRepo := openChannelService(t)

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

	channel, err := service.CreateChannel("OpenAI Prod", "openai", "Primary", "", 0, "price")
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	if _, err := service.CreateChannel("OpenAI Prod", "openai", "Duplicate", "", 0, "price"); err == nil {
		t.Fatal("CreateChannel() duplicate error = nil, want error")
	}

	updated, err := service.UpdateChannel(channel.ID, map[string]any{"description": "Updated channel"})
	if err != nil || updated.Description != "Updated channel" {
		t.Fatalf("UpdateChannel() = %+v, %v", updated, err)
	}

	apiKey, err := service.AddAPIKey(channel.ID)
	if err != nil {
		t.Fatalf("AddAPIKey() error = %v", err)
	}
	if apiKey.ChannelID != channel.ID {
		t.Fatalf("apiKey.ChannelID = %d, want %d", apiKey.ChannelID, channel.ID)
	}

	keys, err := service.GetChannelAPIKeys(channel.ID)
	if err != nil || len(keys) != 1 {
		t.Fatalf("GetChannelAPIKeys() len=%d err=%v", len(keys), err)
	}

	rotated, err := service.RotateAPIKey(channel.ID)
	if err != nil {
		t.Fatalf("RotateAPIKey() error = %v", err)
	}
	activeKeys, err := apiKeyRepo.GetActiveByChannelID(channel.ID)
	if err != nil || len(activeKeys) != 1 || activeKeys[0].ID != rotated.ID {
		t.Fatalf("active keys after rotate = %+v err=%v", activeKeys, err)
	}

	model, err := service.AddModel(channel.ID, "gpt-4", "GPT-4", 0.03, "You are helpful")
	if err != nil {
		t.Fatalf("AddModel() error = %v", err)
	}
	if _, err := service.AddModel(channel.ID, "gpt-4", "Dup", 0.03, "dup"); err == nil {
		t.Fatal("AddModel() duplicate error = nil, want error")
	}

	modelsList, err := service.GetChannelModels(channel.ID)
	if err != nil || len(modelsList) != 1 {
		t.Fatalf("GetChannelModels() len=%d err=%v", len(modelsList), err)
	}

	updatedModel, err := service.UpdateModel(model.ID, map[string]any{"display_name": "GPT-4 Turbo"})
	if err != nil || updatedModel.DisplayName != "GPT-4 Turbo" {
		t.Fatalf("UpdateModel() = %+v, %v", updatedModel, err)
	}

	if err := service.BindUserToChannel(user.ID, channel.ID); err != nil {
		t.Fatalf("BindUserToChannel() error = %v", err)
	}
	userChannels, err := service.GetUserChannels(user.ID)
	if err != nil || len(userChannels) != 1 {
		t.Fatalf("GetUserChannels() len=%d err=%v", len(userChannels), err)
	}

	if err := service.UnbindUserFromChannel(user.ID, channel.ID); err != nil {
		t.Fatalf("UnbindUserFromChannel() error = %v", err)
	}

	if err := service.RemoveAPIKey(rotated.ID); err != nil {
		t.Fatalf("RemoveAPIKey() error = %v", err)
	}
	if _, err := apiKeyRepo.GetByID(rotated.ID); !errors.Is(err, repositories.ErrChannelAPIKeyNotFound) {
		t.Fatalf("GetByID(api key) after delete error = %v", err)
	}

	if err := service.DeleteModel(model.ID); err != nil {
		t.Fatalf("DeleteModel() error = %v", err)
	}
	if _, err := modelRepo.GetByID(model.ID); !errors.Is(err, repositories.ErrModelNotFound) {
		t.Fatalf("GetByID(model) after delete error = %v", err)
	}

	if err := service.DeleteChannel(channel.ID); err != nil {
		t.Fatalf("DeleteChannel() error = %v", err)
	}
	if _, err := channelRepo.GetByID(channel.ID); !errors.Is(err, repositories.ErrChannelNotFound) {
		t.Fatalf("GetByID(channel) after delete error = %v", err)
	}
}

func openChannelService(t *testing.T) (*services.ChannelService, *repositories.UserRepository, *repositories.ChannelRepository, *repositories.ChannelAPIKeyRepository, *repositories.ModelRepository) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "channel_service.db")
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

	userRepo := repositories.NewUserRepository(db)
	channelRepo := repositories.NewChannelRepository(db)
	apiKeyRepo := repositories.NewChannelAPIKeyRepository(db)
	modelRepo := repositories.NewModelRepository(db)

	return services.NewChannelService(channelRepo, apiKeyRepo, modelRepo, userRepo), userRepo, channelRepo, apiKeyRepo, modelRepo
}
