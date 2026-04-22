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

func TestProxyRoutesLifecycle(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)
	config.SetConfigForTesting(&config.Config{JWTSecret: "test-secret", JWTExpire: 60})

	db := openProxyRoutesDB(t)
	adminToken := seedProxyRouteAdmin(t, db)

	app := fiber.New()
	routes.SetupProxyRoutes(app, db)

	createResp := mustProxyJSONRequest(t, app, http.MethodPost, "/proxies", map[string]any{
		"proxy_url": "http://proxy.example.com:8080",
		"type":      "http",
		"username":  "user",
		"password":  "pass",
	}, adminToken)
	if createResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResp.StatusCode, fiber.StatusCreated)
	}
	createBody := decodeProxyResponse(t, createResp)
	proxyID := uint(createBody["id"].(float64))

	listResp := mustProxyJSONRequest(t, app, http.MethodGet, "/proxies", nil, adminToken)
	if listResp.StatusCode != fiber.StatusOK {
		t.Fatalf("list status = %d, want %d", listResp.StatusCode, fiber.StatusOK)
	}

	getResp := mustProxyJSONRequest(t, app, http.MethodGet, "/proxies/1", nil, adminToken)
	if getResp.StatusCode != fiber.StatusOK {
		t.Fatalf("get status = %d, want %d", getResp.StatusCode, fiber.StatusOK)
	}

	checkResp := mustProxyJSONRequest(t, app, http.MethodPost, "/proxies/1/check", nil, adminToken)
	if checkResp.StatusCode != fiber.StatusOK {
		t.Fatalf("check status = %d, want %d", checkResp.StatusCode, fiber.StatusOK)
	}

	deleteResp := mustProxyJSONRequest(t, app, http.MethodDelete, "/proxies/1", nil, adminToken)
	if deleteResp.StatusCode != fiber.StatusOK {
		t.Fatalf("delete status = %d, want %d", deleteResp.StatusCode, fiber.StatusOK)
	}

	if proxyID == 0 {
		t.Fatal("expected non-zero proxy id")
	}
}

func TestProxyRoutesPublicChecker(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)

	db := openProxyRoutesDB(t)
	app := fiber.New()
	routes.SetupProxyRoutes(app, db)

	resp := mustProxyJSONRequest(t, app, http.MethodPost, "/proxies/check", map[string]any{
		"proxy_url": "http://invalid-proxy.test:1",
		"type":      "http",
	}, "")
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("public check status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
	body := decodeProxyResponse(t, resp)
	if body["is_healthy"].(bool) {
		t.Fatal("expected is_healthy = false")
	}
}

func TestProxyRoutesRejectNonAdmin(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)
	config.SetConfigForTesting(&config.Config{JWTSecret: "test-secret", JWTExpire: 60})

	db := openProxyRoutesDB(t)
	userService := services.NewUserService(repositories.NewUserRepository(db))
	user, err := userService.CreateUser("user@example.com", "SecurePass123", "User", "user", 0, 0)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	token, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role, "test-secret", 60)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	app := fiber.New()
	routes.SetupProxyRoutes(app, db)

	resp := mustProxyJSONRequest(t, app, http.MethodPost, "/proxies", map[string]any{
		"proxy_url": "http://proxy.example.com:8080",
		"type":      "http",
	}, token)
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("non-admin create status = %d, want %d", resp.StatusCode, fiber.StatusForbidden)
	}
}

func openProxyRoutesDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "proxy_routes.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.User{}, &models.Proxy{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return db
}

func seedProxyRouteAdmin(t *testing.T, db *gorm.DB) string {
	t.Helper()

	userService := services.NewUserService(repositories.NewUserRepository(db))
	admin, err := userService.CreateUser("admin@example.com", "AdminPassword123", "Admin", "superadmin", -1, -1)
	if err != nil {
		t.Fatalf("CreateUser(admin) error = %v", err)
	}

	token, err := utils.GenerateAccessToken(admin.ID, admin.Email, admin.Role, "test-secret", 60)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	return token
}

func mustProxyJSONRequest(t *testing.T, app *fiber.App, method, path string, body any, accessToken string) *http.Response {
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

func decodeProxyResponse(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return body
}
