package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/handlers"
	"keyraccoon/internal/middleware"
	"keyraccoon/internal/services"
)

func SetupProxyRoutes(app *fiber.App, db *gorm.DB) {
	proxyRepo := repositories.NewProxyRepository(db)
	proxyService := services.NewProxyService(proxyRepo)
	proxyHandler := handlers.NewProxyHandler(proxyService)

	app.Post("/proxies/check", proxyHandler.TestProxy)

	proxies := app.Group("/proxies", middleware.AuthMiddleware, middleware.AdminMiddleware)

	proxies.Post("", proxyHandler.AddProxy)
	proxies.Get("", proxyHandler.GetAllProxies)
	proxies.Get("/:id", proxyHandler.GetProxy)
	proxies.Delete("/:id", proxyHandler.DeleteProxy)
	proxies.Post("/:id/check", proxyHandler.CheckProxyHealth)
}
