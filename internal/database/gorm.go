// Package database handles database migrations and connections
package database

import (
	"better-auth/internal/models"
	"better-auth/pkg/plugins/core"

	"gorm.io/gorm"
)

// MigrateCoreModels migrates core and plugin models to the database
func MigrateCoreModels(db *gorm.DB, plugins []core.Plugin) error {
	allModels := []any{
		&models.User{},
		&models.Session{},
		&models.OAuthAccount{},
	}

	// Get all enabled plugins and their models
	for _, plugin := range plugins {
		models := plugin.GetModels()
		allModels = append(allModels, models...)
	}

	if len(allModels) == 0 {
		return nil
	}

	return db.AutoMigrate(allModels...)
}
