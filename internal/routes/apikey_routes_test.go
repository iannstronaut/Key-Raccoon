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

func TestAPIKeyRoutesLifecycle(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)
	config.SetConfigForTesting(&config.Config{JWTSecret: "test-secret", JWTExpire: 60})

	db := openAPIKeyRoutesDB(t)
	adminToken, _ := seedAPIKeyRouteUsers(t, db)

	channelRepo := repositories.NewChannelRepository(db)
	channel := &models.Channel{Name: "Route Channel", Type: "openai", IsActive: true}
	if err := channelRepo.Create(channel); err != nil {
		t.Fatalf("channelRepo.Create() error = %v", err)
	}

	app := fiber.New()
	routes.SetupAPIKeyRoutes(app, db)

	createResp := mustAPIKeyJSONRequest(t, app, http.MethodPost, "/api-keys", map[string]any{
		"name":         "Route Key",
		"token_limit":  1000,
		"credit_limit": 100,
	}, adminToken)
	if createResp.StatusCode != fiber.StatusCreated {
		body := decodeAPIKeyResponse(t, createResp)
		t.Fatalf("create status = %d, want %d, body = %+v", createResp.StatusCode, fiber.StatusCreated, body)
	}
	createBody := decodeAPIKeyResponse(t, createResp)
	keyID := uint(createBody["id"].(float64))

	listResp := mustAPIKeyJSONRequest(t, app, http.MethodGet, "/api-keys", nil, adminToken)
	if listResp.StatusCode != fiber.StatusOK {
		t.Fatalf("list status = %d, want %d", listResp.StatusCode, fiber.StatusOK)
	}
	listBody := decodeAPIKeyResponse(t, listResp)
	if listBody["total"] != float64(1) {
		t.Fatalf("list total = %v, want 1", listBody["total"])
	}

	getResp := mustAPIKeyJSONRequest(t, app, http.MethodGet, "/api-keys/1", nil, adminToken)
	if getResp.StatusCode != fiber.StatusOK {
		t.Fatalf("get status = %d, want %d", getResp.StatusCode, fiber.StatusOK)
	}

	updateResp := mustAPIKeyJSONRequest(t, app, http.MethodPut, "/api-keys/1", map[string]any{
		"name":         "Updated Route Key",
		"token_limit":  2000,
		"credit_limit": 200,
	}, adminToken)
	if updateResp.StatusCode != fiber.StatusOK {
		t.Fatalf("update status = %d, want %d", updateResp.StatusCode, fiber.StatusOK)
	}
	updateBody := decodeAPIKeyResponse(t, updateResp)
	if updateBody["name"] != "Updated Route Key" {
		t.Fatalf("update name = %v, want Updated Route Key", updateBody["name"])
	}

	usageResp := mustAPIKeyJSONRequest(t, app, http.MethodGet, "/api-keys/1/usage", nil, adminToken)
	if usageResp.StatusCode != fiber.StatusOK {
		t.Fatalf("usage status = %d, want %d", usageResp.StatusCode, fiber.StatusOK)
	}

	bindResp := mustAPIKeyJSONRequest(t, app, http.MethodPost, "/api-keys/1/channels/1/bind", nil, adminToken)
	if bindResp.StatusCode != fiber.StatusOK {
		t.Fatalf("bind status = %d, want %d", bindResp.StatusCode, fiber.StatusOK)
	}

	unbindResp := mustAPIKeyJSONRequest(t, app, http.MethodDelete, "/api-keys/1/channels/1/bind", nil, adminToken)
	if unbindResp.StatusCode != fiber.StatusOK {
		t.Fatalf("unbind status = %d, want %d", unbindResp.StatusCode, fiber.StatusOK)
	}

	deleteResp := mustAPIKeyJSONRequest(t, app, http.MethodDelete, "/api-keys/1", nil, adminToken)
	if deleteResp.StatusCode != fiber.StatusOK {
		t.Fatalf("delete status = %d, want %d", deleteResp.StatusCode, fiber.StatusOK)
	}

	getDeletedResp := mustAPIKeyJSONRequest(t, app, http.MethodGet, "/api-keys/1", nil, adminToken)
	if getDeletedResp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("get deleted status = %d, want %d", getDeletedResp.StatusCode, fiber.StatusNotFound)
	}

	if keyID == 0 {
		t.Fatal("expected created api key id to be non-zero")
	}
}

func TestAPIKeyRoutesUnauthorized(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)
	config.SetConfigForTesting(&config.Config{JWTSecret: "test-secret", JWTExpire: 60})

	db := openAPIKeyRoutesDB(t)
	app := fiber.New()
	routes.SetupAPIKeyRoutes(app, db)

	resp := mustAPIKeyJSONRequest(t, app, http.MethodGet, "/api-keys", nil, "")
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("no-auth status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestAPIKeyRoutesOtherUserAccess(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)
	config.SetConfigForTesting(&config.Config{JWTSecret: "test-secret", JWTExpire: 60})

	db := openAPIKeyRoutesDB(t)
	adminToken, _ := seedAPIKeyRouteUsers(t, db)

	userService := services.NewUserService(repositories.NewUserRepository(db))
	otherUser, err := userService.CreateUser("other@example.com", "SecurePass123", "Other User", "user", 0, 0)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	otherToken, err := utils.GenerateAccessToken(otherUser.ID, otherUser.Email, otherUser.Role, "test-secret", 60)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	app := fiber.New()
	routes.SetupAPIKeyRoutes(app, db)

	createResp := mustAPIKeyJSONRequest(t, app, http.MethodPost, "/api-keys", map[string]any{
		"name": "Admin Key",
	}, adminToken)
	if createResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("admin create status = %d, want %d", createResp.StatusCode, fiber.StatusCreated)
	}

	otherListResp := mustAPIKeyJSONRequest(t, app, http.MethodGet, "/api-keys", nil, otherToken)
	if otherListResp.StatusCode != fiber.StatusOK {
		t.Fatalf("other list status = %d, want %d", otherListResp.StatusCode, fiber.StatusOK)
	}
	otherListBody := decodeAPIKeyResponse(t, otherListResp)
	if otherListBody["total"] != float64(0) {
		t.Fatalf("other list total = %v, want 0", otherListBody["total"])
	}
}

func openAPIKeyRoutesDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "apikey_routes.db")
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

func seedAPIKeyRouteUsers(t *testing.T, db *gorm.DB) (string, uint) {
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

func mustAPIKeyJSONRequest(t *testing.T, app *fiber.App, method, path string, body any, accessToken string) *http.Response {
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

func decodeAPIKeyResponse(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return body
}
