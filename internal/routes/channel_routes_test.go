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

func TestChannelRoutesLifecycle(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)
	config.SetConfigForTesting(&config.Config{JWTSecret: "test-secret", JWTExpire: 60})

	db := openChannelRoutesDB(t)
	adminToken, userID := seedChannelRouteUsers(t, db)

	app := fiber.New()
	routes.SetupChannelRoutes(app, db)

	createResp := mustChannelJSONRequest(t, app, http.MethodPost, "/channels", map[string]any{
		"name":        "OpenAI Production",
		"type":        "openai",
		"description": "Main channel",
	}, adminToken)
	if createResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResp.StatusCode, fiber.StatusCreated)
	}
	channelBody := decodeChannelResponse(t, createResp)
	channelID := uint(channelBody["id"].(float64))

	listResp := mustChannelJSONRequest(t, app, http.MethodGet, "/channels", nil, adminToken)
	if listResp.StatusCode != fiber.StatusOK {
		t.Fatalf("list status = %d, want %d", listResp.StatusCode, fiber.StatusOK)
	}

	getResp := mustChannelJSONRequest(t, app, http.MethodGet, "/channels/1", nil, adminToken)
	if getResp.StatusCode != fiber.StatusOK {
		t.Fatalf("get status = %d, want %d", getResp.StatusCode, fiber.StatusOK)
	}

	updateResp := mustChannelJSONRequest(t, app, http.MethodPut, "/channels/1", map[string]any{
		"description": "Updated channel",
		"is_active":   true,
	}, adminToken)
	if updateResp.StatusCode != fiber.StatusOK {
		t.Fatalf("update status = %d, want %d", updateResp.StatusCode, fiber.StatusOK)
	}

	addKeyResp := mustChannelJSONRequest(t, app, http.MethodPost, "/channels/1/api-keys", nil, adminToken)
	if addKeyResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("add api key status = %d, want %d", addKeyResp.StatusCode, fiber.StatusCreated)
	}

	keysResp := mustChannelJSONRequest(t, app, http.MethodGet, "/channels/1/api-keys", nil, adminToken)
	if keysResp.StatusCode != fiber.StatusOK {
		t.Fatalf("list api keys status = %d, want %d", keysResp.StatusCode, fiber.StatusOK)
	}

	rotateResp := mustChannelJSONRequest(t, app, http.MethodPost, "/channels/1/api-keys/rotate", nil, adminToken)
	if rotateResp.StatusCode != fiber.StatusOK {
		t.Fatalf("rotate api key status = %d, want %d", rotateResp.StatusCode, fiber.StatusOK)
	}

	addModelResp := mustChannelJSONRequest(t, app, http.MethodPost, "/channels/1/models", map[string]any{
		"name":          "gpt-4",
		"display_name":  "GPT-4",
		"token_price":   0.03,
		"system_prompt": "You are helpful",
	}, adminToken)
	if addModelResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("add model status = %d, want %d", addModelResp.StatusCode, fiber.StatusCreated)
	}

	modelsResp := mustChannelJSONRequest(t, app, http.MethodGet, "/channels/1/models", nil, adminToken)
	if modelsResp.StatusCode != fiber.StatusOK {
		t.Fatalf("list models status = %d, want %d", modelsResp.StatusCode, fiber.StatusOK)
	}

	bindResp := mustChannelJSONRequest(t, app, http.MethodPost, "/channels/1/users/2/bind", nil, adminToken)
	if bindResp.StatusCode != fiber.StatusOK {
		t.Fatalf("bind status = %d, want %d", bindResp.StatusCode, fiber.StatusOK)
	}

	unbindResp := mustChannelJSONRequest(t, app, http.MethodDelete, "/channels/1/users/2/bind", nil, adminToken)
	if unbindResp.StatusCode != fiber.StatusOK {
		t.Fatalf("unbind status = %d, want %d", unbindResp.StatusCode, fiber.StatusOK)
	}

	deleteResp := mustChannelJSONRequest(t, app, http.MethodDelete, "/channels/1", nil, adminToken)
	if deleteResp.StatusCode != fiber.StatusOK {
		t.Fatalf("delete status = %d, want %d", deleteResp.StatusCode, fiber.StatusOK)
	}

	if channelID == 0 || userID == 0 {
		t.Fatal("expected non-zero ids from seeded route flow")
	}
}

func TestChannelRoutesRejectNonAdmin(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)
	config.SetConfigForTesting(&config.Config{JWTSecret: "test-secret", JWTExpire: 60})

	db := openChannelRoutesDB(t)
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
	routes.SetupChannelRoutes(app, db)

	resp := mustChannelJSONRequest(t, app, http.MethodPost, "/channels", map[string]any{
		"name": "Blocked",
		"type": "openai",
	}, token)
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("non-admin create status = %d, want %d", resp.StatusCode, fiber.StatusForbidden)
	}
}

func openChannelRoutesDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "channel_routes.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.User{}, &models.Channel{}, &models.ChannelAPIKey{}, &models.Model{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return db
}

func seedChannelRouteUsers(t *testing.T, db *gorm.DB) (string, uint) {
	t.Helper()

	userService := services.NewUserService(repositories.NewUserRepository(db))
	admin, err := userService.CreateUser("admin@example.com", "AdminPassword123", "Admin", "superadmin", -1, -1)
	if err != nil {
		t.Fatalf("CreateUser(admin) error = %v", err)
	}
	user, err := userService.CreateUser("user@example.com", "SecurePass123", "User", "user", 0, 0)
	if err != nil {
		t.Fatalf("CreateUser(user) error = %v", err)
	}

	token, err := utils.GenerateAccessToken(admin.ID, admin.Email, admin.Role, "test-secret", 60)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	return token, user.ID
}

func mustChannelJSONRequest(t *testing.T, app *fiber.App, method, path string, body any, accessToken string) *http.Response {
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

func decodeChannelResponse(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return body
}
