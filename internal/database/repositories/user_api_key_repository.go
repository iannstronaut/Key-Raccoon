package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"keyraccoon/internal/models"
)

var (
	ErrUserAPIKeyNotFound = errors.New("user api key not found")
	ErrUserAPIKeyExists   = errors.New("user api key already exists")
)

type UserAPIKeyRepository struct {
	db *gorm.DB
}

func NewUserAPIKeyRepository(db *gorm.DB) *UserAPIKeyRepository {
	return &UserAPIKeyRepository{db: db}
}

func (r *UserAPIKeyRepository) Create(apiKey *models.UserAPIKey) error {
	return r.db.Create(apiKey).Error
}

func (r *UserAPIKeyRepository) GetByID(id uint) (*models.UserAPIKey, error) {
	var apiKey models.UserAPIKey
	err := r.db.Preload("User").Preload("Channels").Preload("Models.Model").First(&apiKey, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserAPIKeyNotFound
	}
	return &apiKey, err
}

func (r *UserAPIKeyRepository) GetByKey(key string) (*models.UserAPIKey, error) {
	var apiKey models.UserAPIKey
	err := r.db.Preload("User").Preload("Channels").Preload("Models.Model").
		Where("key = ?", key).First(&apiKey).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserAPIKeyNotFound
	}
	return &apiKey, err
}

func (r *UserAPIKeyRepository) GetByUserID(userID uint) ([]models.UserAPIKey, error) {
	var apiKeys []models.UserAPIKey
	err := r.db.Preload("Channels").Preload("Models.Model").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&apiKeys).Error
	return apiKeys, err
}

func (r *UserAPIKeyRepository) GetAll(limit, offset int) ([]models.UserAPIKey, int64, error) {
	var apiKeys []models.UserAPIKey
	var total int64

	if err := r.db.Model(&models.UserAPIKey{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("User").Preload("Channels").Preload("Models.Model").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&apiKeys).Error

	return apiKeys, total, err
}

func (r *UserAPIKeyRepository) Update(apiKey *models.UserAPIKey) error {
	return r.db.Save(apiKey).Error
}

func (r *UserAPIKeyRepository) Delete(id uint) error {
	return r.db.Delete(&models.UserAPIKey{}, id).Error
}

func (r *UserAPIKeyRepository) IncrementUsage(id uint) error {
	now := time.Now()
	return r.db.Model(&models.UserAPIKey{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"usage_count":  gorm.Expr("usage_count + 1"),
			"last_used_at": now,
		}).Error
}

// AddChannel adds a channel to the API key
func (r *UserAPIKeyRepository) AddChannel(apiKeyID, channelID uint) error {
	return r.db.Exec(`
		INSERT INTO user_api_key_channels (user_api_key_id, channel_id)
		VALUES (?, ?)
		ON CONFLICT DO NOTHING
	`, apiKeyID, channelID).Error
}

// RemoveChannel removes a channel from the API key
func (r *UserAPIKeyRepository) RemoveChannel(apiKeyID, channelID uint) error {
	return r.db.Exec(`
		DELETE FROM user_api_key_channels
		WHERE user_api_key_id = ? AND channel_id = ?
	`, apiKeyID, channelID).Error
}

// AddModel adds a model to the API key
func (r *UserAPIKeyRepository) AddModel(apiKeyID, modelID uint) error {
	model := &models.UserAPIKeyModel{
		UserAPIKeyID: apiKeyID,
		ModelID:      modelID,
	}
	return r.db.Create(model).Error
}

// RemoveModel removes a model from the API key
func (r *UserAPIKeyRepository) RemoveModel(apiKeyID, modelID uint) error {
	return r.db.Where("user_api_key_id = ? AND model_id = ?", apiKeyID, modelID).
		Delete(&models.UserAPIKeyModel{}).Error
}

// GetActiveByKey gets an active, non-expired API key
func (r *UserAPIKeyRepository) GetActiveByKey(key string) (*models.UserAPIKey, error) {
	var apiKey models.UserAPIKey
	err := r.db.Preload("User").Preload("Channels").Preload("Models.Model").
		Where("key = ? AND is_active = ? AND (expires_at IS NULL OR expires_at > ?)", 
			key, true, time.Now()).
		First(&apiKey).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserAPIKeyNotFound
	}
	return &apiKey, err
}
