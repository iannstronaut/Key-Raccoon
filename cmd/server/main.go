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

	app.Use(cors.New())
	app.Use(appmiddleware.RequestLogger())
	handlers.RegisterHealthRoutes(app)
	routes.SetupUserRoutes(app, config.GetDB())
	routes.SetupChannelRoutes(app, config.GetDB())
	routes.SetupAPIKeyRoutes(app, config.GetDB())
	routes.SetupProxyRoutes(app, config.GetDB())
	routes.SetupAPIV1Routes(app, config.GetDB())

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
