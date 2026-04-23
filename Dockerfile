# ============================================
# Backend Dockerfile
# ============================================
# Build stage
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o bin/keyraccoon ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=backend-builder /app/bin/keyraccoon .

# Copy config files if needed
COPY --from=backend-builder /app/config/ ./config/

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

CMD ["./keyraccoon"]
