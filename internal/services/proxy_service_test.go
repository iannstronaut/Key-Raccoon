package services_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
	"keyraccoon/internal/services"
)

func TestProxyServiceAddAndGet(t *testing.T) {
	repo, db := openProxyService(t)
	svc := services.NewProxyService(repo)

	proxy, err := svc.AddProxy("http://proxy.example.com:8080", "http", "user", "pass")
	if err != nil {
		t.Fatalf("AddProxy() error = %v", err)
	}
	if proxy.URL != "http://proxy.example.com:8080" {
		t.Fatalf("URL = %q", proxy.URL)
	}

	found, err := svc.GetProxy(proxy.ID)
	if err != nil || found.ID != proxy.ID {
		t.Fatalf("GetProxy() = %+v, %v", found, err)
	}

	proxies, total, err := svc.GetAllProxies(10, 0)
	if err != nil || len(proxies) != 1 || total != 1 {
		t.Fatalf("GetAllProxies() len=%d total=%d err=%v", len(proxies), total, err)
	}

	if err := svc.DeleteProxy(proxy.ID); err != nil {
		t.Fatalf("DeleteProxy() error = %v", err)
	}
	if _, err := svc.GetProxy(proxy.ID); !errors.Is(err, repositories.ErrProxyNotFound) {
		t.Fatalf("GetProxy() after delete error = %v", err)
	}

	_ = db
}

func TestProxyServiceAddProxyInvalidURL(t *testing.T) {
	repo, _ := openProxyService(t)
	svc := services.NewProxyService(repo)

	if _, err := svc.AddProxy("://invalid-url", "http", "", ""); err == nil {
		t.Fatal("AddProxy() error = nil, want error")
	}
	if _, err := svc.AddProxy("not-a-url", "http", "", ""); err == nil {
		t.Fatal("AddProxy() missing scheme error = nil, want error")
	}
}

func TestProxyServiceCheckProxyHealth(t *testing.T) {
	repo, _ := openProxyService(t)
	svc := services.NewProxyService(repo)

	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer proxyServer.Close()

	oldURL := services.HealthCheckURL
	services.HealthCheckURL = proxyServer.URL
	defer func() { services.HealthCheckURL = oldURL }()

	proxy, err := svc.AddProxy(proxyServer.URL, "http", "", "")
	if err != nil {
		t.Fatalf("AddProxy() error = %v", err)
	}

	if err := svc.CheckProxyHealth(proxy.ID); err != nil {
		t.Fatalf("CheckProxyHealth() error = %v", err)
	}

	updated, _ := repo.GetByID(proxy.ID)
	if updated.Status != "healthy" {
		t.Fatalf("status = %q, want healthy", updated.Status)
	}
	if updated.LastCheck == nil {
		t.Fatal("expected last_check to be set")
	}
}

func TestProxyServiceTestProxy(t *testing.T) {
	repo, _ := openProxyService(t)
	svc := services.NewProxyService(repo)

	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer proxyServer.Close()

	oldURL := services.HealthCheckURL
	services.HealthCheckURL = proxyServer.URL
	defer func() { services.HealthCheckURL = oldURL }()

	isHealthy, err := svc.TestProxy(proxyServer.URL, "http", "", "")
	if err != nil {
		t.Fatalf("TestProxy() error = %v", err)
	}
	if !isHealthy {
		t.Fatal("TestProxy() = false, want true")
	}

	isHealthy, err = svc.TestProxy("http://127.0.0.1:1", "http", "", "")
	if err != nil {
		t.Fatalf("TestProxy() error = %v", err)
	}
	if isHealthy {
		t.Fatal("TestProxy() = true, want false")
	}
}

func TestProxyServiceRotationAndFallback(t *testing.T) {
	repo, _ := openProxyService(t)
	svc := services.NewProxyService(repo)

	proxy1 := &models.Proxy{URL: "http://p1.com", Type: "http", IsActive: true, Status: "healthy"}
	proxy2 := &models.Proxy{URL: "http://p2.com", Type: "http", IsActive: true, Status: "healthy"}
	if err := repo.Create(proxy1); err != nil {
		t.Fatalf("Create(p1) error = %v", err)
	}
	if err := repo.Create(proxy2); err != nil {
		t.Fatalf("Create(p2) error = %v", err)
	}

	healthy, err := svc.GetHealthyProxy()
	if err != nil || healthy.ID != proxy1.ID {
		t.Fatalf("GetHealthyProxy() = %+v, %v", healthy, err)
	}

	fallback, err := svc.GetProxyWithFallback()
	if err != nil || fallback.ID != proxy1.ID {
		t.Fatalf("GetProxyWithFallback() = %+v, %v", fallback, err)
	}

	rotated, err := svc.RotateProxy(proxy1.ID)
	if err != nil || rotated.ID != proxy2.ID {
		t.Fatalf("RotateProxy() = %+v, %v", rotated, err)
	}

	p1, _ := repo.GetByID(proxy1.ID)
	if p1.Status != "unhealthy" {
		t.Fatalf("p1 status = %q, want unhealthy", p1.Status)
	}
}

func openProxyService(t *testing.T) (*repositories.ProxyRepository, *gorm.DB) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "proxy_service.db")
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
	return repositories.NewProxyRepository(db), db
}
