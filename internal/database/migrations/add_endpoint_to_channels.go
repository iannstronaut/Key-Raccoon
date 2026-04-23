package migrations

import (
	"gorm.io/gorm"
)

// AddEndpointToChannels adds endpoint column to channels table
func AddEndpointToChannels(db *gorm.DB) error {
	return db.Exec(`
		ALTER TABLE channels 
		ADD COLUMN IF NOT EXISTS endpoint VARCHAR(500) DEFAULT NULL
	`).Error
}
