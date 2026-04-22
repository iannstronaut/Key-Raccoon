package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/handlers"
	"keyraccoon/internal/middleware"
	"keyraccoon/internal/services"
)

func SetupChannelRoutes(app *fiber.App, db *gorm.DB) {
	channelRepo := repositories.NewChannelRepository(db)
	apiKeyRepo := repositories.NewChannelAPIKeyRepository(db)
	modelRepo := repositories.NewModelRepository(db)
	userRepo := repositories.NewUserRepository(db)

	channelService := services.NewChannelService(channelRepo, apiKeyRepo, modelRepo, userRepo)
	channelHandler := handlers.NewChannelHandler(channelService)

	channels := app.Group("/channels", middleware.AuthMiddleware)
	channels.Get("", channelHandler.GetAllChannels)
	channels.Get("/:id", channelHandler.GetChannel)

	channels.Post("", middleware.AdminMiddleware, channelHandler.CreateChannel)
	channels.Put("/:id", middleware.AdminMiddleware, channelHandler.UpdateChannel)
	channels.Delete("/:id", middleware.AdminMiddleware, channelHandler.DeleteChannel)

	channels.Post("/:id/api-keys", middleware.AdminMiddleware, channelHandler.AddAPIKey)
	channels.Get("/:id/api-keys", middleware.AdminMiddleware, channelHandler.GetChannelAPIKeys)
	channels.Post("/:id/api-keys/rotate", middleware.AdminMiddleware, channelHandler.RotateAPIKey)

	channels.Post("/:id/models", middleware.AdminMiddleware, channelHandler.AddModel)
	channels.Get("/:id/models", middleware.AdminMiddleware, channelHandler.GetChannelModels)

	channels.Post("/:id/users/:userID/bind", middleware.AdminMiddleware, channelHandler.BindUserToChannel)
	channels.Delete("/:id/users/:userID/bind", middleware.AdminMiddleware, channelHandler.UnbindUserFromChannel)
}
