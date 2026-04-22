package models

import (
	"time"

	"gorm.io/gorm"
)

type Channel struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Type        string         `gorm:"size:100;not null;default:openai" json:"type"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	APIKeys []ChannelAPIKey `json:"api_keys,omitempty"`
	Models  []Model         `json:"models,omitempty"`
	Users   []User          `gorm:"many2many:user_channels;" json:"users,omitempty"`
}

type ChannelAPIKey struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChannelID uint      `gorm:"index;not null" json:"channel_id"`
	APIKey    string    `gorm:"type:text;not null" json:"api_key"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Channel Channel `json:"channel,omitempty"`
}

type Model struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ChannelID    uint      `gorm:"index;not null" json:"channel_id"`
	Name         string    `gorm:"size:255;not null" json:"name"`
	DisplayName  string    `gorm:"size:255" json:"display_name"`
	IsActive     bool      `gorm:"not null;default:true" json:"is_active"`
	TokenPrice   float64   `gorm:"not null;default:0" json:"token_price"`
	SystemPrompt string    `gorm:"type:text" json:"system_prompt"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Channel Channel `json:"channel,omitempty"`
}
