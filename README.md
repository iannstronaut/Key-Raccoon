# KeyRaccoon

![KeyRaccoon Banner](frontend/public/keyraccoon_banner.png)

OpenAI-compatible API gateway with multi-channel routing, per-channel budgets, usage logging, and a self-service dashboard.

## Features

### API Gateway
- **OpenAI-compatible** — drop-in replacement, works with any OpenAI SDK
- **Multi-channel routing** — route requests to multiple LLM providers (OpenAI, Anthropic, custom endpoints)
- **Model matching** — automatically selects the correct channel for the requested model
- **Proxy support** — optional SOCKS5/HTTP proxy with health checks and auto-rotation

### Budget & Usage
- **Per-channel budgets** — set a cost limit per channel (or unlimited), auto-reject with 429 when exceeded
- **Atomic budget tracking** — concurrent-safe via SQL `UPDATE ... SET budget_used = budget_used + ?`
- **Token tracking** — input/output tokens recorded per request
- **Cost calculation** — `(total_tokens / 1000) * model.token_price`
- **Request logging** — every request logged to PostgreSQL with model, user, channel, tokens, cost, latency, status

### User Management
- **JWT authentication** with role-based access (admin / user)
- **API key system** — `kr_` prefixed keys with token limits, usage limits, expiry
- **Self-service API keys** — users create their own keys from assigned channels
- **Channel assignment** — admin assigns channels to users, users only see their channels

### Dashboard
- **Admin**: manage users, channels, models, API keys, view all logs and analytics
- **User**: view assigned channels with budgets, create/delete own API keys, view own analytics
- **Real-time analytics** — stats auto-refresh every 60 seconds
- **Raycast-inspired dark UI** — glass-morphism theme

## Quick Start

### Docker (recommended)

```bash
git clone https://github.com/username/keyraccoon
cd keyraccoon

# Start everything (PostgreSQL + Redis + app)
docker compose up -d

# App is available at http://localhost
```

Default admin credentials:
- Email: `admin@keyraccoon.com`
- Password: `AdminPassword123`

### Configuration

Copy `.env.example` to `.env` and edit:

```bash
cp .env.example .env
```

```env
# Database
DB_USER=postgres
DB_PASSWORD=your-secure-password
DB_NAME=keyraccoon

# Redis
REDIS_PASS=your-redis-password

# JWT
JWT_SECRET=your-random-secret

# Admin seed account
ADMIN_EMAIL=admin@yourdomain.com
ADMIN_PASSWORD=your-admin-password

# CORS (use specific origin in production)
CORS_ORIGIN=https://yourdomain.com

# Port (default 80)
APP_PORT=80
```

Then:

```bash
docker compose up -d
```

### Local Development

Prerequisites: Go 1.24+, Node 20+, PostgreSQL 15+, Redis 7+

```bash
# Start database and redis
docker compose up -d postgres redis

# Run backend
go run ./cmd/server

# Run frontend (separate terminal)
cd frontend && npm install && npm run dev
```

Backend: `http://localhost:3000` | Frontend: `http://localhost:5173`

## API Usage

### Chat Completions (OpenAI-compatible)

```bash
curl http://localhost/api/v1/chat/completions \
  -H "Authorization: Bearer kr_your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### List Models

```bash
curl http://localhost/api/v1/models \
  -H "Authorization: Bearer kr_your-api-key"
```

### Authentication (Dashboard)

```bash
curl -X POST http://localhost/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@keyraccoon.com", "password": "AdminPassword123"}'
```

## Architecture

```
                    port 80 (nginx)
                     /          \
              static files    proxy /api/*
              (React SPA)         |
                            Go backend :3000
                             /        \
                       PostgreSQL    Redis
```

Single Docker container runs nginx (frontend + reverse proxy) and the Go binary via supervisord.

## Project Structure

```
keyraccoon/
├── cmd/server/              # Server entry point
├── internal/
│   ├── config/              # Database, Redis, env config
│   ├── models/              # GORM models (Channel, User, Model, RequestLog, ...)
│   ├── database/repositories/  # Data access layer
│   ├── services/            # Business logic (ChannelService, LogService, ...)
│   ├── handlers/            # HTTP handlers (ChatHandler, LogHandler, ...)
│   ├── middleware/          # Auth, API key, security middleware
│   └── routes/              # Route registration
├── pkg/logger/              # Structured logging
├── frontend/
│   ├── src/pages/           # React pages (Dashboard, Channels, Analytics, Logs, ...)
│   ├── src/services/api.ts  # API client
│   ├── src/contexts/        # Auth context
│   └── nginx.conf           # Nginx config (used in Docker)
├── Dockerfile               # Multi-stage build (Go + Node + nginx)
├── docker-compose.yml       # App + PostgreSQL + Redis
├── supervisord.conf         # Process manager for Docker
└── .env.example             # Environment template
```

## API Endpoints

### Public API (API key auth)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/chat/completions` | Chat completions (OpenAI-compatible) |
| POST | `/api/v1/embeddings` | Embeddings (stub) |
| GET | `/api/v1/models` | List available models |

### Dashboard API (JWT auth)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/login` | Login |
| GET | `/api/users` | List users (admin) |
| GET/POST/PUT/DELETE | `/api/channels/*` | Channel CRUD (admin) |
| POST | `/api/channels/:id/reset-budget` | Reset channel budget (admin) |
| GET | `/api/channels/:id/users` | List channel users (admin) |
| GET/POST/PUT/DELETE | `/api/user-api-keys/*` | API key management (admin) |
| GET | `/api/user-api-keys/my-channels` | User's assigned channels |
| POST | `/api/user-api-keys/self` | Create own API key |
| DELETE | `/api/user-api-keys/self/:id` | Delete own API key |
| GET | `/api/logs` | All request logs (admin) |
| GET | `/api/logs/stats` | Usage statistics (scoped to user for non-admin) |
| GET | `/api/logs/user/:id` | User's logs (admin or self) |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_PORT` | `80` | Exposed port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `password` | PostgreSQL password |
| `DB_NAME` | `keyraccoon` | Database name |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL mode |
| `REDIS_PASS` | _(empty)_ | Redis password (empty = no auth) |
| `JWT_SECRET` | `change-me-in-production` | JWT signing secret |
| `JWT_EXPIRE` | `60` | JWT expiry in minutes |
| `ADMIN_EMAIL` | `admin@keyraccoon.com` | Seed admin email |
| `ADMIN_PASSWORD` | `AdminPassword123` | Seed admin password |
| `CORS_ORIGIN` | `*` | Allowed CORS origins |
| `ENVIRONMENT` | `production` | Environment name |

## Development

```bash
make dev          # Run backend locally
make build        # Build binary
make test         # Run tests
make tidy         # Sync go.mod
make db-up        # Start postgres + redis
make db-down      # Stop postgres + redis
make docker-up    # docker compose up -d
make docker-down  # docker compose down
make docker-logs  # docker compose logs -f
```

## License

MIT
