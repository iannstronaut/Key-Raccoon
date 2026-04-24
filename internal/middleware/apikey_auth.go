package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"keyraccoon/internal/services"
)

// APIKeyAuthMiddleware extracts and validates the API key from the Authorization header.
// Uses the UserAPIKey system for authentication.
func APIKeyAuthMiddleware(
	userAPIKeyService *services.UserAPIKeyService,
) fiber.Handler {
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

		// Verify UserAPIKey
		userKeyData, err := userAPIKeyService.VerifyAPIKey(apiKey)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "invalid api key",
					"type":    "authentication_error",
				},
			})
		}

		// Set context for UserAPIKey
		c.Locals("api_key_id", userKeyData.ID)
		c.Locals("user_id", userKeyData.UserID)
		c.Locals("user_api_key", userKeyData)
		c.Locals("api_key_type", "user_api_key")
		
		// Set channels - use all channels if empty, otherwise use specified
		if len(userKeyData.Channels) > 0 {
			c.Locals("api_key_channels", userKeyData.Channels)
		} else {
			// Empty means all channels allowed - handler will handle this
			c.Locals("api_key_channels", []interface{}{})
		}
		
		// Set allowed models
		if len(userKeyData.Models) > 0 {
			modelIDs := make([]uint, 0, len(userKeyData.Models))
			for _, m := range userKeyData.Models {
				modelIDs = append(modelIDs, m.ModelID)
			}
			c.Locals("allowed_models", modelIDs)
		}
		
		return c.Next()
	}
}
