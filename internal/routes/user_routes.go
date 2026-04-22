package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/handlers"
	"keyraccoon/internal/middleware"
	"keyraccoon/internal/services"
)

func SetupUserRoutes(router fiber.Router, db *gorm.DB) {
	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler()

	auth := router.Group("/auth")
	auth.Post("/login", userHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/logout", middleware.AuthMiddleware, authHandler.Logout)

	users := router.Group("/users", middleware.AuthMiddleware)
	users.Post("", middleware.AdminMiddleware, userHandler.CreateUser)
	users.Get("", middleware.AdminMiddleware, userHandler.GetAllUsers)
	users.Get("/:id", userHandler.GetUser)
	users.Put("/:id", middleware.AdminMiddleware, userHandler.UpdateUser)
	users.Delete("/:id", middleware.AdminMiddleware, userHandler.DeleteUser)
	users.Get("/:id/usage", userHandler.GetUserUsage)
}
