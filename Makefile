.PHONY: help setup dev build run cli test tidy db-up db-down docker-dev docker-prod docker-logs docker-down docker-clean

help:
	@echo "Available commands:"
	@echo ""
	@echo "Local Development:"
	@echo "  make setup       - Prepare local env file and download deps"
	@echo "  make dev         - Run HTTP server locally"
	@echo "  make build       - Build server binary"
	@echo "  make run         - Run built server binary"
	@echo "  make cli         - Run CLI"
	@echo "  make test        - Run tests"
	@echo "  make tidy        - Sync go.mod/go.sum"
	@echo ""
	@echo "Docker Development:"
	@echo "  make docker-dev      - Start development environment with hot reload"
	@echo "  make docker-dev-logs - View development logs"
	@echo "  make docker-dev-down - Stop development environment"
	@echo ""
	@echo "Docker Production:"
	@echo "  make docker-prod      - Start production environment"
	@echo "  make docker-prod-logs - View production logs"
	@echo "  make docker-prod-down - Stop production environment"
	@echo ""
	@echo "Docker Management:"
	@echo "  make docker-build    - Build Docker images"
	@echo "  make docker-ps       - Show running containers"
	@echo "  make docker-clean    - Remove containers and volumes"
	@echo ""
	@echo "Database:"
	@echo "  make db-up       - Start postgres and redis"
	@echo "  make db-down     - Stop postgres and redis"
	@echo "  make db-shell    - Access PostgreSQL shell"
	@echo "  make db-backup   - Backup database"

# ============================================
# Local Development
# ============================================
setup:
	@if not exist config\\.env.local copy config\\.env.example config\\.env.local
	go mod download

dev:
	go run ./cmd/server

build:
	go build -o bin/keyraccoon.exe ./cmd/server

run:
	./bin/keyraccoon.exe

cli:
	go run ./cmd/cli config

test:
	go test ./...

test-coverage:
	go test -cover ./...

tidy:
	go mod tidy

# ============================================
# Docker Development
# ============================================
docker-dev:
	docker-compose -f docker-compose.dev.yml up -d
	@echo "Development environment started"
	@echo "Frontend: http://localhost:5173"
	@echo "Backend:  http://localhost:3000"

docker-dev-logs:
	docker-compose -f docker-compose.dev.yml logs -f

docker-dev-down:
	docker-compose -f docker-compose.dev.yml down

docker-dev-restart:
	docker-compose -f docker-compose.dev.yml restart

# ============================================
# Docker Production
# ============================================
docker-prod:
	@if not exist .env (echo .env file not found. Copy .env.example to .env first. && exit 1)
	docker-compose -f docker-compose.production.yml up -d
	@echo "Production environment started"
	@echo "Frontend: http://localhost"
	@echo "Backend:  http://localhost:3000"

docker-prod-logs:
	docker-compose -f docker-compose.production.yml logs -f

docker-prod-down:
	docker-compose -f docker-compose.production.yml down

docker-prod-restart:
	docker-compose -f docker-compose.production.yml restart

# ============================================
# Docker Management
# ============================================
docker-build:
	docker-compose build

docker-build-dev:
	docker-compose -f docker-compose.dev.yml build

docker-build-prod:
	docker-compose -f docker-compose.production.yml build

docker-ps:
	docker-compose ps

docker-stats:
	docker stats

docker-clean:
	docker-compose down -v
	@echo "Containers and volumes removed"

docker-clean-all:
	docker-compose down -v --rmi all
	@echo "Containers, volumes, and images removed"

# ============================================
# Database
# ============================================
db-up:
	docker compose up -d postgres redis

db-down:
	docker compose down

db-shell:
	docker-compose exec postgres psql -U postgres -d keyraccoon

db-backup:
	@if not exist backups mkdir backups
	docker-compose exec postgres pg_dump -U postgres keyraccoon > backups/backup_%date:~-4,4%%date:~-10,2%%date:~-7,2%.sql
	@echo "Database backed up to backups/"

redis-shell:
	docker-compose exec redis redis-cli

# ============================================
# Utility
# ============================================
logs:
	docker-compose logs -f

logs-backend:
	docker-compose logs -f backend

logs-frontend:
	docker-compose logs -f frontend

shell-backend:
	docker-compose exec backend sh

shell-frontend:
	docker-compose exec frontend sh

health-check:
	@echo "Checking service health..."
	@curl -f http://localhost:3000/health && echo "Backend healthy" || echo "Backend unhealthy"
	@curl -f http://localhost/health && echo "Frontend healthy" || echo "Frontend unhealthy"
