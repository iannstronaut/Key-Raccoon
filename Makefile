.PHONY: help dev build run test tidy db-up db-down docker-up docker-down docker-logs docker-build docker-clean

help:
	@echo "Available commands:"
	@echo ""
	@echo "  make dev          Run backend locally (go run)"
	@echo "  make build        Build server binary"
	@echo "  make run          Run built binary"
	@echo "  make test         Run tests"
	@echo "  make tidy         Sync go.mod/go.sum"
	@echo ""
	@echo "  make db-up        Start postgres + redis"
	@echo "  make db-down      Stop postgres + redis"
	@echo "  make db-shell     PostgreSQL shell"
	@echo "  make redis-shell  Redis shell"
	@echo ""
	@echo "  make docker-up    Start everything (app + db + redis)"
	@echo "  make docker-down  Stop everything"
	@echo "  make docker-logs  Follow logs"
	@echo "  make docker-build Build Docker image"
	@echo "  make docker-clean Remove containers and volumes"

# ============================================
# Local Development
# ============================================
dev:
	go run ./cmd/server

build:
	go build -o bin/keyraccoon ./cmd/server

run:
	./bin/keyraccoon

test:
	go test ./...

tidy:
	go mod tidy

# ============================================
# Database (local dev — only postgres + redis)
# ============================================
db-up:
	docker compose up -d postgres redis

db-down:
	docker compose stop postgres redis

db-shell:
	docker compose exec postgres psql -U postgres -d keyraccoon

redis-shell:
	docker compose exec redis redis-cli

# ============================================
# Docker (full stack)
# ============================================
docker-up:
	docker compose up -d
	@echo "KeyRaccoon running at http://localhost"

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-build:
	docker compose build

docker-clean:
	docker compose down -v
	@echo "Containers and volumes removed"
