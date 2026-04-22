package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"keyraccoon/internal/config"
	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/handlers"
	"keyraccoon/internal/middleware"
	"keyraccoon/internal/models"
	"keyraccoon/internal/services"
	"keyraccoon/internal/utils"
)

func setupUserHandlerTest(t *testing.T) (*fiber.App, *services.UserService, func()) {
	t.Helper()

	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)
	config.SetConfigForTesting(&config.Config{
		JWTSecret: "test-secret-key-for-unit-tests-only",
	})

	dbPath := filepath.Join(t.TempDir(), "user_handler_test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler()

	app := fiber.New(fiber.Config{
		ErrorHandler: handlers.ErrorHandler,
	})

	auth := app.Group("/auth")
	auth.Post("/login", userHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)

	users := app.Group("/users", middleware.AuthMiddleware)
	users.Post("", middleware.AdminMiddleware, userHandler.CreateUser)
	users.Get("", middleware.AdminMiddleware, userHandler.GetAllUsers)
	users.Get("/:id", userHandler.GetUser)
	users.Put("/:id", middleware.AdminMiddleware, userHandler.UpdateUser)
	users.Delete("/:id", middleware.AdminMiddleware, userHandler.DeleteUser)

	cleanup := func() {
		_ = sqlDB.Close()
	}

	return app, userService, cleanup
}

func TestLoginHandler(t *testing.T) {
	app, userService, cleanup := setupUserHandlerTest(t)
	defer cleanup()

	_, err := userService.CreateUser("admin@test.com", "password123", "Admin User", "admin", 0, 0)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	body := []byte(`{"email":"admin@test.com","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if result["access_token"] == nil || result["access_token"] == "" {
		t.Fatal("access_token is empty")
	}
}

func TestLoginHandlerInvalidCredentials(t *testing.T) {
	app, _, cleanup := setupUserHandlerTest(t)
	defer cleanup()

	body := []byte(`{"email":"nobody@test.com","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestGetUsersUnauthorized(t *testing.T) {
	app, _, cleanup := setupUserHandlerTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestGetUsersAsAdmin(t *testing.T) {
	app, userService, cleanup := setupUserHandlerTest(t)
	defer cleanup()

	user, err := userService.CreateUser("admin2@test.com", "password123", "Admin", "admin", 0, 0)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	token, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role, "test-secret-key-for-unit-tests-only", 60)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestCreateUserAsAdmin(t *testing.T) {
	app, userService, cleanup := setupUserHandlerTest(t)
	defer cleanup()

	admin, err := userService.CreateUser("admin3@test.com", "password123", "Admin", "admin", 0, 0)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	token, err := utils.GenerateAccessToken(admin.ID, admin.Email, admin.Role, "test-secret-key-for-unit-tests-only", 60)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	body := []byte(`{"email":"newuser@test.com","password":"password123","name":"New User","role":"user"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}
