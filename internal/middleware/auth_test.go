package middleware_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"keyraccoon/internal/config"
	"keyraccoon/internal/middleware"
	"keyraccoon/internal/utils"
)

func TestAuthMiddlewareAndRBAC(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)
	config.SetConfigForTesting(&config.Config{JWTSecret: "test-secret", JWTExpire: 60})

	adminToken, err := utils.GenerateAccessToken(12, "admin@example.com", "admin", "test-secret", 15)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	userToken, err := utils.GenerateAccessToken(9, "user@example.com", "user", "test-secret", 15)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	app := fiber.New()
	app.Get("/admin", middleware.AuthMiddleware, middleware.AdminMiddleware, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id": c.Locals("user_id"),
			"role":    c.Locals("user_role"),
		})
	})
	app.Get("/super", middleware.AuthMiddleware, middleware.SuperAdminMiddleware, func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/role", middleware.AuthMiddleware, middleware.RoleMiddleware("admin", "superadmin"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	resp, _ := app.Test(httptest.NewRequest("GET", "/admin", nil))
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("missing auth status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}

	badHeaderReq := httptest.NewRequest("GET", "/admin", nil)
	badHeaderReq.Header.Set("Authorization", "Token abc")
	resp, _ = app.Test(badHeaderReq)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("bad header status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}

	adminReq := httptest.NewRequest("GET", "/admin", nil)
	adminReq.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err = app.Test(adminReq)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK || body["role"] != "admin" {
		t.Fatalf("admin response unexpected: status=%d body=%v", resp.StatusCode, body)
	}

	userReq := httptest.NewRequest("GET", "/admin", nil)
	userReq.Header.Set("Authorization", "Bearer "+userToken)
	resp, _ = app.Test(userReq)
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("user admin status = %d, want %d", resp.StatusCode, fiber.StatusForbidden)
	}

	superReq := httptest.NewRequest("GET", "/super", nil)
	superReq.Header.Set("Authorization", "Bearer "+adminToken)
	resp, _ = app.Test(superReq)
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("non-superadmin status = %d, want %d", resp.StatusCode, fiber.StatusForbidden)
	}

	roleReq := httptest.NewRequest("GET", "/role", nil)
	roleReq.Header.Set("Authorization", "Bearer "+adminToken)
	resp, _ = app.Test(roleReq)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("role route status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
}
