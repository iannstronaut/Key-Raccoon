package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"keyraccoon/internal/config"
	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
)

type APIKeyService struct {
	apiKeyRepo *repositories.APIKeyRepository
	userRepo   *repositories.UserRepository
	redis      *redis.Client
}

func NewAPIKeyService(
	apiKeyRepo *repositories.APIKeyRepository,
	userRepo *repositories.UserRepository,
) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo: apiKeyRepo,
		userRepo:   userRepo,
		redis:      config.GetRedis(),
	}
}

// CreateAPIKey creates a new API key for a user.
func (s *APIKeyService) CreateAPIKey(userID uint, name string, tokenLimit int64, creditLimit float64) (*models.APIKey, error) {
	if s.apiKeyRepo == nil || s.userRepo == nil {
		return nil, errors.New("api key service dependencies are not initialized")
	}

	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	keyStr, err := s.apiKeyRepo.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate api key: %w", err)
	}

	apiKey := &models.APIKey{
		UserID:      userID,
		Key:         keyStr,
		Name:        name,
		TokenLimit:  tokenLimit,
		CreditLimit: creditLimit,
		TokenUsed:   0,
		CreditUsed:  0,
		IsActive:    true,
	}

	if err := s.apiKeyRepo.Create(apiKey); err != nil {
		return nil, fmt.Errorf("failed to create api key: %w", err)
	}

	return apiKey, nil
}

// GetUserAPIKeys returns all API keys for a user.
func (s *APIKeyService) GetUserAPIKeys(userID uint) ([]models.APIKey, error) {
	return s.apiKeyRepo.GetByUserID(userID)
}

// GetAPIKeyByID returns an API key by ID.
func (s *APIKeyService) GetAPIKeyByID(keyID uint) (*models.APIKey, error) {
	return s.apiKeyRepo.GetByID(keyID)
}

// GetAPIKeyByKey returns an API key by key string.
func (s *APIKeyService) GetAPIKeyByKey(key string) (*models.APIKey, error) {
	return s.apiKeyRepo.GetByKey(key)
}

// VerifyAPIKey verifies that an API key exists and is active.
func (s *APIKeyService) VerifyAPIKey(key string) (*models.APIKey, error) {
	apiKey, err := s.apiKeyRepo.GetByKey(key)
	if err != nil {
		return nil, errors.New("invalid api key")
	}

	if !apiKey.IsActive {
		return nil, errors.New("api key is disabled")
	}

	return apiKey, nil
}

// UpdateAPIKey updates fields of an API key.
func (s *APIKeyService) UpdateAPIKey(keyID uint, updates map[string]any) (*models.APIKey, error) {
	allowedFields := map[string]bool{
		"name":         true,
		"is_active":    true,
		"token_limit":  true,
		"credit_limit": true,
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

// UpdateLastUsed updates the last used timestamp.
func (s *APIKeyService) UpdateLastUsed(keyID uint) error {
	return s.apiKeyRepo.UpdateLastUsed(keyID)
}

// DisableAPIKey deactivates an API key.
func (s *APIKeyService) DisableAPIKey(keyID uint) error {
	return s.apiKeyRepo.Disable(keyID)
}

// DeleteAPIKey removes an API key.
func (s *APIKeyService) DeleteAPIKey(keyID uint) error {
	return s.apiKeyRepo.Delete(keyID)
}

// ============ USAGE TRACKING ============

// HasTokenAvailable checks token availability.
func (s *APIKeyService) HasTokenAvailable(keyID uint, requiredToken int64) (bool, error) {
	return s.apiKeyRepo.HasTokenAvailable(keyID, requiredToken)
}

// HasCreditAvailable checks credit availability.
func (s *APIKeyService) HasCreditAvailable(keyID uint, requiredCredit float64) (bool, error) {
	return s.apiKeyRepo.HasCreditAvailable(keyID, requiredCredit)
}

// RecordTokenUsage records token usage in Redis (real-time) or database (fallback).
func (s *APIKeyService) RecordTokenUsage(keyID uint, tokens int64) error {
	if s.redis != nil {
		ctx := context.Background()
		key := fmt.Sprintf("apikey:%d:tokens", keyID)

		if err := s.redis.IncrBy(ctx, key, tokens).Err(); err == nil {
			s.redis.Expire(ctx, key, 24*time.Hour)
			return nil
		}
	}

	return s.apiKeyRepo.UpdateTokenUsage(keyID, tokens)
}

// RecordCreditUsage records credit usage in Redis (real-time) or database (fallback).
func (s *APIKeyService) RecordCreditUsage(keyID uint, credit float64) error {
	if s.redis != nil {
		ctx := context.Background()
		key := fmt.Sprintf("apikey:%d:credit", keyID)

		if err := s.redis.IncrByFloat(ctx, key, credit).Err(); err == nil {
			s.redis.Expire(ctx, key, 24*time.Hour)
			return nil
		}
	}

	return s.apiKeyRepo.UpdateCreditUsage(keyID, credit)
}

// GetRealtimeUsage returns combined database and Redis usage stats.
func (s *APIKeyService) GetRealtimeUsage(keyID uint) (map[string]any, error) {
	apiKey, err := s.apiKeyRepo.GetByID(keyID)
	if err != nil {
		return nil, err
	}

	tokenUsed := apiKey.TokenUsed
	creditUsed := apiKey.CreditUsed

	if s.redis != nil {
		ctx := context.Background()

		rtTokens, err := s.redis.Get(ctx, fmt.Sprintf("apikey:%d:tokens", keyID)).Int64()
		if err == nil {
			tokenUsed += rtTokens
		}

		rtCredit, err := s.redis.Get(ctx, fmt.Sprintf("apikey:%d:credit", keyID)).Float64()
		if err == nil {
			creditUsed += rtCredit
		}
	}

	tokenRemaining := int64(-1)
	if apiKey.TokenLimit != -1 {
		tokenRemaining = apiKey.TokenLimit - tokenUsed
	}

	creditRemaining := float64(-1)
	if apiKey.CreditLimit != -1 {
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

// ============ CHANNEL BINDING ============

// BindChannel binds a channel to an API key.
func (s *APIKeyService) BindChannel(keyID, channelID uint) error {
	return s.apiKeyRepo.BindChannel(keyID, channelID)
}

// UnbindChannel unbinds a channel from an API key.
func (s *APIKeyService) UnbindChannel(keyID, channelID uint) error {
	return s.apiKeyRepo.UnbindChannel(keyID, channelID)
}
