package models

import "time"

type Proxy struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	URL       string     `gorm:"type:text;not null" json:"url"`
	Type      string     `gorm:"size:50;not null;default:http" json:"type"`
	Username  string     `gorm:"size:255" json:"username"`
	Password  string     `gorm:"size:255" json:"-"`
	IsActive  bool       `gorm:"not null;default:true" json:"is_active"`
	Status    string     `gorm:"size:50;not null;default:unknown" json:"status"`
	LastCheck *time.Time `json:"last_check"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
