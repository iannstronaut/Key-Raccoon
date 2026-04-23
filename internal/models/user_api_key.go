package models

import (
	"time"

	"gorm.io/gorm"
)

type UserAPIKey struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"index;not null" json:"user_id"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Key         string         `gorm:"type:text;uniqueIndex;not null" json:"key"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	
	// Token/Credit tracking (compatible with legacy APIKey)
	TokenLimit  int64          `gorm:"not null;default:0" json:"token_limit"`   // 0 = unlimited, -1 = unlimited
	CreditLimit float64        `gorm:"not null;default:0" json:"credit_limit"`  // 0 = unlimited, -1 = unlimited
	TokenUsed   int64          `gorm:"not null;default:0" json:"token_used"`
	CreditUsed  float64        `gorm:"not null;default:0" json:"credit_used"`
	
	// Legacy fields (kept for backward compatibility)
	UsageLimit  int64          `gorm:"default:0" json:"usage_limit"` // 0 = unlimited
	UsageCount  int64          `gorm:"default:0" json:"usage_count"`
	
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time     `json:"last_used_at,omitempty"`
	LastUsed    *time.Time     `json:"last_used,omitempty"` // Alias for LastUsedAt
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User     User                     `json:"user,omitempty"`
	Channels []Channel                `gorm:"many2many:user_api_key_channels;" json:"channels,omitempty"`
	Models   []UserAPIKeyModel        `json:"models,omitempty"`
}

// UserAPIKeyModel represents allowed models for a user API key
type UserAPIKeyModel struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserAPIKeyID uint      `gorm:"index;not null" json:"user_api_key_id"`
	ModelID      uint      `gorm:"index;not null" json:"model_id"`
	CreatedAt    time.Time `json:"created_at"`

	UserAPIKey UserAPIKey `json:"-"`
	Model      Model      `json:"model,omitempty"`
}

// IsExpired checks if the API key has expired
func (k *UserAPIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// IsLimitReached checks if usage limit has been reached
func (k *UserAPIKey) IsLimitReached() bool {
	if k.UsageLimit == 0 {
		return false // unlimited
	}
	return k.UsageCount >= k.UsageLimit
}

// CanUse checks if the API key can be used
func (k *UserAPIKey) CanUse() bool {
	return k.IsActive && !k.IsExpired() && !k.IsLimitReached()
}

// IncrementUsage increments the usage count
func (k *UserAPIKey) IncrementUsage() {
	k.UsageCount++
	now := time.Now()
	k.LastUsedAt = &now
	k.LastUsed = &now
}

// HasTokenAvailable checks if token quota is available
func (k *UserAPIKey) HasTokenAvailable(requiredToken int64) bool {
	if k.TokenLimit == -1 || k.TokenLimit == 0 {
		return true // unlimited
	}
	return (k.TokenUsed + requiredToken) <= k.TokenLimit
}

// HasCreditAvailable checks if credit quota is available
func (k *UserAPIKey) HasCreditAvailable(requiredCredit float64) bool {
	if k.CreditLimit == -1 || k.CreditLimit == 0 {
		return true // unlimited
	}
	return (k.CreditUsed + requiredCredit) <= k.CreditLimit
}
