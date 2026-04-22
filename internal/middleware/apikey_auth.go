package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"keyraccoon/internal/services"
)

// APIKeyAuthMiddleware extracts and validates the API key from the Authorization header.
func APIKeyAuthMiddleware(apiKeyService *services.APIKeyService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "missing authentication",
					"type":    "authentication_error",
				},
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "invalid authentication format",
					"type":    "authentication_error",
				},
			})
		}

		apiKey := parts[1]

		keyData, err := apiKeyService.VerifyAPIKey(apiKey)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "invalid api key",
					"type":    "authentication_error",
				},
			})
		}

		c.Locals("api_key_id", keyData.ID)
		c.Locals("user_id", keyData.UserID)
		c.Locals("api_key_channels", keyData.Channels)

		return c.Next()
	}
}
