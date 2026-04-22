# KeyRaccoon

KeyRaccoon adalah fondasi platform manajemen API dengan compatibility OpenAI SDK, manajemen user, channel routing, dan proxy management.

## Phase 1

Phase 1 menyiapkan:

- HTTP server berbasis Fiber
- CLI berbasis Cobra
- Konfigurasi environment
- PostgreSQL via GORM
- Redis bootstrap
- JWT utility
- Logging dasar
- Health endpoint

## Quick Start

```bash
copy config/.env.example config/.env.local
docker compose up -d
go mod tidy
go run ./cmd/server
```

Health check:

```bash
curl http://localhost:3000/health
```
