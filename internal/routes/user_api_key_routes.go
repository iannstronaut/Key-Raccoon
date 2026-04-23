package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/handlers"
	"keyraccoon/internal/middleware"
	"keyraccoon/internal/services"
)

func SetupUserAPIKeyRoutes(router fiber.Router, db *gorm.DB) {
	apiKeyRepo := repositories.NewUserAPIKeyRepository(db)
	userRepo := repositories.NewUserRepository(db)
	channelRepo := repositories.NewChannelRepository(db)
	modelRepo := repositories.NewModelRepository(db)

	apiKeyService := services.NewUserAPIKeyService(apiKeyRepo, userRepo, channelRepo, modelRepo)
	apiKeyHandler := handlers.NewUserAPIKeyHandler(apiKeyService)

	apiKeys := router.Group("/user-api-keys", middleware.AuthMiddleware)
	
	// Admin can view all API keys
	apiKeys.Get("", middleware.AdminMiddleware, apiKeyHandler.GetAllAPIKeys)
	apiKeys.Get("/:id", apiKeyHandler.GetAPIKey)
	
	// Get API keys for a specific user
	apiKeys.Get("/user/:userID", apiKeyHandler.GetUserAPIKeys)
	
	// Admin can create/update/delete API keys
	apiKeys.Post("", middleware.AdminMiddleware, apiKeyHandler.CreateAPIKey)
	apiKeys.Put("/:id", middleware.AdminMiddleware, apiKeyHandler.UpdateAPIKey)
	apiKeys.Delete("/:id", middleware.AdminMiddleware, apiKeyHandler.DeleteAPIKey)
	
	// Manage channels
	apiKeys.Post("/:id/channels", middleware.AdminMiddleware, apiKeyHandler.AddChannel)
	apiKeys.Delete("/:id/channels/:channelID", middleware.AdminMiddleware, apiKeyHandler.RemoveChannel)
	
	// Manage models
	apiKeys.Post("/:id/models", middleware.AdminMiddleware, apiKeyHandler.AddModel)
	apiKeys.Delete("/:id/models/:modelID", middleware.AdminMiddleware, apiKeyHandler.RemoveModel)
}
