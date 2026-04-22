package repositories

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"

	"keyraccoon/internal/models"
)

var ErrAPIKeyNotFound = errors.New("api key not found")

type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// GenerateAPIKey generates a random API key string.
func (r *APIKeyRepository) GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "pk-" + hex.EncodeToString(bytes), nil
}

// Create creates a new API key.
func (r *APIKeyRepository) Create(apiKey *models.APIKey) error {
	return r.db.Create(apiKey).Error
}

// GetByID retrieves an API key by ID with preloaded relations.
func (r *APIKeyRepository) GetByID(id uint) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.Preload("User").Preload("Channels").First(&apiKey, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAPIKeyNotFound
	}
	return &apiKey, err
}

// GetByKey retrieves an API key by key string with preloaded relations.
func (r *APIKeyRepository) GetByKey(key string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.Preload("User").Preload("Channels").
		Where("key = ?", key).First(&apiKey).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAPIKeyNotFound
	}
	return &apiKey, err
}

// GetByUserID retrieves all API keys for a user.
func (r *APIKeyRepository) GetByUserID(userID uint) ([]models.APIKey, error) {
	var apiKeys []models.APIKey
	err := r.db.Where("user_id = ?", userID).
		Preload("Channels").Find(&apiKeys).Error
	return apiKeys, err
}

// GetActiveByUserID retrieves active API keys for a user.
func (r *APIKeyRepository) GetActiveByUserID(userID uint) ([]models.APIKey, error) {
	var apiKeys []models.APIKey
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).
		Preload("Channels").Find(&apiKeys).Error
	return apiKeys, err
}

// Update saves changes to an API key.
func (r *APIKeyRepository) Update(apiKey *models.APIKey) error {
	return r.db.Save(apiKey).Error
}

// UpdateFields updates specific fields of an API key.
func (r *APIKeyRepository) UpdateFields(keyID uint, updates map[string]any) error {
	result := r.db.Model(&models.APIKey{}).Where("id = ?", keyID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

// Disable deactivates an API key.
func (r *APIKeyRepository) Disable(id uint) error {
	result := r.db.Model(&models.APIKey{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

// Delete removes an API key.
func (r *APIKeyRepository) Delete(id uint) error {
	result := r.db.Delete(&models.APIKey{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

// UpdateTokenUsage increments token usage.
func (r *APIKeyRepository) UpdateTokenUsage(keyID uint, tokens int64) error {
	result := r.db.Model(&models.APIKey{}).Where("id = ?", keyID).
		Update("token_used", gorm.Expr("token_used + ?", tokens))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

// UpdateCreditUsage increments credit usage.
func (r *APIKeyRepository) UpdateCreditUsage(keyID uint, credit float64) error {
	result := r.db.Model(&models.APIKey{}).Where("id = ?", keyID).
		Update("credit_used", gorm.Expr("credit_used + ?", credit))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

// UpdateLastUsed updates the last used timestamp.
func (r *APIKeyRepository) UpdateLastUsed(keyID uint) error {
	now := time.Now().UTC()
	return r.UpdateFields(keyID, map[string]any{"last_used": now})
}

// HasTokenAvailable checks if an API key has enough token quota.
func (r *APIKeyRepository) HasTokenAvailable(keyID uint, requiredToken int64) (bool, error) {
	var apiKey models.APIKey
	err := r.db.Select("token_limit, token_used").First(&apiKey, keyID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, ErrAPIKeyNotFound
	}
	if err != nil {
		return false, err
	}

	if apiKey.TokenLimit == -1 {
		return true, nil
	}

	return (apiKey.TokenUsed + requiredToken) <= apiKey.TokenLimit, nil
}

// HasCreditAvailable checks if an API key has enough credit quota.
func (r *APIKeyRepository) HasCreditAvailable(keyID uint, requiredCredit float64) (bool, error) {
	var apiKey models.APIKey
	err := r.db.Select("credit_limit, credit_used").First(&apiKey, keyID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, ErrAPIKeyNotFound
	}
	if err != nil {
		return false, err
	}

	if apiKey.CreditLimit == -1 {
		return true, nil
	}

	return (apiKey.CreditUsed + requiredCredit) <= apiKey.CreditLimit, nil
}

// BindChannel binds a channel to an API key.
func (r *APIKeyRepository) BindChannel(apiKeyID, channelID uint) error {
	var apiKey models.APIKey
	if err := r.db.First(&apiKey, apiKeyID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAPIKeyNotFound
		}
		return err
	}

	var channel models.Channel
	if err := r.db.First(&channel, channelID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChannelNotFound
		}
		return err
	}

	return r.db.Model(&apiKey).Association("Channels").Append(&channel)
}

// UnbindChannel unbinds a channel from an API key.
func (r *APIKeyRepository) UnbindChannel(apiKeyID, channelID uint) error {
	var apiKey models.APIKey
	if err := r.db.First(&apiKey, apiKeyID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAPIKeyNotFound
		}
		return err
	}

	var channel models.Channel
	if err := r.db.First(&channel, channelID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChannelNotFound
		}
		return err
	}

	return r.db.Model(&apiKey).Association("Channels").Delete(&channel)
}
