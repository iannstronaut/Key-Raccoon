package repositories

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"keyraccoon/internal/models"
)

var ErrModelNotFound = errors.New("model not found")

type ModelRepository struct {
	db *gorm.DB
}

func NewModelRepository(db *gorm.DB) *ModelRepository {
	return &ModelRepository{db: db}
}

func (r *ModelRepository) Create(model *models.Model) error {
	return r.db.Create(model).Error
}

func (r *ModelRepository) GetByID(id uint) (*models.Model, error) {
	var model models.Model
	err := r.db.First(&model, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrModelNotFound
	}
	return &model, err
}

func (r *ModelRepository) GetByChannelID(channelID uint) ([]models.Model, error) {
	var modelsList []models.Model
	err := r.db.Where("channel_id = ?", channelID).Order("id ASC").Find(&modelsList).Error
	return modelsList, err
}

func (r *ModelRepository) GetActiveByChannelID(channelID uint) ([]models.Model, error) {
	var modelsList []models.Model
	err := r.db.Where("channel_id = ? AND is_active = ?", channelID, true).Order("id ASC").Find(&modelsList).Error
	return modelsList, err
}

func (r *ModelRepository) Update(model *models.Model) error {
	return r.db.Save(model).Error
}

func (r *ModelRepository) UpdateFields(modelID uint, updates map[string]any) error {
	result := r.db.Model(&models.Model{}).Where("id = ?", modelID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrModelNotFound
	}
	return nil
}

func (r *ModelRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Model{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrModelNotFound
	}
	return nil
}

func (r *ModelRepository) GetByNameAndChannelID(name string, channelID uint) (*models.Model, error) {
	var model models.Model
	err := r.db.Where("LOWER(name) = ? AND channel_id = ?", strings.ToLower(strings.TrimSpace(name)), channelID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrModelNotFound
	}
	return &model, err
}
