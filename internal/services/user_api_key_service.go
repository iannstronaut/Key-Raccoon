package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"keyraccoon/internal/config"
	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
)

type UserAPIKeyService struct {
	apiKeyRepo  *repositories.UserAPIKeyRepository
	userRepo    *repositories.UserRepository
	channelRepo *repositories.ChannelRepository
	modelRepo   *repositories.ModelRepository
	redis       *redis.Client
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
		redis:       config.GetRedis(),
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

// VerifyAPIKey verifies that an API key exists, is active, not expired, and within usage limit
func (s *UserAPIKeyService) VerifyAPIKey(key string) (*models.UserAPIKey, error) {
	apiKey, err := s.apiKeyRepo.GetByKey(key)
	if err != nil {
		return nil, errors.New("invalid api key")
	}

	if !apiKey.CanUse() {
		if !apiKey.IsActive {
			return nil, errors.New("api key is disabled")
		}
		if apiKey.IsExpired() {
			return nil, errors.New("api key has expired")
		}
		if apiKey.IsLimitReached() {
			return nil, errors.New("api key usage limit reached")
		}
		return nil, errors.New("api key cannot be used")
	}

	return apiKey, nil
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

// CreateSelfServiceAPIKey creates an API key for the authenticated user (self-service)
func (s *UserAPIKeyService) CreateSelfServiceAPIKey(
	userID uint,
	name string,
	channelIDs []uint,
	modelIDs []uint,
	tokenLimit int64,
	expiresAt *time.Time,
) (*models.UserAPIKey, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("api key name is required")
	}

	// Verify user exists
	if _, err := s.userRepo.GetByID(userID); err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if len(channelIDs) == 0 {
		return nil, errors.New("at least one channel is required")
	}

	// Verify all channels are assigned to this user
	userChannels, err := s.channelRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user channels: %w", err)
	}

	userChannelMap := make(map[uint]bool)
	for _, ch := range userChannels {
		userChannelMap[ch.ID] = true
	}

	for _, chID := range channelIDs {
		if !userChannelMap[chID] {
			return nil, fmt.Errorf("channel %d is not assigned to you", chID)
		}
	}

	// Verify models belong to selected channels
	if len(modelIDs) > 0 {
		for _, modelID := range modelIDs {
			model, err := s.modelRepo.GetByID(modelID)
			if err != nil {
				return nil, fmt.Errorf("model %d not found", modelID)
			}
			found := false
			for _, chID := range channelIDs {
				if model.ChannelID == chID {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("model %d does not belong to any selected channel", modelID)
			}
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
		TokenLimit: tokenLimit,
		UsageLimit: 0,
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

// DeleteSelfAPIKey deletes an API key owned by the authenticated user
func (s *UserAPIKeyService) DeleteSelfAPIKey(keyID uint, userID uint) error {
	apiKey, err := s.apiKeyRepo.GetByID(keyID)
	if err != nil {
		return err
	}
	if apiKey.UserID != userID {
		return errors.New("you can only delete your own api keys")
	}
	return s.apiKeyRepo.Delete(keyID)
}

// ============ TOKEN/CREDIT TRACKING (Compatible with legacy APIKey) ============

// HasTokenAvailable checks token availability
func (s *UserAPIKeyService) HasTokenAvailable(keyID uint, requiredToken int64) (bool, error) {
	return s.apiKeyRepo.HasTokenAvailable(keyID, requiredToken)
}

// HasCreditAvailable checks credit availability
func (s *UserAPIKeyService) HasCreditAvailable(keyID uint, requiredCredit float64) (bool, error) {
	return s.apiKeyRepo.HasCreditAvailable(keyID, requiredCredit)
}

// RecordTokenUsage records token usage in both Redis (real-time cache) and database (persistent)
func (s *UserAPIKeyService) RecordTokenUsage(keyID uint, tokens int64) error {
	// Always update the database for persistent tracking
	dbErr := s.apiKeyRepo.UpdateTokenUsage(keyID, tokens)

	// Also update Redis for real-time reads (optional, best-effort)
	ctx := context.Background()
	if s.redis != nil {
		key := fmt.Sprintf("user_apikey:%d:tokens", keyID)
		if err := s.redis.IncrBy(ctx, key, tokens).Err(); err == nil {
			s.redis.Expire(ctx, key, 24*time.Hour)
		}
	}

	return dbErr
}

// RecordCreditUsage records credit usage in both Redis (real-time cache) and database (persistent)
func (s *UserAPIKeyService) RecordCreditUsage(keyID uint, credit float64) error {
	// Always update the database for persistent tracking
	dbErr := s.apiKeyRepo.UpdateCreditUsage(keyID, credit)

	// Also update Redis for real-time reads (optional, best-effort)
	ctx := context.Background()
	if s.redis != nil {
		key := fmt.Sprintf("user_apikey:%d:credit", keyID)
		if err := s.redis.IncrByFloat(ctx, key, credit).Err(); err == nil {
			s.redis.Expire(ctx, key, 24*time.Hour)
		}
	}

	return dbErr
}

// GetRealtimeUsage returns usage stats from the database
func (s *UserAPIKeyService) GetRealtimeUsage(keyID uint) (map[string]any, error) {
	apiKey, err := s.apiKeyRepo.GetByID(keyID)
	if err != nil {
		return nil, err
	}

	tokenUsed := apiKey.TokenUsed
	creditUsed := apiKey.CreditUsed

	tokenRemaining := int64(-1)
	if apiKey.TokenLimit != -1 && apiKey.TokenLimit != 0 {
		tokenRemaining = apiKey.TokenLimit - tokenUsed
	}

	creditRemaining := float64(-1)
	if apiKey.CreditLimit != -1 && apiKey.CreditLimit != 0 {
		creditRemaining = apiKey.CreditLimit - creditUsed
	}

	return map[string]any{
		"token_limit":      apiKey.TokenLimit,
		"token_used":       tokenUsed,
		"token_remaining":  tokenRemaining,
		"credit_limit":     apiKey.CreditLimit,
		"credit_used":      creditUsed,
		"credit_remaining": creditRemaining,
	}, nil
}

// UpdateLastUsed updates the last used timestamp
func (s *UserAPIKeyService) UpdateLastUsed(keyID uint) error {
	return s.apiKeyRepo.UpdateLastUsed(keyID)
}

// DisableAPIKey deactivates an API key
func (s *UserAPIKeyService) DisableAPIKey(keyID uint) error {
	return s.apiKeyRepo.Disable(keyID)
}

// BindChannel binds a channel to an API key
func (s *UserAPIKeyService) BindChannel(keyID, channelID uint) error {
	return s.apiKeyRepo.BindChannel(keyID, channelID)
}

// UnbindChannel unbinds a channel from an API key
func (s *UserAPIKeyService) UnbindChannel(keyID, channelID uint) error {
	return s.apiKeyRepo.UnbindChannel(keyID, channelID)
}

// GetAPIKeyByID returns an API key by ID (alias for GetAPIKey)
func (s *UserAPIKeyService) GetAPIKeyByID(keyID uint) (*models.UserAPIKey, error) {
	return s.apiKeyRepo.GetByID(keyID)
}

// UpdateAPIKey updates fields of an API key (enhanced version)
func (s *UserAPIKeyService) UpdateAPIKeyFields(keyID uint, updates map[string]any) (*models.UserAPIKey, error) {
	allowedFields := map[string]bool{
		"name":         true,
		"is_active":    true,
		"token_limit":  true,
		"credit_limit": true,
		"usage_limit":  true,
		"expires_at":   true,
	}

	for key := range updates {
		if !allowedFields[key] {
			return nil, fmt.Errorf("field %s cannot be updated", key)
		}
	}

	if err := s.apiKeyRepo.UpdateFields(keyID, updates); err != nil {
		return nil, err
	}

	return s.apiKeyRepo.GetByID(keyID)
}
