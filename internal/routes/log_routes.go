package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/handlers"
	"keyraccoon/internal/middleware"
	"keyraccoon/internal/services"
)

func SetupLogRoutes(router fiber.Router, db *gorm.DB, logService *services.LogService) {
	userAPIKeyRepo := repositories.NewUserAPIKeyRepository(db)
	userRepo := repositories.NewUserRepository(db)
	channelRepo := repositories.NewChannelRepository(db)
	modelRepo := repositories.NewModelRepository(db)

	userAPIKeyService := services.NewUserAPIKeyService(userAPIKeyRepo, userRepo, channelRepo, modelRepo)
	logHandler := handlers.NewLogHandler(logService, userAPIKeyService)

	logs := router.Group("/logs", middleware.AuthMiddleware)
	logs.Get("", middleware.AdminMiddleware, logHandler.GetLogs)
	logs.Get("/stats", logHandler.GetStats) // all users; handler scopes non-admin to own data
	logs.Get("/user/:userID", logHandler.GetUserLogs)
	logs.Get("/api-key/:keyID", logHandler.GetAPIKeyLogs)
}
