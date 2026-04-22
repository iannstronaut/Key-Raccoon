package routes

import (
	"github.com/gofiber/fiber/v2"
)

func SetupDashboardRoutes(router fiber.Router) {
	// Serve login page at root /login
	router.Static("/login", "./public/login.html")
	router.Get("/login", func(c *fiber.Ctx) error {
		return c.SendFile("./public/login.html")
	})

	// Serve dashboard static assets
	router.Static("/dashboard", "./public/dashboard")

	// Dashboard main page
	router.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.SendFile("./public/dashboard/index.html")
	})

	// Catch-all for SPA deep links
	router.Get("/dashboard/*", func(c *fiber.Ctx) error {
		return c.SendFile("./public/dashboard/index.html")
	})
}
