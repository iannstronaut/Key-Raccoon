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

func TestProxyRepositoryCRUD(t *testing.T) {
	db := openProxyRepoDB(t)
	repo := repositories.NewProxyRepository(db)

	proxy := &models.Proxy{
		URL:      "http://proxy.example.com:8080",
		Type:     "http",
		Username: "user",
		Password: "pass",
		IsActive: true,
		Status:   "unknown",
	}
	if err := repo.Create(proxy); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.GetByID(proxy.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if found.URL != proxy.URL {
		t.Fatalf("GetByID() URL = %q, want %q", found.URL, proxy.URL)
	}

	proxies, total, err := repo.GetAll(10, 0)
	if err != nil || len(proxies) != 1 || total != 1 {
		t.Fatalf("GetAll() got len=%d total=%d err=%v", len(proxies), total, err)
	}

	if err := repo.UpdateStatus(proxy.ID, "healthy"); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	if err := repo.UpdateLastCheck(proxy.ID); err != nil {
		t.Fatalf("UpdateLastCheck() error = %v", err)
	}

	healthy, err := repo.GetHealthy()
	if err != nil || healthy.ID != proxy.ID {
		t.Fatalf("GetHealthy() = %+v, %v", healthy, err)
	}

	active, err := repo.GetActive()
	if err != nil || len(active) != 1 {
		t.Fatalf("GetActive() len=%d err=%v", len(active), err)
	}

	proxy.Status = "unhealthy"
	if err := repo.Update(proxy); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	found, _ = repo.GetByID(proxy.ID)
	if found.Status != "unhealthy" {
		t.Fatalf("status = %q, want unhealthy", found.Status)
	}

	if err := repo.Disable(proxy.ID); err != nil {
		t.Fatalf("Disable() error = %v", err)
	}
	found, _ = repo.GetByID(proxy.ID)
	if found.IsActive {
		t.Fatal("expected proxy to be disabled")
	}

	if err := repo.Delete(proxy.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repo.GetByID(proxy.ID); !errors.Is(err, repositories.ErrProxyNotFound) {
		t.Fatalf("GetByID() after delete error = %v, want ErrProxyNotFound", err)
	}
}

func TestProxyRepositoryNotFound(t *testing.T) {
	db := openProxyRepoDB(t)
	repo := repositories.NewProxyRepository(db)

	if _, err := repo.GetByID(999); !errors.Is(err, repositories.ErrProxyNotFound) {
		t.Fatalf("GetByID() error = %v, want ErrProxyNotFound", err)
	}
	if err := repo.UpdateStatus(999, "healthy"); !errors.Is(err, repositories.ErrProxyNotFound) {
		t.Fatalf("UpdateStatus() error = %v, want ErrProxyNotFound", err)
	}
	if err := repo.UpdateLastCheck(999); !errors.Is(err, repositories.ErrProxyNotFound) {
		t.Fatalf("UpdateLastCheck() error = %v, want ErrProxyNotFound", err)
	}
	if err := repo.Disable(999); !errors.Is(err, repositories.ErrProxyNotFound) {
		t.Fatalf("Disable() error = %v, want ErrProxyNotFound", err)
	}
	if err := repo.Delete(999); !errors.Is(err, repositories.ErrProxyNotFound) {
		t.Fatalf("Delete() error = %v, want ErrProxyNotFound", err)
	}
}

func openProxyRepoDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "proxy_repository.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.Proxy{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return db
}
