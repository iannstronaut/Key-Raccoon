package services

import (
	"errors"
	"fmt"
	"strings"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
)

type ChannelService struct {
	channelRepo *repositories.ChannelRepository
	apiKeyRepo  *repositories.ChannelAPIKeyRepository
	modelRepo   *repositories.ModelRepository
	userRepo    *repositories.UserRepository
}

func NewChannelService(
	channelRepo *repositories.ChannelRepository,
	apiKeyRepo *repositories.ChannelAPIKeyRepository,
	modelRepo *repositories.ModelRepository,
	userRepo *repositories.UserRepository,
) *ChannelService {
	return &ChannelService{
		channelRepo: channelRepo,
		apiKeyRepo:  apiKeyRepo,
		modelRepo:   modelRepo,
		userRepo:    userRepo,
	}
}

func (s *ChannelService) CreateChannel(name, channelType, endpoint, description string, budget float64, budgetType string) (*models.Channel, error) {
	if s.channelRepo == nil || s.apiKeyRepo == nil || s.modelRepo == nil {
		return nil, errors.New("channel service dependencies are not initialized")
	}

	name = strings.TrimSpace(name)
	channelType = strings.ToLower(strings.TrimSpace(channelType))
	endpoint = strings.TrimSpace(endpoint)
	description = strings.TrimSpace(description)
	budgetType = strings.ToLower(strings.TrimSpace(budgetType))

	if name == "" {
		return nil, errors.New("channel name required")
	}

	validTypes := map[string]bool{
		"openai":    true,
		"anthr0pic": true,
		"cohere":    true,
		"custom":    true,
	}
	if !validTypes[channelType] {
		return nil, fmt.Errorf("invalid channel type: %s", channelType)
	}

	// Validate endpoint for custom type
	if channelType == "custom" && endpoint == "" {
		return nil, errors.New("endpoint is required for custom channel type")
	}

	// Validate budget_type: default to "price" if empty
	if budgetType == "" {
		budgetType = "price"
	}
	if budgetType != "price" && budgetType != "token" {
		return nil, fmt.Errorf("invalid budget type: %s (must be 'price' or 'token')", budgetType)
	}

	if budget < 0 {
		budget = 0
	}

	if _, err := s.channelRepo.GetByName(name); err == nil {
		return nil, errors.New("channel name already exists")
	} else if !errors.Is(err, repositories.ErrChannelNotFound) {
		return nil, err
	}

	channel := &models.Channel{
		Name:        name,
		Type:        channelType,
		Endpoint:    endpoint,
		Description: description,
		Budget:      budget,
		BudgetType:  budgetType,
		IsActive:    true,
	}

	if err := s.channelRepo.Create(channel); err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}
	return channel, nil
}

func (s *ChannelService) GetChannel(channelID uint) (*models.Channel, error) {
	return s.channelRepo.GetByID(channelID)
}

func (s *ChannelService) GetAllChannels(limit, offset int) ([]models.Channel, int64, error) {
	return s.channelRepo.GetAll(limit, offset)
}

func (s *ChannelService) UpdateChannel(channelID uint, updates map[string]any) (*models.Channel, error) {
	allowedFields := map[string]bool{
		"name":        true,
		"description": true,
		"endpoint":    true,
		"is_active":   true,
		"budget":      true,
		"budget_type": true,
	}

	for key := range updates {
		if !allowedFields[key] {
			return nil, fmt.Errorf("field %s cannot be updated", key)
		}
	}

	if name, ok := updates["name"].(string); ok {
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, errors.New("channel name required")
		}
		updates["name"] = name
	}

	if bt, ok := updates["budget_type"].(string); ok {
		bt = strings.ToLower(strings.TrimSpace(bt))
		if bt != "price" && bt != "token" {
			return nil, fmt.Errorf("invalid budget type: %s (must be 'price' or 'token')", bt)
		}
		updates["budget_type"] = bt
	}

	if err := s.channelRepo.UpdateFields(channelID, updates); err != nil {
		return nil, err
	}

	return s.channelRepo.GetByID(channelID)
}

func (s *ChannelService) DeleteChannel(channelID uint) error {
	return s.channelRepo.Delete(channelID)
}

func (s *ChannelService) AddAPIKey(channelID uint) (*models.ChannelAPIKey, error) {
	channel, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return nil, err
	}

	apiKeyStr, err := s.apiKeyRepo.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate api key: %w", err)
	}

	apiKey := &models.ChannelAPIKey{
		ChannelID: channel.ID,
		APIKey:    apiKeyStr,
		IsActive:  true,
	}

	if err := s.apiKeyRepo.Create(apiKey); err != nil {
		return nil, fmt.Errorf("failed to create api key: %w", err)
	}

	return apiKey, nil
}

func (s *ChannelService) AddAPIKeyWithValue(channelID uint, apiKeyValue string) (*models.ChannelAPIKey, error) {
	channel, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return nil, err
	}

	// Validate API key is not empty
	apiKeyValue = strings.TrimSpace(apiKeyValue)
	if apiKeyValue == "" {
		return nil, errors.New("api key cannot be empty")
	}

	apiKey := &models.ChannelAPIKey{
		ChannelID: channel.ID,
		APIKey:    apiKeyValue,
		IsActive:  true,
	}

	if err := s.apiKeyRepo.Create(apiKey); err != nil {
		return nil, fmt.Errorf("failed to create api key: %w", err)
	}

	return apiKey, nil
}

func (s *ChannelService) GetChannelAPIKeys(channelID uint) ([]models.ChannelAPIKey, error) {
	if _, err := s.channelRepo.GetByID(channelID); err != nil {
		return nil, err
	}
	return s.apiKeyRepo.GetByChannelID(channelID)
}

func (s *ChannelService) RotateAPIKey(channelID uint) (*models.ChannelAPIKey, error) {
	if _, err := s.channelRepo.GetByID(channelID); err != nil {
		return nil, err
	}

	activeKeys, err := s.apiKeyRepo.GetActiveByChannelID(channelID)
	if err != nil {
		return nil, err
	}
	if len(activeKeys) == 0 {
		return nil, errors.New("no active api keys found")
	}

	for _, key := range activeKeys {
		if err := s.apiKeyRepo.Disable(key.ID); err != nil {
			return nil, err
		}
	}

	return s.AddAPIKey(channelID)
}

func (s *ChannelService) RemoveAPIKey(keyID uint) error {
	return s.apiKeyRepo.Delete(keyID)
}

func (s *ChannelService) AddModel(channelID uint, name, displayName string, tokenPrice float64, systemPrompt string) (*models.Model, error) {
	channel, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return nil, err
	}

	name = strings.TrimSpace(name)
	displayName = strings.TrimSpace(displayName)
	systemPrompt = strings.TrimSpace(systemPrompt)

	if name == "" {
		return nil, errors.New("model name required")
	}
	if tokenPrice < 0 {
		return nil, errors.New("token price must be non-negative")
	}

	if _, err := s.modelRepo.GetByNameAndChannelID(name, channelID); err == nil {
		return nil, errors.New("model already exists on this channel")
	} else if !errors.Is(err, repositories.ErrModelNotFound) {
		return nil, err
	}

	model := &models.Model{
		ChannelID:    channel.ID,
		Name:         strings.ToLower(name),
		DisplayName:  displayName,
		TokenPrice:   tokenPrice,
		SystemPrompt: systemPrompt,
		IsActive:     true,
	}

	if err := s.modelRepo.Create(model); err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}
	return model, nil
}

func (s *ChannelService) GetChannelModels(channelID uint) ([]models.Model, error) {
	if _, err := s.channelRepo.GetByID(channelID); err != nil {
		return nil, err
	}
	return s.modelRepo.GetByChannelID(channelID)
}

func (s *ChannelService) UpdateModel(modelID uint, updates map[string]any) (*models.Model, error) {
	allowedFields := map[string]bool{
		"display_name":  true,
		"token_price":   true,
		"system_prompt": true,
		"is_active":     true,
	}

	for key := range updates {
		if !allowedFields[key] {
			return nil, fmt.Errorf("field %s cannot be updated", key)
		}
	}

	if tokenPrice, ok := updates["token_price"].(float64); ok && tokenPrice < 0 {
		return nil, errors.New("token price must be non-negative")
	}

	if err := s.modelRepo.UpdateFields(modelID, updates); err != nil {
		return nil, err
	}

	return s.modelRepo.GetByID(modelID)
}

func (s *ChannelService) DeleteModel(modelID uint) error {
	return s.modelRepo.Delete(modelID)
}

func (s *ChannelService) BindUserToChannel(userID, channelID uint) error {
	if s.userRepo != nil {
		if _, err := s.userRepo.GetByID(userID); err != nil {
			return err
		}
	}
	if _, err := s.channelRepo.GetByID(channelID); err != nil {
		return err
	}
	return s.channelRepo.BindUserToChannel(userID, channelID)
}

func (s *ChannelService) UnbindUserFromChannel(userID, channelID uint) error {
	return s.channelRepo.UnbindUserFromChannel(userID, channelID)
}

func (s *ChannelService) GetUserChannels(userID uint) ([]models.Channel, error) {
	return s.channelRepo.GetByUserID(userID)
}

func (s *ChannelService) GetChannelUsers(channelID uint) ([]models.User, error) {
	channel, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return nil, err
	}
	return s.channelRepo.GetUsersByChannelID(channel.ID)
}

func (s *ChannelService) GetUserChannelsWithModels(userID uint) ([]models.Channel, error) {
	return s.channelRepo.GetByUserIDWithModels(userID)
}

// CheckBudget checks if a channel has remaining budget.
// Returns true if budget is unlimited (0) or budget_used < budget.
func (s *ChannelService) CheckBudget(channelID uint) (bool, error) {
	channel, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return false, err
	}
	return channel.HasBudgetAvailable(), nil
}

// RecordBudgetUsage atomically increments budget_used for a channel.
// Safe for concurrent access.
func (s *ChannelService) RecordBudgetUsage(channelID uint, cost float64) error {
	if cost <= 0 {
		return nil
	}
	return s.channelRepo.IncrementBudgetUsed(channelID, cost)
}

// ResetBudgetUsed resets budget_used to 0 for a channel.
func (s *ChannelService) ResetBudgetUsed(channelID uint) error {
	return s.channelRepo.ResetBudgetUsed(channelID)
}
