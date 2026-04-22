package repositories

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"keyraccoon/internal/models"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("LOWER(email) = ?", strings.ToLower(email)).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *UserRepository) GetAll(limit, offset int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Order("id ASC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) UpdateFields(userID uint, updates map[string]any) error {
	result := r.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) Delete(id uint) error {
	result := r.db.Delete(&models.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) UpdateLastLogin(userID uint, loginTime time.Time) error {
	return r.UpdateFields(userID, map[string]any{
		"last_login": loginTime,
	})
}

func (r *UserRepository) UpdateTokenUsage(userID uint, tokenUsed int64) error {
	result := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("token_used", gorm.Expr("token_used + ?", tokenUsed))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) UpdateCreditUsage(userID uint, creditUsed float64) error {
	result := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("credit_used", gorm.Expr("credit_used + ?", creditUsed))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) HasTokenAvailable(userID uint, requiredToken int64) (bool, error) {
	var user models.User
	err := r.db.Select("token_limit, token_used").First(&user, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, ErrUserNotFound
	}
	if err != nil {
		return false, err
	}

	if user.TokenLimit == -1 {
		return true, nil
	}

	return (user.TokenUsed + requiredToken) <= user.TokenLimit, nil
}

func (r *UserRepository) HasCreditAvailable(userID uint, requiredCredit float64) (bool, error) {
	var user models.User
	err := r.db.Select("credit_limit, credit_used").First(&user, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, ErrUserNotFound
	}
	if err != nil {
		return false, err
	}

	if user.CreditLimit == -1 {
		return true, nil
	}

	return (user.CreditUsed + requiredCredit) <= user.CreditLimit, nil
}
