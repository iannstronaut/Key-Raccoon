package handlers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"keyraccoon/internal/config"
	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
	"keyraccoon/internal/services"
	"keyraccoon/internal/utils"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	if strings.TrimSpace(req.Email) == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email and password are required",
		})
	}

	user, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	cfg := config.Get()
	if cfg == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "config is not initialized",
		})
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role, cfg.JWTSecret, cfg.JWTExpire)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate access token",
		})
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, user.Role, cfg.JWTSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate refresh token",
		})
	}

	return c.JSON(fiber.Map{
		"user":          sanitizeUser(user),
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    cfg.JWTExpire * 60,
	})
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req struct {
		Email       string  `json:"email"`
		Password    string  `json:"password"`
		Name        string  `json:"name"`
		Role        string  `json:"role"`
		TokenLimit  int64   `json:"token_limit"`
		CreditLimit float64 `json:"credit_limit"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	user, err := h.userService.CreateUser(req.Email, req.Password, req.Name, req.Role, req.TokenLimit, req.CreditLimit)
	if err != nil {
		return statusFromError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(sanitizeUser(user))
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	userID, err := parseUserID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	user, err := h.userService.GetUser(userID)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(sanitizeUser(user))
}

func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	users, total, err := h.userService.GetAllUsers(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	responseUsers := make([]fiber.Map, 0, len(users))
	for i := range users {
		responseUsers = append(responseUsers, sanitizeUser(&users[i]))
	}

	return c.JSON(fiber.Map{
		"users":  responseUsers,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	userID, err := parseUserID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	var req struct {
		Name        *string  `json:"name"`
		IsActive    *bool    `json:"is_active"`
		TokenLimit  *int64   `json:"token_limit"`
		CreditLimit *float64 `json:"credit_limit"`
		Role        *string  `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	updates := make(map[string]any)
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.TokenLimit != nil {
		updates["token_limit"] = *req.TokenLimit
	}
	if req.CreditLimit != nil {
		updates["credit_limit"] = *req.CreditLimit
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}

	user, err := h.userService.UpdateUser(userID, updates)
	if err != nil {
		return statusFromError(c, err)
	}

	return c.JSON(sanitizeUser(user))
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	userID, err := parseUserID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "user deleted successfully",
	})
}

func (h *UserHandler) GetUserUsage(c *fiber.Ctx) error {
	userID, err := parseUserID(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	usage, err := h.userService.GetUserUsage(userID)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(usage)
}

func parseUserID(raw string) (uint, error) {
	value, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

func sanitizeUser(user *models.User) fiber.Map {
	return fiber.Map{
		"id":           user.ID,
		"email":        user.Email,
		"name":         user.Name,
		"role":         user.Role,
		"is_active":    user.IsActive,
		"token_limit":  user.TokenLimit,
		"credit_limit": user.CreditLimit,
		"token_used":   user.TokenUsed,
		"credit_used":  user.CreditUsed,
		"last_login":   user.LastLogin,
		"created_at":   user.CreatedAt,
		"updated_at":   user.UpdatedAt,
	}
}

func statusFromError(c *fiber.Ctx, err error) error {
	status := fiber.StatusBadRequest
	if errors.Is(err, repositories.ErrUserNotFound) {
		status = fiber.StatusNotFound
	}
	return c.Status(status).JSON(fiber.Map{
		"error": err.Error(),
	})
}
