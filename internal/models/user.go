package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Email       string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Password    string         `gorm:"size:255;not null" json:"-"`
	Name        string         `gorm:"size:255" json:"name"`
	Role        string         `gorm:"size:50;not null;default:user" json:"role"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	TokenLimit  int64          `gorm:"not null;default:0" json:"token_limit"`
	CreditLimit float64        `gorm:"not null;default:0" json:"credit_limit"`
	TokenUsed   int64          `gorm:"not null;default:0" json:"token_used"`
	CreditUsed  float64        `gorm:"not null;default:0" json:"credit_used"`
	LastLogin   *time.Time     `json:"last_login"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Channels []Channel `gorm:"many2many:user_channels;" json:"channels,omitempty"`
	APIKeys  []APIKey  `json:"api_keys,omitempty"`
}
