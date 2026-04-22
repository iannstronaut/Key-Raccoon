package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"keyraccoon/internal/config"
)

func RegisterHealthRoutes(router fiber.Router) {
	router.Get("/health", HealthCheck)
}

func HealthCheck(c *fiber.Ctx) error {
	status := fiber.Map{
		"status":      "ok",
		"service":     "keyraccoon",
		"version":     "0.1.0",
		"database_ok": config.GetDB() != nil,
		"redis_ok":    false,
	}

	if client := config.GetRedis(); client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		status["redis_ok"] = client.Ping(ctx).Err() == nil
	}

	return c.Status(fiber.StatusOK).JSON(status)
}
