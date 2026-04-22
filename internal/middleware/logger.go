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
		logger.Info(
			"http request",
			"method", c.Method(),
			"path", c.OriginalURL(),
			"status", c.Response().StatusCode(),
			"duration", time.Since(start).String(),
		)
		return err
	}
}
