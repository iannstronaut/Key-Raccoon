package handlers

import (
	"github.com/gofiber/fiber/v2"

	"keyraccoon/internal/config"
	"keyraccoon/internal/utils"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "refresh_token is required",
		})
	}

	cfg := config.Get()
	if cfg == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "config is not initialized",
		})
	}

	claims, err := utils.ValidateToken(req.RefreshToken, cfg.JWTSecret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid refresh token",
		})
	}
	if claims.TokenType != "refresh" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token type",
		})
	}

	accessToken, err := utils.GenerateAccessToken(claims.UserID, claims.Email, claims.Role, cfg.JWTSecret, cfg.JWTExpire)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate access token",
		})
	}

	return c.JSON(fiber.Map{
		"access_token": accessToken,
		"expires_in":   cfg.JWTExpire * 60,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "logged out successfully",
	})
}
