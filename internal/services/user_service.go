package services

import (
	"errors"

	"gorm.io/gorm"

	"keyraccoon/internal/models"
	"keyraccoon/internal/utils"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) CreateUser(user *models.User) error {
	if s.db == nil {
		return errors.New("database is not initialized")
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	if user.Role == "" {
		user.Role = "user"
	}

	return s.db.Create(user).Error
}

func (s *UserService) FindByEmail(email string) (*models.User, error) {
	if s.db == nil {
		return nil, errors.New("database is not initialized")
	}

	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
