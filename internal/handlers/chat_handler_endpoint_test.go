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

func TestChatHandlerUsesChannelEndpoint(t *testing.T) {
	// Setup database
	dbPath := filepath.Join(t.TempDir(), "chat_endpoint_test.db")
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

	// Create user
	userRepo := repositories.NewUserRepository(db)
	user := &models.User{
		Email:    "test@example.com",
		Password: "hashed",
		Name:     "Test User",
		Role:     "user",
		IsActive: true,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("userRepo.Create() error = %v", err)
	}

	// Create mock upstream server
	var receivedURL string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.String()
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
						"content": "Test response",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]int{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// Create channel with custom endpoint
	channelRepo := repositories.NewChannelRepository(db)
	channel := &models.Channel{
		Name:     "Custom Endpoint Channel",
		Type:     "openai",
		Endpoint: mockServer.URL, // Use mock server URL
		IsActive: true,
	}
	if err := channelRepo.Create(channel); err != nil {
		t.Fatalf("channelRepo.Create() error = %v", err)
	}

	// Add channel API key
	channelAPIKeyRepo := repositories.NewChannelAPIKeyRepository(db)
	channelAPIKey := &models.ChannelAPIKey{
		ChannelID: channel.ID,
		APIKey:    "test-channel-key",
		IsActive:  true,
	}
	if err := db.Create(channelAPIKey).Error; err != nil {
		t.Fatalf("Create channel API key error = %v", err)
	}

	// Create user API key with channel binding
	userAPIKeyRepo := repositories.NewUserAPIKeyRepository(db)
	modelRepo := repositories.NewModelRepository(db)
	userAPIKeyService := services.NewUserAPIKeyService(userAPIKeyRepo, userRepo, channelRepo, modelRepo)
	apiKey, err := userAPIKeyService.CreateAPIKey(user.ID, "Test Key", 0, nil, []uint{channel.ID}, []uint{})
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	// Create handler
	channelService := services.NewChannelService(channelRepo, channelAPIKeyRepo, modelRepo, userRepo)
	proxyService := services.NewProxyService(repositories.NewProxyRepository(db))
	handler := handlers.NewChatHandler(userAPIKeyService, channelService, proxyService)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/chat/completions", func(c *fiber.Ctx) error {
		c.Locals("api_key_id", apiKey.ID)
		c.Locals("api_key_channels", []models.Channel{*channel})
		return handler.ChatCompletion(c)
	})

	// Make request
	body := map[string]interface{}{
		"model":    "gpt-4",
		"messages": []map[string]string{{"role": "user", "content": "Hello"}},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000) // 5 second timeout
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	// Verify that the request was sent to the custom endpoint
	if receivedURL != "/chat/completions" {
		t.Fatalf("receivedURL = %s, want /chat/completions", receivedURL)
	}

	t.Logf("✅ Successfully used channel endpoint: %s", mockServer.URL)
}
