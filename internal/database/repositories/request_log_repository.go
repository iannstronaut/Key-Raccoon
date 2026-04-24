package repositories

import (
	"time"

	"gorm.io/gorm"

	"keyraccoon/internal/models"
)

type RequestLogRepository struct {
	db *gorm.DB
}

func NewRequestLogRepository(db *gorm.DB) *RequestLogRepository {
	return &RequestLogRepository{db: db}
}

func (r *RequestLogRepository) Create(log *models.RequestLog) error {
	return r.db.Create(log).Error
}

func (r *RequestLogRepository) BulkCreate(logs []models.RequestLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.CreateInBatches(logs, 100).Error
}

func (r *RequestLogRepository) GetByAPIKeyID(keyID uint, limit, offset int) ([]models.RequestLog, int64, error) {
	var logs []models.RequestLog
	var total int64

	query := r.db.Model(&models.RequestLog{}).Where("user_api_key_id = ?", keyID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (r *RequestLogRepository) GetByUserID(userID uint, limit, offset int) ([]models.RequestLog, int64, error) {
	var logs []models.RequestLog
	var total int64

	query := r.db.Model(&models.RequestLog{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (r *RequestLogRepository) GetByChannelID(channelID uint, limit, offset int) ([]models.RequestLog, int64, error) {
	var logs []models.RequestLog
	var total int64

	query := r.db.Model(&models.RequestLog{}).Where("channel_id = ?", channelID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

type LogFilters struct {
	Status    string
	Model     string
	ChannelID uint
	UserID    uint
	APIKeyID  uint
	DateFrom  *time.Time
	DateTo    *time.Time
}

func (r *RequestLogRepository) GetAll(limit, offset int, filters LogFilters) ([]models.RequestLog, int64, error) {
	var logs []models.RequestLog
	var total int64

	query := r.db.Model(&models.RequestLog{})

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Model != "" {
		query = query.Where("model_name = ?", filters.Model)
	}
	if filters.ChannelID != 0 {
		query = query.Where("channel_id = ?", filters.ChannelID)
	}
	if filters.UserID != 0 {
		query = query.Where("user_id = ?", filters.UserID)
	}
	if filters.APIKeyID != 0 {
		query = query.Where("user_api_key_id = ?", filters.APIKeyID)
	}
	if filters.DateFrom != nil {
		query = query.Where("created_at >= ?", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		query = query.Where("created_at <= ?", *filters.DateTo)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return logs, total, err
}

type UsageStats struct {
	TotalRequests  int64   `json:"total_requests"`
	TotalTokens    int64   `json:"total_tokens"`
	TotalCost      float64 `json:"total_cost"`
	SuccessCount   int64   `json:"success_count"`
	FailedCount    int64   `json:"failed_count"`
	AvgLatencyMs   float64 `json:"avg_latency_ms"`
}

func (r *RequestLogRepository) GetUsageStats(filters LogFilters) (*UsageStats, error) {
	var stats UsageStats

	query := r.db.Model(&models.RequestLog{})

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.ChannelID != 0 {
		query = query.Where("channel_id = ?", filters.ChannelID)
	}
	if filters.UserID != 0 {
		query = query.Where("user_id = ?", filters.UserID)
	}
	if filters.DateFrom != nil {
		query = query.Where("created_at >= ?", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		query = query.Where("created_at <= ?", *filters.DateTo)
	}

	err := query.Select(`
		COUNT(*) as total_requests,
		COALESCE(SUM(total_tokens), 0) as total_tokens,
		COALESCE(SUM(cost), 0) as total_cost,
		COALESCE(SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END), 0) as success_count,
		COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0) as failed_count,
		COALESCE(AVG(latency_ms), 0) as avg_latency_ms
	`).Scan(&stats).Error

	return &stats, err
}
