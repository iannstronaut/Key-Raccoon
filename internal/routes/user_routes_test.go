package routes_test

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
	"keyraccoon/internal/models"
	"keyraccoon/internal/routes"
	"keyraccoon/internal/services"
	"keyraccoon/internal/utils"
)

func TestUserRoutesLoginRefreshCRUDAndLogout(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)
	config.SetConfigForTesting(&config.Config{JWTSecret: "test-secret", JWTExpire: 60})

	db := openRoutesDB(t)
	seedRouteUsers(t, db)

	app := fiber.New()
	routes.SetupUserRoutes(app, db)

	loginResp := mustJSONRequest(t, app, http.MethodPost, "/auth/login", map[string]any{
		"email":    "admin@example.com",
		"password": "AdminPassword123",
	}, "")
	if loginResp.StatusCode != fiber.StatusOK {
		t.Fatalf("login status = %d, want %d", loginResp.StatusCode, fiber.StatusOK)
	}
	loginBody := decodeResponse(t, loginResp)
	accessToken := loginBody["access_token"].(string)
	refreshToken := loginBody["refresh_token"].(string)

	createResp := mustJSONRequest(t, app, http.MethodPost, "/users", map[string]any{
		"email":        "user@example.com",
		"password":     "SecurePass123",
		"name":         "John Doe",
		"role":         "user",
		"token_limit":  700,
		"credit_limit": 15.5,
	}, accessToken)
	if createResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResp.StatusCode, fiber.StatusCreated)
	}
	createdBody := decodeResponse(t, createResp)
	if _, hasPassword := createdBody["password"]; hasPassword {
		t.Fatal("create response should not include password")
	}
	userID := uint(createdBody["id"].(float64))

	listResp := mustJSONRequest(t, app, http.MethodGet, "/users?limit=20&offset=0", nil, accessToken)
	if listResp.StatusCode != fiber.StatusOK {
		t.Fatalf("list status = %d, want %d", listResp.StatusCode, fiber.StatusOK)
	}
	listBody := decodeResponse(t, listResp)
	if listBody["total"].(float64) < 2 {
		t.Fatalf("list total = %v, want at least 2", listBody["total"])
	}

	getResp := mustJSONRequest(t, app, http.MethodGet, "/users/2", nil, accessToken)
	if getResp.StatusCode != fiber.StatusOK {
		t.Fatalf("get status = %d, want %d", getResp.StatusCode, fiber.StatusOK)
	}

	updateResp := mustJSONRequest(t, app, http.MethodPut, "/users/2", map[string]any{
		"name":        "Jane Updated",
		"is_active":   true,
		"token_limit": 800,
		"role":        "admin",
	}, accessToken)
	if updateResp.StatusCode != fiber.StatusOK {
		t.Fatalf("update status = %d, want %d", updateResp.StatusCode, fiber.StatusOK)
	}
	updateBody := decodeResponse(t, updateResp)
	if updateBody["name"] != "Jane Updated" || updateBody["role"] != "admin" {
		t.Fatalf("unexpected update body: %+v", updateBody)
	}

	usageResp := mustJSONRequest(t, app, http.MethodGet, "/users/2/usage", nil, accessToken)
	if usageResp.StatusCode != fiber.StatusOK {
		t.Fatalf("usage status = %d, want %d", usageResp.StatusCode, fiber.StatusOK)
	}

	refreshResp := mustJSONRequest(t, app, http.MethodPost, "/auth/refresh", map[string]any{
		"refresh_token": refreshToken,
	}, "")
	if refreshResp.StatusCode != fiber.StatusOK {
		t.Fatalf("refresh status = %d, want %d", refreshResp.StatusCode, fiber.StatusOK)
	}

	logoutResp := mustJSONRequest(t, app, http.MethodPost, "/auth/logout", nil, accessToken)
	if logoutResp.StatusCode != fiber.StatusOK {
		t.Fatalf("logout status = %d, want %d", logoutResp.StatusCode, fiber.StatusOK)
	}

	deleteResp := mustJSONRequest(t, app, http.MethodDelete, "/users/2", nil, accessToken)
	if deleteResp.StatusCode != fiber.StatusOK {
		t.Fatalf("delete status = %d, want %d", deleteResp.StatusCode, fiber.StatusOK)
	}

	getDeletedResp := mustJSONRequest(t, app, http.MethodGet, "/users/2", nil, accessToken)
	if getDeletedResp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("get deleted status = %d, want %d", getDeletedResp.StatusCode, fiber.StatusNotFound)
	}

	if userID == 0 {
		t.Fatal("expected created user id to be non-zero")
	}
}

func TestUserRoutesRejectNonAdminUserCreation(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)
	config.SetConfigForTesting(&config.Config{JWTSecret: "test-secret", JWTExpire: 60})

	db := openRoutesDB(t)
	app := fiber.New()
	routes.SetupUserRoutes(app, db)

	userService := services.NewUserService(repositories.NewUserRepository(db))
	normalUser, err := userService.CreateUser("plain@example.com", "SecurePass123", "Plain User", "user", 0, 0)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	token, err := utils.GenerateAccessToken(normalUser.ID, normalUser.Email, normalUser.Role, "test-secret", 60)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	resp := mustJSONRequest(t, app, http.MethodPost, "/users", map[string]any{
		"email":    "new@example.com",
		"password": "SecurePass123",
		"name":     "New User",
		"role":     "user",
	}, token)
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("non-admin create status = %d, want %d", resp.StatusCode, fiber.StatusForbidden)
	}
}

func openRoutesDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "user_routes.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return db
}

func seedRouteUsers(t *testing.T, db *gorm.DB) {
	t.Helper()

	userService := services.NewUserService(repositories.NewUserRepository(db))
	if _, err := userService.CreateUser("admin@example.com", "AdminPassword123", "Admin", "superadmin", -1, -1); err != nil {
		t.Fatalf("CreateUser(admin) error = %v", err)
	}
	if _, err := userService.CreateUser("existing@example.com", "SecurePass123", "Existing User", "user", 500, 20); err != nil {
		t.Fatalf("CreateUser(existing) error = %v", err)
	}
}

func mustJSONRequest(t *testing.T, app *fiber.App, method, path string, body any, accessToken string) *http.Response {
	t.Helper()

	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	return resp
}

func decodeResponse(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return body
}
