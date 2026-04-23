package migrations

import (
	"gorm.io/gorm"
)

// CreateUserAPIKeysTable creates user_api_keys and related tables
func CreateUserAPIKeysTable(db *gorm.DB) error {
	// Create user_api_keys table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_api_keys (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			name VARCHAR(255) NOT NULL,
			key TEXT NOT NULL UNIQUE,
			is_active BOOLEAN NOT NULL DEFAULT true,
			usage_limit BIGINT DEFAULT 0,
			usage_count BIGINT DEFAULT 0,
			expires_at TIMESTAMP,
			last_used_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_user_api_keys_user_id ON user_api_keys(user_id);
		CREATE INDEX IF NOT EXISTS idx_user_api_keys_key ON user_api_keys(key);
		CREATE INDEX IF NOT EXISTS idx_user_api_keys_deleted_at ON user_api_keys(deleted_at);
	`).Error; err != nil {
		return err
	}

	// Create user_api_key_channels junction table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_api_key_channels (
			user_api_key_id BIGINT NOT NULL,
			channel_id BIGINT NOT NULL,
			PRIMARY KEY (user_api_key_id, channel_id),
			FOREIGN KEY (user_api_key_id) REFERENCES user_api_keys(id) ON DELETE CASCADE,
			FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
		)
	`).Error; err != nil {
		return err
	}

	// Create user_api_key_models table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_api_key_models (
			id BIGSERIAL PRIMARY KEY,
			user_api_key_id BIGINT NOT NULL,
			model_id BIGINT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_api_key_id) REFERENCES user_api_keys(id) ON DELETE CASCADE,
			FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
			UNIQUE(user_api_key_id, model_id)
		)
	`).Error; err != nil {
		return err
	}

	// Create indexes for junction tables
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_user_api_key_models_user_api_key_id ON user_api_key_models(user_api_key_id);
		CREATE INDEX IF NOT EXISTS idx_user_api_key_models_model_id ON user_api_key_models(model_id);
	`).Error; err != nil {
		return err
	}

	return nil
}
