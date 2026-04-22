package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"keyraccoon/internal/config"
	"keyraccoon/internal/utils"
	"keyraccoon/pkg/logger"
)

func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing authorization header",
		})
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid authorization header format",
		})
	}

	cfg := config.Get()
	if cfg == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "config is not initialized",
		})
	}

	claims, err := utils.ValidateToken(strings.TrimSpace(parts[1]), cfg.JWTSecret)
	if err != nil {
		logger.Error("token validation failed", "error", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	c.Locals("user_id", claims.UserID)
	c.Locals("user_email", claims.Email)
	c.Locals("user_role", claims.Role)

	return c.Next()
}

func RequireAuth() fiber.Handler {
	return AuthMiddleware
}

func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, _ := c.Locals("user_role").(string)
		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient permissions",
		})
	}
}

func SuperAdminMiddleware(c *fiber.Ctx) error {
	userRole, _ := c.Locals("user_role").(string)
	if userRole != "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "only superadmin can access this",
		})
	}
	return c.Next()
}

func AdminMiddleware(c *fiber.Ctx) error {
	userRole, _ := c.Locals("user_role").(string)
	if userRole != "superadmin" && userRole != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "admin access required",
		})
	}
	return c.Next()
}
