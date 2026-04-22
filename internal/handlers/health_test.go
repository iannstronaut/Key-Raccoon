package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"keyraccoon/internal/config"
	"keyraccoon/internal/handlers"
)

func TestHealthCheckWithoutDependencies(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)

	app := fiber.New()
	handlers.RegisterHealthRoutes(app)

	resp, err := app.Test(httptest.NewRequest("GET", "/health", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	body := decodeMap(t, resp)
	if body["status"] != "ok" {
		t.Fatalf("status = %v, want ok", body["status"])
	}
	if body["database_ok"] != false || body["redis_ok"] != false {
		t.Fatalf("unexpected dependency status: %+v", body)
	}
}

func TestHealthCheckWithDependencies(t *testing.T) {
	config.ResetForTesting()
	t.Cleanup(config.ResetForTesting)

	dbPath := filepath.Join(t.TempDir(), "health.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	config.SetDBForTesting(db)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	t.Cleanup(mini.Close)

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	config.SetRedisForTesting(client)

	app := fiber.New()
	handlers.RegisterHealthRoutes(app)

	resp, err := app.Test(httptest.NewRequest("GET", "/health", nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	body := decodeMap(t, resp)
	if body["database_ok"] != true || body["redis_ok"] != true {
		t.Fatalf("unexpected dependency status: %+v", body)
	}
}

func decodeMap(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return body
}
