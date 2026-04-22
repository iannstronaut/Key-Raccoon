package services_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"keyraccoon/internal/config"
	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
	"keyraccoon/internal/services"
)

func TestAPIKeyServiceLifecycle(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)

	mini := startMiniRedis(t)
	cfg := &config.Config{
		RedisHost: mini.Host(),
		RedisPort: mini.Port(),
	}
	if err := config.InitRedis(cfg); err != nil {
		t.Fatalf("InitRedis() error = %v", err)
	}

	service, userRepo, channelRepo := openAPIKeyService(t)

	user := &models.User{
		Email:       "serviceuser@example.com",
		Password:    "hashed",
		Name:        "Service User",
		Role:        "user",
		IsActive:    true,
		TokenLimit:  1000,
		CreditLimit: 100,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("userRepo.Create() error = %v", err)
	}

	channel := &models.Channel{
		Name:     "Service Channel",
		Type:     "openai",
		IsActive: true,
	}
	if err := channelRepo.Create(channel); err != nil {
		t.Fatalf("channelRepo.Create() error = %v", err)
	}

	apiKey, err := service.CreateAPIKey(user.ID, "My Key", 500, 50)
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}
	if apiKey.UserID != user.ID {
		t.Fatalf("CreateAPIKey() user_id = %d, want %d", apiKey.UserID, user.ID)
	}
	if apiKey.Key == "" {
		t.Fatal("CreateAPIKey() key is empty")
	}

	keys, err := service.GetUserAPIKeys(user.ID)
	if err != nil {
		t.Fatalf("GetUserAPIKeys() error = %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("GetUserAPIKeys() len = %d, want 1", len(keys))
	}

	foundByID, err := service.GetAPIKeyByID(apiKey.ID)
	if err != nil {
		t.Fatalf("GetAPIKeyByID() error = %v", err)
	}
	if foundByID.ID != apiKey.ID {
		t.Fatalf("GetAPIKeyByID() id = %d, want %d", foundByID.ID, apiKey.ID)
	}

	foundByKey, err := service.GetAPIKeyByKey(apiKey.Key)
	if err != nil {
		t.Fatalf("GetAPIKeyByKey() error = %v", err)
	}
	if foundByKey.ID != apiKey.ID {
		t.Fatalf("GetAPIKeyByKey() id = %d, want %d", foundByKey.ID, apiKey.ID)
	}

	verified, err := service.VerifyAPIKey(apiKey.Key)
	if err != nil {
		t.Fatalf("VerifyAPIKey() error = %v", err)
	}
	if verified.ID != apiKey.ID {
		t.Fatalf("VerifyAPIKey() id = %d, want %d", verified.ID, apiKey.ID)
	}

	_, err = service.VerifyAPIKey("invalid-key")
	if err == nil {
		t.Fatal("VerifyAPIKey(invalid) error = nil, want error")
	}

	updated, err := service.UpdateAPIKey(apiKey.ID, map[string]any{"name": "Updated Key"})
	if err != nil {
		t.Fatalf("UpdateAPIKey() error = %v", err)
	}
	if updated.Name != "Updated Key" {
		t.Fatalf("UpdateAPIKey() name = %q, want Updated Key", updated.Name)
	}

	if err := service.UpdateLastUsed(apiKey.ID); err != nil {
		t.Fatalf("UpdateLastUsed() error = %v", err)
	}

	ok, err := service.HasTokenAvailable(apiKey.ID, 400)
	if err != nil || !ok {
		t.Fatalf("HasTokenAvailable(400) = %v, %v, want true, nil", ok, err)
	}
	ok, err = service.HasTokenAvailable(apiKey.ID, 501)
	if err != nil || ok {
		t.Fatalf("HasTokenAvailable(501) = %v, %v, want false, nil", ok, err)
	}

	ok, err = service.HasCreditAvailable(apiKey.ID, 40)
	if err != nil || !ok {
		t.Fatalf("HasCreditAvailable(40) = %v, %v, want true, nil", ok, err)
	}
	ok, err = service.HasCreditAvailable(apiKey.ID, 51)
	if err != nil || ok {
		t.Fatalf("HasCreditAvailable(51) = %v, %v, want false, nil", ok, err)
	}

	if err := service.RecordTokenUsage(apiKey.ID, 50); err != nil {
		t.Fatalf("RecordTokenUsage() error = %v", err)
	}
	if err := service.RecordCreditUsage(apiKey.ID, 5.5); err != nil {
		t.Fatalf("RecordCreditUsage() error = %v", err)
	}

	usage, err := service.GetRealtimeUsage(apiKey.ID)
	if err != nil {
		t.Fatalf("GetRealtimeUsage() error = %v", err)
	}
	if usage["token_used"] != int64(50) {
		t.Fatalf("GetRealtimeUsage() token_used = %v, want 50", usage["token_used"])
	}
	if usage["credit_used"] != 5.5 {
		t.Fatalf("GetRealtimeUsage() credit_used = %v, want 5.5", usage["credit_used"])
	}

	if err := service.BindChannel(apiKey.ID, channel.ID); err != nil {
		t.Fatalf("BindChannel() error = %v", err)
	}
	foundByID, _ = service.GetAPIKeyByID(apiKey.ID)
	if len(foundByID.Channels) != 1 {
		t.Fatalf("BindChannel() channels len = %d, want 1", len(foundByID.Channels))
	}

	if err := service.UnbindChannel(apiKey.ID, channel.ID); err != nil {
		t.Fatalf("UnbindChannel() error = %v", err)
	}
	foundByID, _ = service.GetAPIKeyByID(apiKey.ID)
	if len(foundByID.Channels) != 0 {
		t.Fatalf("UnbindChannel() channels len = %d, want 0", len(foundByID.Channels))
	}

	if err := service.DisableAPIKey(apiKey.ID); err != nil {
		t.Fatalf("DisableAPIKey() error = %v", err)
	}
	_, err = service.VerifyAPIKey(apiKey.Key)
	if err == nil {
		t.Fatal("VerifyAPIKey(disabled) error = nil, want error")
	}

	if err := service.DeleteAPIKey(apiKey.ID); err != nil {
		t.Fatalf("DeleteAPIKey() error = %v", err)
	}
	if _, err := service.GetAPIKeyByID(apiKey.ID); !errors.Is(err, repositories.ErrAPIKeyNotFound) {
		t.Fatalf("GetAPIKeyByID() after delete error = %v, want ErrAPIKeyNotFound", err)
	}
}

func TestAPIKeyServiceUnlimitedLimits(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)

	mini := startMiniRedis(t)
	cfg := &config.Config{
		RedisHost: mini.Host(),
		RedisPort: mini.Port(),
	}
	if err := config.InitRedis(cfg); err != nil {
		t.Fatalf("InitRedis() error = %v", err)
	}

	service, userRepo, _ := openAPIKeyService(t)

	user := &models.User{
		Email:    "unlimitedsvc@example.com",
		Password: "hashed",
		Name:     "Unlimited Service User",
		Role:     "user",
		IsActive: true,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("userRepo.Create() error = %v", err)
	}

	apiKey, err := service.CreateAPIKey(user.ID, "Unlimited Key", -1, -1)
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	ok, err := service.HasTokenAvailable(apiKey.ID, 999999999)
	if err != nil || !ok {
		t.Fatalf("HasTokenAvailable(unlimited) = %v, %v, want true, nil", ok, err)
	}
	ok, err = service.HasCreditAvailable(apiKey.ID, 999999999.99)
	if err != nil || !ok {
		t.Fatalf("HasCreditAvailable(unlimited) = %v, %v, want true, nil", ok, err)
	}
}

func TestAPIKeyServiceWithoutRedis(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)

	service, userRepo, _ := openAPIKeyService(t)

	user := &models.User{
		Email:    "noredis@example.com",
		Password: "hashed",
		Name:     "No Redis User",
		Role:     "user",
		IsActive: true,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("userRepo.Create() error = %v", err)
	}

	apiKey, err := service.CreateAPIKey(user.ID, "No Redis Key", 100, 10)
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	if err := service.RecordTokenUsage(apiKey.ID, 10); err != nil {
		t.Fatalf("RecordTokenUsage() without redis error = %v", err)
	}
	if err := service.RecordCreditUsage(apiKey.ID, 1.5); err != nil {
		t.Fatalf("RecordCreditUsage() without redis error = %v", err)
	}

	usage, err := service.GetRealtimeUsage(apiKey.ID)
	if err != nil {
		t.Fatalf("GetRealtimeUsage() without redis error = %v", err)
	}
	if usage["token_used"] != int64(10) {
		t.Fatalf("GetRealtimeUsage() token_used = %v, want 10", usage["token_used"])
	}
	if usage["credit_used"] != 1.5 {
		t.Fatalf("GetRealtimeUsage() credit_used = %v, want 1.5", usage["credit_used"])
	}
}

func startMiniRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	t.Cleanup(mini.Close)
	return mini
}

func openAPIKeyService(t *testing.T) (*services.APIKeyService, *repositories.UserRepository, *repositories.ChannelRepository) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "apikey_service.db")
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

	apiKeyRepo := repositories.NewAPIKeyRepository(db)
	userRepo := repositories.NewUserRepository(db)
	channelRepo := repositories.NewChannelRepository(db)

	return services.NewAPIKeyService(apiKeyRepo, userRepo), userRepo, channelRepo
}
