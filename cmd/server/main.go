package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"keyraccoon/internal/config"
	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/handlers"
	appmiddleware "keyraccoon/internal/middleware"
	"keyraccoon/internal/routes"
	"keyraccoon/internal/services"
	"keyraccoon/pkg/logger"
)

func main() {
	logger.Init()

	cfg, err := config.Init()
	if err != nil {
		logger.Fatal("failed to load config", "error", err)
	}

	if err := config.InitDatabase(cfg); err != nil {
		logger.Fatal("failed to initialize database", "error", err)
	}

	if err := config.InitRedis(cfg); err != nil {
		logger.Fatal("failed to initialize redis", "error", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      "KeyRaccoon",
		ServerHeader: "KeyRaccoon",
		ErrorHandler: handlers.ErrorHandler,
	})

	corsConfig := cors.Config{
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		ExposeHeaders:    "Content-Length,Content-Type",
		AllowCredentials: true,
		MaxAge:           86400,
	}

	if cfg.CORSOrigin == "*" {
		corsConfig.AllowOriginsFunc = func(origin string) bool {
			return true
		}
	} else {
		corsConfig.AllowOrigins = cfg.CORSOrigin
	}

	app.Use(cors.New(corsConfig))
	appmiddleware.SecurityMiddleware(app)
	app.Use(appmiddleware.RequestLogger())
	handlers.RegisterHealthRoutes(app)

	// API routes under /api prefix
	api := app.Group("/api")
	routes.SetupUserRoutes(api, config.GetDB())
	routes.SetupChannelRoutes(api, config.GetDB())
	routes.SetupAPIKeyRoutes(api, config.GetDB())
	routes.SetupProxyRoutes(api, config.GetDB())
	routes.SetupAPIV1Routes(api, config.GetDB())

	routes.SetupDashboardRoutes(app)

	proxyRepo := repositories.NewProxyRepository(config.GetDB())
	proxyService := services.NewProxyService(proxyRepo)
	scheduler := services.NewSchedulerService(proxyService)
	scheduler.StartInBackground()

	addr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	logger.Info("server starting", "address", addr, "environment", cfg.Environment)

	if err := app.Listen(addr); err != nil {
		logger.Fatal("failed to start server", "error", err)
		os.Exit(1)
	}
}
