package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/handlers"
	"keyraccoon/internal/middleware"
	"keyraccoon/internal/services"
)

func SetupAPIKeyRoutes(router fiber.Router, db *gorm.DB) {
	apiKeyRepo := repositories.NewAPIKeyRepository(db)
	userRepo := repositories.NewUserRepository(db)

	apiKeyService := services.NewAPIKeyService(apiKeyRepo, userRepo)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService)

	apikeys := router.Group("/api-keys", middleware.AuthMiddleware)

	apikeys.Post("", apiKeyHandler.CreateAPIKey)
	apikeys.Get("", apiKeyHandler.GetUserAPIKeys)
	apikeys.Get("/:id", apiKeyHandler.GetAPIKey)
	apikeys.Put("/:id", apiKeyHandler.UpdateAPIKey)
	apikeys.Get("/:id/usage", apiKeyHandler.GetAPIKeyUsage)
	apikeys.Delete("/:id", apiKeyHandler.DeleteAPIKey)
	apikeys.Post("/:id/channels/:channelID/bind", apiKeyHandler.BindChannel)
	apikeys.Delete("/:id/channels/:channelID/bind", apiKeyHandler.UnbindChannel)
}
