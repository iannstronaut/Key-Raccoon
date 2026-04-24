# =============================================
# Stage 1: Build Go backend
# =============================================
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /app/bin/keyraccoon ./cmd/server

# =============================================
# Stage 2: Build React frontend
# =============================================
FROM node:20-alpine AS frontend-builder

WORKDIR /app

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./

ARG VITE_API_BASE_URL=""
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

RUN npm run build

# =============================================
# Stage 3: Runtime — nginx + Go binary
# =============================================
FROM alpine:3.20

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    nginx \
    supervisor

# Copy Go binary
COPY --from=backend-builder /app/bin/keyraccoon /usr/local/bin/keyraccoon

# Copy frontend build
COPY --from=frontend-builder /app/dist /usr/share/nginx/html

# Copy nginx config
COPY frontend/nginx.conf /etc/nginx/http.d/default.conf

# Copy supervisord config
COPY supervisord.conf /etc/supervisord.conf

# Create nginx pid directory
RUN mkdir -p /run/nginx

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost/health > /dev/null && \
        wget -qO- http://localhost:3000/health > /dev/null || exit 1

CMD ["supervisord", "-c", "/etc/supervisord.conf"]
