package repositories_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
)

func TestAPIKeyRepositoryCRUDAndUsage(t *testing.T) {
	repo, userRepo, channelRepo := openAPIKeyRepository(t)

	user := &models.User{
		Email:       "apikeyuser@example.com",
		Password:    "hashed",
		Name:        "API Key User",
		Role:        "user",
		IsActive:    true,
		TokenLimit:  1000,
		CreditLimit: 100,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("userRepo.Create() error = %v", err)
	}

	channel := &models.Channel{
		Name:     "Test Channel",
		Type:     "openai",
		IsActive: true,
	}
	if err := channelRepo.Create(channel); err != nil {
		t.Fatalf("channelRepo.Create() error = %v", err)
	}

	keyStr, err := repo.GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey() error = %v", err)
	}
	if keyStr == "" {
		t.Fatal("GenerateAPIKey() returned empty string")
	}

	apiKey := &models.APIKey{
		UserID:      user.ID,
		Key:         keyStr,
		Name:        "Test Key",
		TokenLimit:  500,
		CreditLimit: 50,
		TokenUsed:   0,
		CreditUsed:  0,
		IsActive:    true,
	}
	if err := repo.Create(apiKey); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	foundByID, err := repo.GetByID(apiKey.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if foundByID.Key != apiKey.Key {
		t.Fatalf("GetByID() key = %q, want %q", foundByID.Key, apiKey.Key)
	}

	foundByKey, err := repo.GetByKey(apiKey.Key)
	if err != nil {
		t.Fatalf("GetByKey() error = %v", err)
	}
	if foundByKey.ID != apiKey.ID {
		t.Fatalf("GetByKey() id = %d, want %d", foundByKey.ID, apiKey.ID)
	}

	keys, err := repo.GetByUserID(user.ID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("GetByUserID() len = %d, want 1", len(keys))
	}

	activeKeys, err := repo.GetActiveByUserID(user.ID)
	if err != nil {
		t.Fatalf("GetActiveByUserID() error = %v", err)
	}
	if len(activeKeys) != 1 {
		t.Fatalf("GetActiveByUserID() len = %d, want 1", len(activeKeys))
	}

	if err := repo.UpdateFields(apiKey.ID, map[string]any{"name": "Updated Key"}); err != nil {
		t.Fatalf("UpdateFields() error = %v", err)
	}
	foundByID, _ = repo.GetByID(apiKey.ID)
	if foundByID.Name != "Updated Key" {
		t.Fatalf("name = %q, want Updated Key", foundByID.Name)
	}

	if err := repo.UpdateTokenUsage(apiKey.ID, 100); err != nil {
		t.Fatalf("UpdateTokenUsage() error = %v", err)
	}
	if err := repo.UpdateCreditUsage(apiKey.ID, 12.5); err != nil {
		t.Fatalf("UpdateCreditUsage() error = %v", err)
	}

	ok, err := repo.HasTokenAvailable(apiKey.ID, 401)
	if err != nil || ok {
		t.Fatalf("HasTokenAvailable() = %v, %v, want false, nil", ok, err)
	}
	ok, err = repo.HasCreditAvailable(apiKey.ID, 38)
	if err != nil || ok {
		t.Fatalf("HasCreditAvailable() = %v, %v, want false, nil", ok, err)
	}

	if err := repo.UpdateLastUsed(apiKey.ID); err != nil {
		t.Fatalf("UpdateLastUsed() error = %v", err)
	}
	foundByID, _ = repo.GetByID(apiKey.ID)
	if foundByID.LastUsed == nil {
		t.Fatal("UpdateLastUsed() last_used is nil")
	}

	if err := repo.BindChannel(apiKey.ID, channel.ID); err != nil {
		t.Fatalf("BindChannel() error = %v", err)
	}
	foundByID, _ = repo.GetByID(apiKey.ID)
	if len(foundByID.Channels) != 1 {
		t.Fatalf("BindChannel() channels len = %d, want 1", len(foundByID.Channels))
	}

	if err := repo.UnbindChannel(apiKey.ID, channel.ID); err != nil {
		t.Fatalf("UnbindChannel() error = %v", err)
	}
	foundByID, _ = repo.GetByID(apiKey.ID)
	if len(foundByID.Channels) != 0 {
		t.Fatalf("UnbindChannel() channels len = %d, want 0", len(foundByID.Channels))
	}

	if err := repo.Disable(apiKey.ID); err != nil {
		t.Fatalf("Disable() error = %v", err)
	}
	foundByID, _ = repo.GetByID(apiKey.ID)
	if foundByID.IsActive {
		t.Fatal("Disable() is_active = true, want false")
	}

	if err := repo.Delete(apiKey.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repo.GetByID(apiKey.ID); !errors.Is(err, repositories.ErrAPIKeyNotFound) {
		t.Fatalf("GetByID() after delete error = %v, want ErrAPIKeyNotFound", err)
	}
}

func TestAPIKeyRepositoryUnlimitedLimits(t *testing.T) {
	repo, userRepo, _ := openAPIKeyRepository(t)

	user := &models.User{
		Email:    "unlimited@example.com",
		Password: "hashed",
		Name:     "Unlimited User",
		Role:     "user",
		IsActive: true,
	}
	if err := userRepo.Create(user); err != nil {
		t.Fatalf("userRepo.Create() error = %v", err)
	}

	apiKey := &models.APIKey{
		UserID:      user.ID,
		Key:         "pk-unlimited-test-key-1234567890123456789012345678",
		Name:        "Unlimited Key",
		TokenLimit:  -1,
		CreditLimit: -1,
		IsActive:    true,
	}
	if err := repo.Create(apiKey); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	ok, err := repo.HasTokenAvailable(apiKey.ID, 999999999)
	if err != nil || !ok {
		t.Fatalf("HasTokenAvailable(unlimited) = %v, %v, want true, nil", ok, err)
	}
	ok, err = repo.HasCreditAvailable(apiKey.ID, 999999999.99)
	if err != nil || !ok {
		t.Fatalf("HasCreditAvailable(unlimited) = %v, %v, want true, nil", ok, err)
	}
}

func openAPIKeyRepository(t *testing.T) (*repositories.APIKeyRepository, *repositories.UserRepository, *repositories.ChannelRepository) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "apikey_repository.db")
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
	return repositories.NewAPIKeyRepository(db), repositories.NewUserRepository(db), repositories.NewChannelRepository(db)
}
