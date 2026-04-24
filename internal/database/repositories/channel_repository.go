package repositories

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"keyraccoon/internal/models"
)

var ErrChannelNotFound = errors.New("channel not found")

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) Create(channel *models.Channel) error {
	return r.db.Create(channel).Error
}

func (r *ChannelRepository) GetByID(id uint) (*models.Channel, error) {
	var channel models.Channel
	err := r.db.Preload("APIKeys").Preload("Models").First(&channel, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrChannelNotFound
	}
	return &channel, err
}

func (r *ChannelRepository) GetByName(name string) (*models.Channel, error) {
	var channel models.Channel
	err := r.db.Where("LOWER(name) = ?", strings.ToLower(strings.TrimSpace(name))).First(&channel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrChannelNotFound
	}
	return &channel, err
}

func (r *ChannelRepository) GetAll(limit, offset int) ([]models.Channel, int64, error) {
	var channels []models.Channel
	var total int64

	if err := r.db.Model(&models.Channel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Preload("APIKeys").Preload("Models").Order("id ASC").Limit(limit).Offset(offset).Find(&channels).Error; err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

func (r *ChannelRepository) GetActive(limit, offset int) ([]models.Channel, int64, error) {
	var channels []models.Channel
	var total int64

	if err := r.db.Model(&models.Channel{}).Where("is_active = ?", true).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Where("is_active = ?", true).Preload("APIKeys").Preload("Models").Order("id ASC").Limit(limit).Offset(offset).Find(&channels).Error; err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

func (r *ChannelRepository) Update(channel *models.Channel) error {
	return r.db.Save(channel).Error
}

func (r *ChannelRepository) UpdateFields(channelID uint, updates map[string]any) error {
	result := r.db.Model(&models.Channel{}).Where("id = ?", channelID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrChannelNotFound
	}
	return nil
}

func (r *ChannelRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Channel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrChannelNotFound
	}
	return nil
}

func (r *ChannelRepository) GetByUserID(userID uint) ([]models.Channel, error) {
	var user models.User
	err := r.db.Preload("Channels.APIKeys").Preload("Channels.Models").First(&user, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user.Channels, nil
}

func (r *ChannelRepository) BindUserToChannel(userID, channelID uint) error {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
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

	return r.db.Model(&user).Association("Channels").Append(&channel)
}

func (r *ChannelRepository) GetUsersByChannelID(channelID uint) ([]models.User, error) {
	var channel models.Channel
	err := r.db.Preload("Users").First(&channel, channelID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrChannelNotFound
	}
	if err != nil {
		return nil, err
	}
	return channel.Users, nil
}

func (r *ChannelRepository) GetByUserIDWithModels(userID uint) ([]models.Channel, error) {
	var user models.User
	err := r.db.Preload("Channels.Models").First(&user, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user.Channels, nil
}

// IncrementBudgetUsed atomically increments budget_used using SQL expression.
// This is safe for concurrent access — the DB handles the atomicity.
func (r *ChannelRepository) IncrementBudgetUsed(channelID uint, amount float64) error {
	result := r.db.Model(&models.Channel{}).
		Where("id = ?", channelID).
		Update("budget_used", gorm.Expr("budget_used + ?", amount))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrChannelNotFound
	}
	return nil
}

// ResetBudgetUsed resets budget_used to 0 for a channel (admin action).
func (r *ChannelRepository) ResetBudgetUsed(channelID uint) error {
	result := r.db.Model(&models.Channel{}).
		Where("id = ?", channelID).
		Update("budget_used", 0)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrChannelNotFound
	}
	return nil
}

func (r *ChannelRepository) UnbindUserFromChannel(userID, channelID uint) error {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
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

	return r.db.Model(&user).Association("Channels").Delete(&channel)
}
