package models

import "time"

type RequestLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserAPIKeyID uint      `gorm:"index;not null" json:"user_api_key_id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	ChannelID    uint      `gorm:"index;not null" json:"channel_id"`
	ModelName    string    `gorm:"size:255;not null" json:"model_name"`

	// Denormalized fields for self-contained logs (no joins needed)
	ChannelName string `gorm:"size:255" json:"channel_name"`
	UserEmail   string `gorm:"size:255" json:"user_email"`
	TokenPrice  float64 `gorm:"not null;default:0" json:"token_price"` // price per 1K tokens at time of request

	// Token usage
	InputTokens  int64 `gorm:"not null;default:0" json:"input_tokens"`
	OutputTokens int64 `gorm:"not null;default:0" json:"output_tokens"`
	TotalTokens  int64 `gorm:"not null;default:0" json:"total_tokens"`

	// Cost (calculated: totalTokens / 1000 * tokenPrice)
	Cost float64 `gorm:"not null;default:0" json:"cost"`

	// Status
	Status       string `gorm:"size:50;not null;default:pending;index" json:"status"` // pending, success, failed
	ErrorMessage string `gorm:"type:text" json:"error_message,omitempty"`

	// Timing
	LatencyMs int64 `gorm:"not null;default:0" json:"latency_ms"`

	// Request metadata
	RequestIP string `gorm:"size:100" json:"request_ip"`

	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
