package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"keyraccoon/internal/config"
	"keyraccoon/internal/utils"
)

func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
		if tokenString == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing bearer token")
		}

		cfg := config.Get()
		if cfg == nil {
			return fiber.NewError(fiber.StatusInternalServerError, "config is not initialized")
		}

		claims, err := utils.ValidateToken(tokenString, cfg.JWTSecret)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)

		return c.Next()
	}
}
