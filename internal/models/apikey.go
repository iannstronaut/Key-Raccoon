package models

import "time"

type APIKey struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"index;not null" json:"user_id"`
	Key         string     `gorm:"uniqueIndex;size:255;not null" json:"key"`
	Name        string     `gorm:"size:255" json:"name"`
	TokenLimit  int64      `gorm:"not null;default:0" json:"token_limit"`
	CreditLimit float64    `gorm:"not null;default:0" json:"credit_limit"`
	TokenUsed   int64      `gorm:"not null;default:0" json:"token_used"`
	CreditUsed  float64    `gorm:"not null;default:0" json:"credit_used"`
	IsActive    bool       `gorm:"not null;default:true" json:"is_active"`
	LastUsed    *time.Time `json:"last_used"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	User     User      `json:"user,omitempty"`
	Channels []Channel `gorm:"many2many:apikey_channels;" json:"channels,omitempty"`
}
