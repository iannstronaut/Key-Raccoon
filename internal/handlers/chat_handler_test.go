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

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/handlers"
	"keyraccoon/internal/models"
	"keyraccoon/internal/services"
)

func TestChatHandlerMissingModel(t *testing.T) {
	handler, apiKey := setupChatHandler(t)

	app := fiber.New()
	app.Post("/chat/completions", func(c *fiber.Ctx) error {
		c.Locals("api_key_id", apiKey.ID)
		c.Locals("api_key_channels", []models.Channel{})
		return handler.ChatCompletion(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/chat/completions", bytes.NewReader([]byte(`{
		"messages": [{"role": "user", "content": "Hello"}]
	}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("missing model status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestChatHandlerMissingMessages(t *testing.T) {
	handler, apiKey := setupChatHandler(t)

	app := fiber.New()
	app.Post("/chat/completions", func(c *fiber.Ctx) error {
		c.Locals("api_key_id", apiKey.ID)
		c.Locals("api_key_channels", []models.Channel{})
		return handler.ChatCompletion(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/chat/completions", bytes.NewReader([]byte(`{
		"model": "gpt-4"
	}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("missing messages status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestChatHandlerNoChannels(t *testing.T) {
	handler, apiKey := setupChatHandler(t)

	app := fiber.New()
	app.Post("/chat/completions", func(c *fiber.Ctx) error {
		c.Locals("api_key_id", apiKey.ID)
		c.Locals("api_key_channels", []models.Channel{})
		return handler.ChatCompletion(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/chat/completions", bytes.NewReader([]byte(`{
		"model": "gpt-4",
		"messages": [{"role": "user", "content": "Hello"}]
	}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey.Key)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("no channels status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestChatHandlerTokenLimitExceeded(t *testing.T) {
	handler, apiKey, channel := setupChatHandlerWithChannel(t)

	app := fiber.New()
	app.Post("/chat/completions", func(c *fiber.Ctx) error {
		c.Locals("api_key_id", apiKey.ID)
		c.Locals("api_key_channels", []models.Channel{*channel})
		return handler.ChatCompletion(c)
	})

	// Create a large message to exceed the 10 token limit
	largeContent := make([]byte, 1000)
	for i := range largeContent {
		largeContent[i] = 'a'
	}
	body := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []map[string]string{{"role": "user", "content": string(largeContent)}},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey.Key)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("token limit status = %d, want %d", resp.StatusCode, fiber.StatusTooManyRequests)
	}
}

func TestChatHandlerSuccessfulCompletion(t *testing.T) {
	handler, apiKey, channel := setupChatHandlerWithChannel(t)

	// Mock upstream server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"id":      "chatcmpl-test",
			"object":  "chat.completion",
			"created": 1234567890,
			"model":   "gpt-4",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]string{
						"role":    "assistant",
						"content": "Hello! How can I help you?",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 15,
				"total_tokens":      25,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	handler.SetOpenAIBaseURL(mockServer.URL)

	app := fiber.New()
	app.Post("/chat/completions", func(c *fiber.Ctx) error {
		c.Locals("api_key_id", apiKey.ID)
		c.Locals("api_key_channels", []models.Channel{*channel})
		return handler.ChatCompletion(c)
	})

	body := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []map[string]string{{"role": "user", "content": "Hello"}},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey.Key)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("chat completion status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if respBody["id"] != "chatcmpl-test" {
		t.Fatalf("id = %v, want chatcmpl-test", respBody["id"])
	}
}

func TestChatHandlerListModels(t *testing.T) {
	handler, apiKey, channel := setupChatHandlerWithChannelAndModel(t)

	app := fiber.New()
	app.Get("/models", func(c *fiber.Ctx) error {
		c.Locals("api_key_id", apiKey.ID)
		c.Locals("api_key_channels", []models.Channel{*channel})
		return handler.ListModels(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/models", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey.Key)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("list models status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	data, ok := respBody["data"].([]interface{})
	if !ok || len(data) != 1 {
		t.Fatalf("models data length = %d, want 1", len(data))
	}
}

func TestChatHandlerEmbeddings(t *testing.T) {
	handler, _ := setupChatHandler(t)

	app := fiber.New()
	app.Post("/embeddings", handler.Embeddings)

	req := httptest.NewRequest(http.MethodPost, "/embeddings", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("embeddings status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
}

func setupChatHandler(t *testing.T) (*handlers.ChatHandler, *models.UserAPIKey) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "chat_handler.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.User{}, &models.Channel{}, &models.UserAPIKey{}, &models.UserAPIKeyModel{}, &models.ChannelAPIKey{}, &models.Model{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	user := &models.User{
		Email:    "chatuser@example.com",
		Password: "hashed",
		Name:     "Chat User",
		Role:     "user",
		IsActive: true,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("userRepo.Create() error = %v", err)
	}

	userAPIKeyRepo := repositories.NewUserAPIKeyRepository(db)
	channelRepo := repositories.NewChannelRepository(db)
	modelRepo := repositories.NewModelRepository(db)
	userAPIKeyService := services.NewUserAPIKeyService(userAPIKeyRepo, userRepo, channelRepo, modelRepo)
	apiKey, err := userAPIKeyService.CreateAPIKey(user.ID, "Chat Key", 10, nil, []uint{}, []uint{})
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	channelAPIKeyRepo := repositories.NewChannelAPIKeyRepository(db)
	channelService := services.NewChannelService(channelRepo, channelAPIKeyRepo, modelRepo, userRepo)
	proxyService := services.NewProxyService(repositories.NewProxyRepository(db))

	handler := handlers.NewChatHandler(userAPIKeyService, channelService, proxyService, nil)
	return handler, apiKey
}

func setupChatHandlerWithChannel(t *testing.T) (*handlers.ChatHandler, *models.UserAPIKey, *models.Channel) {
	t.Helper()

	handler, apiKey := setupChatHandler(t)

	// We need to access the services from the handler, but they're private.
	// Instead, recreate them here.
	dbPath := filepath.Join(t.TempDir(), "chat_handler_channel.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.User{}, &models.Channel{}, &models.UserAPIKey{}, &models.UserAPIKeyModel{}, &models.ChannelAPIKey{}, &models.Model{}, &models.Proxy{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	user := &models.User{
		Email:    "chatuser2@example.com",
		Password: "hashed",
		Name:     "Chat User 2",
		Role:     "user",
		IsActive: true,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("userRepo.Create() error = %v", err)
	}

	channelRepo := repositories.NewChannelRepository(db)
	channel := &models.Channel{
		Name:     "Test Channel",
		Type:     "openai",
		IsActive: true,
	}
	if err := channelRepo.Create(channel); err != nil {
		t.Fatalf("channelRepo.Create() error = %v", err)
	}

	userAPIKeyRepo := repositories.NewUserAPIKeyRepository(db)
	modelRepo := repositories.NewModelRepository(db)
	userAPIKeyService := services.NewUserAPIKeyService(userAPIKeyRepo, userRepo, channelRepo, modelRepo)
	apiKey, err = userAPIKeyService.CreateAPIKey(user.ID, "Chat Key 2", 0, nil, []uint{channel.ID}, []uint{})
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	// Set token limit to 10 for testing
	apiKey.TokenLimit = 10
	if err := userAPIKeyRepo.Update(apiKey); err != nil {
		t.Fatalf("Update TokenLimit error = %v", err)
	}

	channelAPIKeyRepo := repositories.NewChannelAPIKeyRepository(db)
	channelService := services.NewChannelService(channelRepo, channelAPIKeyRepo, modelRepo, userRepo)
	proxyService := services.NewProxyService(repositories.NewProxyRepository(db))

	// Add a channel API key for forwarding
	if _, err := channelService.AddAPIKey(channel.ID); err != nil {
		t.Fatalf("AddAPIKey() error = %v", err)
	}

	handler = handlers.NewChatHandler(userAPIKeyService, channelService, proxyService, nil)
	return handler, apiKey, channel
}

func setupChatHandlerWithChannelAndModel(t *testing.T) (*handlers.ChatHandler, *models.UserAPIKey, *models.Channel) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "chat_handler_model.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.User{}, &models.Channel{}, &models.UserAPIKey{}, &models.UserAPIKeyModel{}, &models.ChannelAPIKey{}, &models.Model{}, &models.Proxy{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	user := &models.User{
		Email:    "chatuser3@example.com",
		Password: "hashed",
		Name:     "Chat User 3",
		Role:     "user",
		IsActive: true,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("userRepo.Create() error = %v", err)
	}

	channelRepo := repositories.NewChannelRepository(db)
	channel := &models.Channel{
		Name:     "Model Channel",
		Type:     "openai",
		IsActive: true,
	}
	if err := channelRepo.Create(channel); err != nil {
		t.Fatalf("channelRepo.Create() error = %v", err)
	}

	userAPIKeyRepo := repositories.NewUserAPIKeyRepository(db)
	modelRepo := repositories.NewModelRepository(db)
	userAPIKeyService := services.NewUserAPIKeyService(userAPIKeyRepo, userRepo, channelRepo, modelRepo)
	apiKey, err := userAPIKeyService.CreateAPIKey(user.ID, "Chat Key 3", 1000, nil, []uint{channel.ID}, []uint{})
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	channelAPIKeyRepo := repositories.NewChannelAPIKeyRepository(db)
	channelService := services.NewChannelService(channelRepo, channelAPIKeyRepo, modelRepo, userRepo)
	proxyService := services.NewProxyService(repositories.NewProxyRepository(db))

	if _, err := channelService.AddAPIKey(channel.ID); err != nil {
		t.Fatalf("AddAPIKey() error = %v", err)
	}

	if _, err := channelService.AddModel(channel.ID, "gpt-4", "GPT-4", 0.03, ""); err != nil {
		t.Fatalf("AddModel() error = %v", err)
	}

	handler := handlers.NewChatHandler(userAPIKeyService, channelService, proxyService, nil)
	return handler, apiKey, channel
}
