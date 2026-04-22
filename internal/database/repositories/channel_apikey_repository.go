package repositories

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"

	"keyraccoon/internal/models"
)

var ErrChannelAPIKeyNotFound = errors.New("channel api key not found")

type ChannelAPIKeyRepository struct {
	db *gorm.DB
}

func NewChannelAPIKeyRepository(db *gorm.DB) *ChannelAPIKeyRepository {
	return &ChannelAPIKeyRepository{db: db}
}

func (r *ChannelAPIKeyRepository) GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sk-" + hex.EncodeToString(bytes), nil
}

func (r *ChannelAPIKeyRepository) Create(apiKey *models.ChannelAPIKey) error {
	return r.db.Create(apiKey).Error
}

func (r *ChannelAPIKeyRepository) GetByID(id uint) (*models.ChannelAPIKey, error) {
	var apiKey models.ChannelAPIKey
	err := r.db.Preload("Channel").First(&apiKey, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrChannelAPIKeyNotFound
	}
	return &apiKey, err
}

func (r *ChannelAPIKeyRepository) GetByKey(key string) (*models.ChannelAPIKey, error) {
	var apiKey models.ChannelAPIKey
	err := r.db.Preload("Channel").Where("api_key = ?", key).First(&apiKey).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrChannelAPIKeyNotFound
	}
	return &apiKey, err
}

func (r *ChannelAPIKeyRepository) GetByChannelID(channelID uint) ([]models.ChannelAPIKey, error) {
	var apiKeys []models.ChannelAPIKey
	err := r.db.Where("channel_id = ?", channelID).Order("id ASC").Find(&apiKeys).Error
	return apiKeys, err
}

func (r *ChannelAPIKeyRepository) GetActiveByChannelID(channelID uint) ([]models.ChannelAPIKey, error) {
	var apiKeys []models.ChannelAPIKey
	err := r.db.Where("channel_id = ? AND is_active = ?", channelID, true).Order("id ASC").Find(&apiKeys).Error
	return apiKeys, err
}

func (r *ChannelAPIKeyRepository) Disable(id uint) error {
	result := r.db.Model(&models.ChannelAPIKey{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrChannelAPIKeyNotFound
	}
	return nil
}

func (r *ChannelAPIKeyRepository) Delete(id uint) error {
	result := r.db.Delete(&models.ChannelAPIKey{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrChannelAPIKeyNotFound
	}
	return nil
}

func (r *ChannelAPIKeyRepository) UpdateLastUsed(id uint) error {
	result := r.db.Model(&models.ChannelAPIKey{}).Where("id = ?", id).Update("updated_at", time.Now().UTC())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrChannelAPIKeyNotFound
	}
	return nil
}
