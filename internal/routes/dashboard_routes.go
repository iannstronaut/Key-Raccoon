package routes

import (
	"github.com/gofiber/fiber/v2"
)

func SetupDashboardRoutes(app *fiber.App) {
	// Serve login page at root /login
	app.Static("/login", "./public/login.html")
	app.Get("/login", func(c *fiber.Ctx) error {
		return c.SendFile("./public/login.html")
	})

	// Serve dashboard static assets
	app.Static("/dashboard", "./public/dashboard")

	// Dashboard main page
	app.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.SendFile("./public/dashboard/index.html")
	})

	// Catch-all for SPA deep links
	app.Get("/dashboard/*", func(c *fiber.Ctx) error {
		return c.SendFile("./public/dashboard/index.html")
	})
}
