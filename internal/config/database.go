package config

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"keyraccoon/internal/database"
	"keyraccoon/internal/models"
	"keyraccoon/pkg/logger"
)

var db *gorm.DB

func InitDatabase(cfg *Config) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	databaseConnection, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := databaseConnection.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	if err := databaseConnection.AutoMigrate(
		&models.User{},
		&models.Channel{},
		&models.ChannelAPIKey{},
		&models.Model{},
		&models.Proxy{},
		&models.UserAPIKey{},
		&models.UserAPIKeyModel{},
		&models.RequestLog{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	if err := database.Seed(databaseConnection, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	db = databaseConnection
	logger.Info("database connected and migrated")
	return nil
}

func GetDB() *gorm.DB {
	return db
}
