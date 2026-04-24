package main

import (
	"fmt"
	"log"

	"keyraccoon/internal/config"
	"keyraccoon/internal/models"
)

func main() {
	// Load config
	cfg, err := config.Init()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize database
	if err := config.InitDatabase(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	db := config.GetDB()

	// Run migration for UserAPIKey
	fmt.Println("Running migration for UserAPIKey...")
	if err := db.AutoMigrate(&models.UserAPIKey{}, &models.UserAPIKeyModel{}); err != nil {
		log.Fatal("Failed to migrate UserAPIKey:", err)
	}

	fmt.Println("✅ Migration completed successfully!")
	fmt.Println("Tables created:")
	fmt.Println("  - user_api_keys")
	fmt.Println("  - user_api_key_models")
	fmt.Println("  - user_api_key_channels (junction table)")
}
