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

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
	"keyraccoon/internal/routes"
	"keyraccoon/internal/services"
)

func TestAPIV1RoutesUnauthorized(t *testing.T) {
	db := openAPIV1RoutesDB(t)
	app := fiber.New()
	routes.SetupAPIV1Routes(app, db)

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("no auth status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestAPIV1RoutesInvalidKey(t *testing.T) {
	db := openAPIV1RoutesDB(t)
	app := fiber.New()
	routes.SetupAPIV1Routes(app, db)

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set("Authorization", "Bearer invalid-key")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("invalid key status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestAPIV1RoutesListModels(t *testing.T) {
	db, apiKey := seedAPIV1RoutesData(t)
	app := fiber.New()
	routes.SetupAPIV1Routes(app, db)

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey.Key)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("list models status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	data, ok := body["data"].([]interface{})
	if !ok || len(data) != 1 {
		t.Fatalf("models data length = %d, want 1", len(data))
	}
}

func TestAPIV1RoutesEmbeddings(t *testing.T) {
	db, apiKey := seedAPIV1RoutesData(t)
	app := fiber.New()
	routes.SetupAPIV1Routes(app, db)

	req := httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey.Key)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("embeddings status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
}

func TestAPIV1RoutesChatCompletionValidation(t *testing.T) {
	db, apiKey := seedAPIV1RoutesData(t)
	app := fiber.New()
	routes.SetupAPIV1Routes(app, db)

	// Missing model
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(`{"messages": [{"role": "user", "content": "Hello"}]}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey.Key)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("missing model status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}

	// Missing messages
	req = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte(`{"model": "gpt-4"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey.Key)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("missing messages status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func openAPIV1RoutesDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "api_v1_routes.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.User{}, &models.Channel{}, &models.APIKey{}, &models.ChannelAPIKey{}, &models.Model{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return db
}

func seedAPIV1RoutesData(t *testing.T) (*gorm.DB, *models.APIKey) {
	t.Helper()

	db := openAPIV1RoutesDB(t)

	userRepo := repositories.NewUserRepository(db)
	user := &models.User{
		Email:    "apiv1user@example.com",
		Password: "hashed",
		Name:     "API V1 User",
		Role:     "user",
		IsActive: true,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("userRepo.Create() error = %v", err)
	}

	channelRepo := repositories.NewChannelRepository(db)
	channel := &models.Channel{
		Name:     "API V1 Channel",
		Type:     "openai",
		IsActive: true,
	}
	if err := channelRepo.Create(channel); err != nil {
		t.Fatalf("channelRepo.Create() error = %v", err)
	}

	apiKeyRepo := repositories.NewAPIKeyRepository(db)
	apiKeyService := services.NewAPIKeyService(apiKeyRepo, userRepo)
	apiKey, err := apiKeyService.CreateAPIKey(user.ID, "API V1 Key", 1000, 100)
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	if err := apiKeyService.BindChannel(apiKey.ID, channel.ID); err != nil {
		t.Fatalf("BindChannel() error = %v", err)
	}

	channelAPIKeyRepo := repositories.NewChannelAPIKeyRepository(db)
	modelRepo := repositories.NewModelRepository(db)
	channelService := services.NewChannelService(channelRepo, channelAPIKeyRepo, modelRepo, userRepo)

	if _, err := channelService.AddAPIKey(channel.ID); err != nil {
		t.Fatalf("AddAPIKey() error = %v", err)
	}

	if _, err := channelService.AddModel(channel.ID, "gpt-4", "GPT-4", 0.03, ""); err != nil {
		t.Fatalf("AddModel() error = %v", err)
	}

	return db, apiKey
}
