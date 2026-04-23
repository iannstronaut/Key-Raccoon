package migrations

import (
	"gorm.io/gorm"
)

// AddTokenCreditToUserAPIKeys adds token and credit tracking columns to user_api_keys table
func AddTokenCreditToUserAPIKeys(db *gorm.DB) error {
	// Add new columns if they don't exist
	if !db.Migrator().HasColumn(&UserAPIKeyMigration{}, "token_limit") {
		if err := db.Migrator().AddColumn(&UserAPIKeyMigration{}, "token_limit"); err != nil {
			return err
		}
	}
	
	if !db.Migrator().HasColumn(&UserAPIKeyMigration{}, "credit_limit") {
		if err := db.Migrator().AddColumn(&UserAPIKeyMigration{}, "credit_limit"); err != nil {
			return err
		}
	}
	
	if !db.Migrator().HasColumn(&UserAPIKeyMigration{}, "token_used") {
		if err := db.Migrator().AddColumn(&UserAPIKeyMigration{}, "token_used"); err != nil {
			return err
		}
	}
	
	if !db.Migrator().HasColumn(&UserAPIKeyMigration{}, "credit_used") {
		if err := db.Migrator().AddColumn(&UserAPIKeyMigration{}, "credit_used"); err != nil {
			return err
		}
	}
	
	if !db.Migrator().HasColumn(&UserAPIKeyMigration{}, "last_used") {
		if err := db.Migrator().AddColumn(&UserAPIKeyMigration{}, "last_used"); err != nil {
			return err
		}
	}

	return nil
}

// UserAPIKeyMigration is a temporary struct for migration
type UserAPIKeyMigration struct {
	TokenLimit  int64   `gorm:"not null;default:0"`
	CreditLimit float64 `gorm:"not null;default:0"`
	TokenUsed   int64   `gorm:"not null;default:0"`
	CreditUsed  float64 `gorm:"not null;default:0"`
	LastUsed    *string `gorm:"type:timestamp"`
}

func (UserAPIKeyMigration) TableName() string {
	return "user_api_keys"
}
