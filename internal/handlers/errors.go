package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	appErrors "keyraccoon/internal/utils"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.StatusCode).JSON(fiber.Map{
			"error":   appErr.Code,
			"message": appErr.Message,
		})
	}

	code := fiber.StatusInternalServerError
	if fiberErr, ok := err.(*fiber.Error); ok {
		code = fiberErr.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   "internal_error",
		"message": err.Error(),
	})
}
