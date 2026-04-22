package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/handlers"
	"keyraccoon/internal/middleware"
	"keyraccoon/internal/services"
)

// SetupAPIV1Routes registers OpenAI-compatible API v1 routes.
func SetupAPIV1Routes(app *fiber.App, db *gorm.DB) {
	apiKeyRepo := repositories.NewAPIKeyRepository(db)
	userRepo := repositories.NewUserRepository(db)
	channelRepo := repositories.NewChannelRepository(db)
	apiKeyChannelRepo := repositories.NewChannelAPIKeyRepository(db)
	modelRepo := repositories.NewModelRepository(db)
	proxyRepo := repositories.NewProxyRepository(db)

	apiKeyService := services.NewAPIKeyService(apiKeyRepo, userRepo)
	channelService := services.NewChannelService(channelRepo, apiKeyChannelRepo, modelRepo, userRepo)
	proxyService := services.NewProxyService(proxyRepo)

	chatHandler := handlers.NewChatHandler(apiKeyService, channelService, proxyService)

	api := app.Group("/api/v1", middleware.APIKeyAuthMiddleware(apiKeyService))

	api.Post("/chat/completions", chatHandler.ChatCompletion)
	api.Post("/embeddings", chatHandler.Embeddings)
	api.Get("/models", chatHandler.ListModels)
}
