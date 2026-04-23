package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
)

type UserAPIKeyService struct {
	apiKeyRepo  *repositories.UserAPIKeyRepository
	userRepo    *repositories.UserRepository
	channelRepo *repositories.ChannelRepository
	modelRepo   *repositories.ModelRepository
}

func NewUserAPIKeyService(
	apiKeyRepo *repositories.UserAPIKeyRepository,
	userRepo *repositories.UserRepository,
	channelRepo *repositories.ChannelRepository,
	modelRepo *repositories.ModelRepository,
) *UserAPIKeyService {
	return &UserAPIKeyService{
		apiKeyRepo:  apiKeyRepo,
		userRepo:    userRepo,
		channelRepo: channelRepo,
		modelRepo:   modelRepo,
	}
}

// GenerateAPIKey generates a secure random API key
func (s *UserAPIKeyService) GenerateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "kr_" + base64.URLEncoding.EncodeToString(b)[:43], nil
}

// CreateAPIKey creates a new user API key
func (s *UserAPIKeyService) CreateAPIKey(
	userID uint,
	name string,
	usageLimit int64,
	expiresAt *time.Time,
	channelIDs []uint,
	modelIDs []uint,
) (*models.UserAPIKey, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("api key name is required")
	}

	// Verify user exists
	if _, err := s.userRepo.GetByID(userID); err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify channels exist
	for _, channelID := range channelIDs {
		if _, err := s.channelRepo.GetByID(channelID); err != nil {
			return nil, fmt.Errorf("channel %d not found", channelID)
		}
	}

	// Verify models exist
	for _, modelID := range modelIDs {
		if _, err := s.modelRepo.GetByID(modelID); err != nil {
			return nil, fmt.Errorf("model %d not found", modelID)
		}
	}

	// Generate API key
	key, err := s.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate api key: %w", err)
	}

	apiKey := &models.UserAPIKey{
		UserID:     userID,
		Name:       name,
		Key:        key,
		IsActive:   true,
		UsageLimit: usageLimit,
		UsageCount: 0,
		ExpiresAt:  expiresAt,
	}

	if err := s.apiKeyRepo.Create(apiKey); err != nil {
		return nil, fmt.Errorf("failed to create api key: %w", err)
	}

	// Add channels
	for _, channelID := range channelIDs {
		if err := s.apiKeyRepo.AddChannel(apiKey.ID, channelID); err != nil {
			return nil, fmt.Errorf("failed to add channel: %w", err)
		}
	}

	// Add models
	for _, modelID := range modelIDs {
		if err := s.apiKeyRepo.AddModel(apiKey.ID, modelID); err != nil {
			return nil, fmt.Errorf("failed to add model: %w", err)
		}
	}

	// Reload with relations
	return s.apiKeyRepo.GetByID(apiKey.ID)
}

// GetAPIKey gets an API key by ID
func (s *UserAPIKeyService) GetAPIKey(id uint) (*models.UserAPIKey, error) {
	return s.apiKeyRepo.GetByID(id)
}

// GetAPIKeyByKey gets an API key by key string
func (s *UserAPIKeyService) GetAPIKeyByKey(key string) (*models.UserAPIKey, error) {
	return s.apiKeyRepo.GetByKey(key)
}

// GetUserAPIKeys gets all API keys for a user
func (s *UserAPIKeyService) GetUserAPIKeys(userID uint) ([]models.UserAPIKey, error) {
	return s.apiKeyRepo.GetByUserID(userID)
}

// GetAllAPIKeys gets all API keys with pagination
func (s *UserAPIKeyService) GetAllAPIKeys(limit, offset int) ([]models.UserAPIKey, int64, error) {
	return s.apiKeyRepo.GetAll(limit, offset)
}

// UpdateAPIKey updates an API key
func (s *UserAPIKeyService) UpdateAPIKey(id uint, updates map[string]interface{}) (*models.UserAPIKey, error) {
	apiKey, err := s.apiKeyRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	allowedFields := map[string]bool{
		"name":         true,
		"is_active":    true,
		"usage_limit":  true,
		"expires_at":   true,
	}

	for key := range updates {
		if !allowedFields[key] {
			return nil, fmt.Errorf("field %s cannot be updated", key)
		}
	}

	if name, ok := updates["name"].(string); ok {
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, errors.New("api key name cannot be empty")
		}
		apiKey.Name = name
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		apiKey.IsActive = isActive
	}

	if usageLimit, ok := updates["usage_limit"].(int64); ok {
		if usageLimit < 0 {
			return nil, errors.New("usage limit cannot be negative")
		}
		apiKey.UsageLimit = usageLimit
	}

	if expiresAt, ok := updates["expires_at"].(*time.Time); ok {
		apiKey.ExpiresAt = expiresAt
	}

	if err := s.apiKeyRepo.Update(apiKey); err != nil {
		return nil, err
	}

	return s.apiKeyRepo.GetByID(id)
}

// DeleteAPIKey deletes an API key
func (s *UserAPIKeyService) DeleteAPIKey(id uint) error {
	if _, err := s.apiKeyRepo.GetByID(id); err != nil {
		return err
	}
	return s.apiKeyRepo.Delete(id)
}

// AddChannel adds a channel to an API key
func (s *UserAPIKeyService) AddChannel(apiKeyID, channelID uint) error {
	if _, err := s.apiKeyRepo.GetByID(apiKeyID); err != nil {
		return err
	}
	if _, err := s.channelRepo.GetByID(channelID); err != nil {
		return err
	}
	return s.apiKeyRepo.AddChannel(apiKeyID, channelID)
}

// RemoveChannel removes a channel from an API key
func (s *UserAPIKeyService) RemoveChannel(apiKeyID, channelID uint) error {
	return s.apiKeyRepo.RemoveChannel(apiKeyID, channelID)
}

// AddModel adds a model to an API key
func (s *UserAPIKeyService) AddModel(apiKeyID, modelID uint) error {
	if _, err := s.apiKeyRepo.GetByID(apiKeyID); err != nil {
		return err
	}
	if _, err := s.modelRepo.GetByID(modelID); err != nil {
		return err
	}
	return s.apiKeyRepo.AddModel(apiKeyID, modelID)
}

// RemoveModel removes a model from an API key
func (s *UserAPIKeyService) RemoveModel(apiKeyID, modelID uint) error {
	return s.apiKeyRepo.RemoveModel(apiKeyID, modelID)
}

// ValidateAPIKey validates an API key for usage
func (s *UserAPIKeyService) ValidateAPIKey(key string) (*models.UserAPIKey, error) {
	apiKey, err := s.apiKeyRepo.GetActiveByKey(key)
	if err != nil {
		return nil, err
	}

	if !apiKey.CanUse() {
		if apiKey.IsExpired() {
			return nil, errors.New("api key has expired")
		}
		if apiKey.IsLimitReached() {
			return nil, errors.New("api key usage limit reached")
		}
		return nil, errors.New("api key is not active")
	}

	return apiKey, nil
}

// IncrementUsage increments the usage count for an API key
func (s *UserAPIKeyService) IncrementUsage(id uint) error {
	return s.apiKeyRepo.IncrementUsage(id)
}
