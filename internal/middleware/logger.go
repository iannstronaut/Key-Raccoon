package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"keyraccoon/pkg/logger"
)

func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		args := []interface{}{
			"method", c.Method(),
			"path", c.OriginalURL(),
			"status", c.Response().StatusCode(),
			"duration_ms", duration.Milliseconds(),
			"ip", c.IP(),
		}

		if userID, ok := c.Locals("user_id").(uint); ok && userID > 0 {
			args = append(args, "user_id", userID)
		}
		if keyID, ok := c.Locals("api_key_id").(uint); ok && keyID > 0 {
			args = append(args, "api_key_id", keyID)
		}

		if c.Response().StatusCode() >= 400 {
			logger.Error("http request", args...)
		} else {
			logger.Info("http request", args...)
		}
		return err
	}
}
