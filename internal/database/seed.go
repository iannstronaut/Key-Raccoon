package database

import (
	"errors"

	"gorm.io/gorm"

	"keyraccoon/internal/models"
	"keyraccoon/internal/utils"
)

func Seed(db *gorm.DB, adminEmail, adminPassword string) error {
	var existing models.User
	err := db.Where("email = ?", adminEmail).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	password, err := utils.HashPassword(adminPassword)
	if err != nil {
		return err
	}

	return db.Create(&models.User{
		Email:       adminEmail,
		Password:    password,
		Name:        "Default Superadmin",
		Role:        "superadmin",
		IsActive:    true,
		TokenLimit:  -1,
		CreditLimit: -1,
	}).Error
}
