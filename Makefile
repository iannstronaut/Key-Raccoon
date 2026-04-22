.PHONY: help setup dev build run cli test tidy db-up db-down

help:
	@echo "Available commands:"
	@echo "  make setup   - Prepare local env file and download deps"
	@echo "  make dev     - Run HTTP server"
	@echo "  make build   - Build server binary"
	@echo "  make run     - Run built server binary"
	@echo "  make cli     - Run CLI"
	@echo "  make test    - Run tests"
	@echo "  make tidy    - Sync go.mod/go.sum"
	@echo "  make db-up   - Start postgres and redis"
	@echo "  make db-down - Stop postgres and redis"

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

tidy:
	go mod tidy

db-up:
	docker compose up -d

db-down:
	docker compose down
