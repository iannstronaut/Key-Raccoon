package migrations

import (
	"fmt"

	"gorm.io/gorm"

	"keyraccoon/internal/models"
)

func AddIndexes(db *gorm.DB) error {
	indexes := []struct {
		model interface{}
		name  string
		cols  string
	}{
		{&models.User{}, "idx_users_email", "email"},
		{&models.User{}, "idx_users_role", "role"},
		{&models.User{}, "idx_users_active", "is_active"},
		{&models.Channel{}, "idx_channels_active", "is_active"},
		{&models.APIKey{}, "idx_apikey_key", "key"},
		{&models.APIKey{}, "idx_apikey_user_id", "user_id"},
		{&models.APIKey{}, "idx_apikey_active", "is_active"},
		{&models.ChannelAPIKey{}, "idx_channel_api_key_channel_id", "channel_id"},
		{&models.Model{}, "idx_model_channel_id", "channel_id"},
		{&models.Proxy{}, "idx_proxy_active", "is_active"},
		{&models.Proxy{}, "idx_proxy_status", "status"},
	}

	for _, idx := range indexes {
		if err := db.Migrator().CreateIndex(idx.model, idx.cols); err != nil {
			return fmt.Errorf("create index %s on %s: %w", idx.name, idx.cols, err)
		}
	}

	return nil
}
