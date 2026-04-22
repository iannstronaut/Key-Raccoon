package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
	"keyraccoon/internal/utils"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(email, password, name, role string, tokenLimit int64, creditLimit float64) (*models.User, error) {
	if s.userRepo == nil {
		return nil, errors.New("user repository is not initialized")
	}

	email = strings.TrimSpace(strings.ToLower(email))
	name = strings.TrimSpace(name)
	role = strings.TrimSpace(strings.ToLower(role))

	if email == "" || password == "" || name == "" {
		return nil, errors.New("email, password, and name are required")
	}

	if _, err := s.userRepo.GetByEmail(email); err == nil {
		return nil, errors.New("email already exists")
	} else if !errors.Is(err, repositories.ErrUserNotFound) {
		return nil, err
	}

	if role != "superadmin" && role != "admin" && role != "user" {
		return nil, errors.New("invalid role. must be 'superadmin', 'admin', or 'user'")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	if tokenLimit == 0 {
		tokenLimit = 1_000_000
	}
	if creditLimit == 0 {
		creditLimit = -1
	}

	user := &models.User{
		Email:       email,
		Password:    hashedPassword,
		Name:        name,
		Role:        role,
		IsActive:    true,
		TokenLimit:  tokenLimit,
		CreditLimit: creditLimit,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *UserService) Login(email, password string) (*models.User, error) {
	if s.userRepo == nil {
		return nil, errors.New("user repository is not initialized")
	}

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("user account is disabled")
	}

	if err := utils.VerifyPassword(password, user.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	now := time.Now().UTC()
	user.LastLogin = &now
	if err := s.userRepo.UpdateLastLogin(user.ID, now); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUser(userID uint) (*models.User, error) {
	return s.userRepo.GetByID(userID)
}

func (s *UserService) FindByEmail(email string) (*models.User, error) {
	return s.userRepo.GetByEmail(email)
}

func (s *UserService) GetAllUsers(limit, offset int) ([]models.User, int64, error) {
	return s.userRepo.GetAll(limit, offset)
}

func (s *UserService) UpdateUser(userID uint, updates map[string]any) (*models.User, error) {
	if s.userRepo == nil {
		return nil, errors.New("user repository is not initialized")
	}

	allowedFields := map[string]bool{
		"name":         true,
		"is_active":    true,
		"token_limit":  true,
		"credit_limit": true,
		"role":         true,
	}

	for key := range updates {
		if !allowedFields[key] {
			return nil, fmt.Errorf("field %s cannot be updated", key)
		}
	}

	if role, ok := updates["role"].(string); ok {
		role = strings.ToLower(strings.TrimSpace(role))
		if role != "superadmin" && role != "admin" && role != "user" {
			return nil, errors.New("invalid role. must be 'superadmin', 'admin', or 'user'")
		}
		updates["role"] = role
	}

	if err := s.userRepo.UpdateFields(userID, updates); err != nil {
		return nil, err
	}

	return s.userRepo.GetByID(userID)
}

func (s *UserService) DisableUser(userID uint) error {
	_, err := s.UpdateUser(userID, map[string]any{"is_active": false})
	return err
}

func (s *UserService) DeleteUser(userID uint) error {
	return s.userRepo.Delete(userID)
}

func (s *UserService) SetTokenLimit(userID uint, limit int64) error {
	_, err := s.UpdateUser(userID, map[string]any{
		"token_limit": limit,
	})
	if err != nil {
		return err
	}
	return s.userRepo.UpdateFields(userID, map[string]any{"token_used": 0})
}

func (s *UserService) SetCreditLimit(userID uint, limit float64) error {
	_, err := s.UpdateUser(userID, map[string]any{
		"credit_limit": limit,
	})
	if err != nil {
		return err
	}
	return s.userRepo.UpdateFields(userID, map[string]any{"credit_used": 0})
}

func (s *UserService) GetUserUsage(userID uint) (map[string]any, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	tokenRemaining := int64(-1)
	if user.TokenLimit != -1 {
		tokenRemaining = user.TokenLimit - user.TokenUsed
	}

	creditRemaining := float64(-1)
	if user.CreditLimit != -1 {
		creditRemaining = user.CreditLimit - user.CreditUsed
	}

	return map[string]any{
		"token_limit":      user.TokenLimit,
		"token_used":       user.TokenUsed,
		"token_remaining":  tokenRemaining,
		"credit_limit":     user.CreditLimit,
		"credit_used":      user.CreditUsed,
		"credit_remaining": creditRemaining,
	}, nil
}
