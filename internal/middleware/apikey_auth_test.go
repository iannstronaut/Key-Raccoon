package middleware_test

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/middleware"
	"keyraccoon/internal/models"
	"keyraccoon/internal/services"
)

func TestAPIKeyAuthMiddlewareMissingHeader(t *testing.T) {
	db := openAPIKeyAuthDB(t)
	apiKeyService := setupAPIKeyAuthService(db)

	app := fiber.New()
	app.Get("/test", middleware.APIKeyAuthMiddleware(apiKeyService), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	resp, _ := app.Test(httptest.NewRequest("GET", "/test", nil))
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("missing header status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestAPIKeyAuthMiddlewareInvalidFormat(t *testing.T) {
	db := openAPIKeyAuthDB(t)
	apiKeyService := setupAPIKeyAuthService(db)

	app := fiber.New()
	app.Get("/test", middleware.APIKeyAuthMiddleware(apiKeyService), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Token abc")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("invalid format status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestAPIKeyAuthMiddlewareInvalidKey(t *testing.T) {
	db := openAPIKeyAuthDB(t)
	apiKeyService := setupAPIKeyAuthService(db)

	app := fiber.New()
	app.Get("/test", middleware.APIKeyAuthMiddleware(apiKeyService), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-key")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("invalid key status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestAPIKeyAuthMiddlewareSuccess(t *testing.T) {
	db := openAPIKeyAuthDB(t)
	apiKeyService := setupAPIKeyAuthService(db)

	user := &models.User{
		Email:    "apikeyuser@example.com",
		Password: "hashed",
		Name:     "API Key User",
		Role:     "user",
		IsActive: true,
	}
	userRepo := repositories.NewUserRepository(db)
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("userRepo.Create() error = %v", err)
	}

	apiKey, err := apiKeyService.CreateAPIKey(user.ID, "Test Key", 1000, 100)
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	app := fiber.New()
	app.Get("/test", middleware.APIKeyAuthMiddleware(apiKeyService), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"api_key_id": c.Locals("api_key_id"),
			"user_id":    c.Locals("user_id"),
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey.Key)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("valid key status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body["api_key_id"] != float64(apiKey.ID) {
		t.Fatalf("api_key_id = %v, want %d", body["api_key_id"], apiKey.ID)
	}
	if body["user_id"] != float64(user.ID) {
		t.Fatalf("user_id = %v, want %d", body["user_id"], user.ID)
	}
}

func openAPIKeyAuthDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "apikey_auth.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.User{}, &models.Channel{}, &models.APIKey{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return db
}

func setupAPIKeyAuthService(db *gorm.DB) *services.APIKeyService {
	apiKeyRepo := repositories.NewAPIKeyRepository(db)
	userRepo := repositories.NewUserRepository(db)
	return services.NewAPIKeyService(apiKeyRepo, userRepo)
}
