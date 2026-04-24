package services

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"keyraccoon/internal/config"
	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
	"keyraccoon/pkg/logger"
)

const (
	// Redis keys for real-time counters (dashboard widgets)
	redisKeyTodayRequests = "stats:today:requests"
	redisKeyTodayTokens   = "stats:today:tokens"
)

type LogService struct {
	logRepo *repositories.RequestLogRepository
	redis   *redis.Client
}

func NewLogService(logRepo *repositories.RequestLogRepository) *LogService {
	return &LogService{
		logRepo: logRepo,
		redis:   config.GetRedis(),
	}
}

// LogRequest writes a log entry directly to PostgreSQL and updates Redis counters.
// This is called from a goroutine so it won't block the HTTP response.
func (s *LogService) LogRequest(log models.RequestLog) {
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	// Write directly to PostgreSQL — immediate persistence
	if err := s.logRepo.Create(&log); err != nil {
		logger.Error("failed to write request log to database", "error", err)
		return
	}

	// Update Redis real-time counters (best-effort, non-blocking)
	s.updateRealtimeCounters(log)
}

// updateRealtimeCounters bumps Redis counters for dashboard widgets
func (s *LogService) updateRealtimeCounters(log models.RequestLog) {
	if s.redis == nil {
		return
	}

	ctx := context.Background()
	today := time.Now().Format("2006-01-02")
	ttl := 48 * time.Hour // keep counters for 2 days

	reqKey := fmt.Sprintf("%s:%s", redisKeyTodayRequests, today)
	tokKey := fmt.Sprintf("%s:%s", redisKeyTodayTokens, today)

	pipe := s.redis.Pipeline()
	pipe.Incr(ctx, reqKey)
	pipe.Expire(ctx, reqKey, ttl)
	pipe.IncrBy(ctx, tokKey, log.TotalTokens)
	pipe.Expire(ctx, tokKey, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		logger.Error("failed to update Redis real-time counters", "error", err)
	}
}

// GetTodayRealtimeStats returns today's request count and token count from Redis (fast)
func (s *LogService) GetTodayRealtimeStats() (requests int64, tokens int64) {
	if s.redis == nil {
		return 0, 0
	}

	ctx := context.Background()
	today := time.Now().Format("2006-01-02")

	reqKey := fmt.Sprintf("%s:%s", redisKeyTodayRequests, today)
	tokKey := fmt.Sprintf("%s:%s", redisKeyTodayTokens, today)

	requests, _ = s.redis.Get(ctx, reqKey).Int64()
	tokens, _ = s.redis.Get(ctx, tokKey).Int64()
	return
}

// GetLogs returns paginated logs with filters (admin)
func (s *LogService) GetLogs(limit, offset int, filters repositories.LogFilters) ([]models.RequestLog, int64, error) {
	return s.logRepo.GetAll(limit, offset, filters)
}

// GetLogsByAPIKey returns logs for a specific API key
func (s *LogService) GetLogsByAPIKey(keyID uint, limit, offset int) ([]models.RequestLog, int64, error) {
	return s.logRepo.GetByAPIKeyID(keyID, limit, offset)
}

// GetLogsByUser returns logs for a specific user
func (s *LogService) GetLogsByUser(userID uint, limit, offset int) ([]models.RequestLog, int64, error) {
	return s.logRepo.GetByUserID(userID, limit, offset)
}

// GetUsageStats returns aggregated usage statistics
func (s *LogService) GetUsageStats(filters repositories.LogFilters) (*repositories.UsageStats, error) {
	return s.logRepo.GetUsageStats(filters)
}
