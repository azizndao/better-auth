package database

import (
	"context"
	"fmt"

	"better-auth/internal/models"
	"better-auth/pkg/plugins/core"

	"gorm.io/gorm"
)

// DB implements Database interface using GORM with plugin support
type DB struct {
	db             *gorm.DB
	pluginRegistry *core.PluginRegistry
}

// NewGormDB creates a new GORM-based database with dialector and config
func NewGormDB(dialector gorm.Dialector, config *gorm.Config) (*DB, error) {
	if config == nil {
		config = &gorm.Config{}
	}

	db, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	gormDB := &DB{
		db: db,
	}

	// Migrate core tables
	if err := gormDB.migrateCoreModels(); err != nil {
		return nil, fmt.Errorf("failed to migrate core models: %w", err)
	}

	return gormDB, nil
}

// SetPluginRegistry sets the plugin registry for conditional migrations
func (db *DB) SetPluginRegistry(registry *core.PluginRegistry) {
	db.pluginRegistry = registry
}

// GetGormDB returns the underlying GORM database instance
func (db *DB) GetGormDB() *gorm.DB {
	return db.db
}

// migrateCoreModels migrates the core authentication models
func (db *DB) migrateCoreModels() error {
	coreModels := []any{
		&models.User{},
		&models.Session{},
		&models.OAuthAccount{},
	}

	return db.db.AutoMigrate(coreModels...)
}

// MigratePluginModels migrates models for enabled plugins
func (db *DB) MigratePluginModels() error {
	if db.pluginRegistry == nil {
		return nil
	}

	// Get all enabled plugins and their models
	var allModels []any
	for _, plugin := range db.pluginRegistry.GetAllEnabled() {
		models := plugin.DatabaseModels()
		for _, model := range models {
			allModels = append(allModels, model)
		}
	}

	if len(allModels) == 0 {
		return nil
	}

	return db.db.AutoMigrate(allModels...)
}

// CreateUser creates a new user
func (db *DB) CreateUser(ctx context.Context, user *models.User) error {
	return db.db.WithContext(ctx).Create(user).Error
}

// GetUser retrieves a user by ID
func (db *DB) GetUser(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	if err := db.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := db.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user
func (db *DB) UpdateUser(ctx context.Context, user *models.User) error {
	return db.db.WithContext(ctx).Save(user).Error
}

// DeleteUser deletes a user
func (db *DB) DeleteUser(ctx context.Context, id string) error {
	return db.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}

// CreateSession creates a new session
func (db *DB) CreateSession(ctx context.Context, session *models.Session) error {
	return db.db.WithContext(ctx).Create(session).Error
}

// GetSession retrieves a session by token
func (db *DB) GetSession(ctx context.Context, token string) (*models.Session, error) {
	var session models.Session
	if err := db.db.WithContext(ctx).First(&session, "token = ?", token).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSession updates a session
func (db *DB) UpdateSession(ctx context.Context, session *models.Session) error {
	return db.db.WithContext(ctx).Save(session).Error
}

// DeleteSession deletes a session
func (db *DB) DeleteSession(ctx context.Context, token string) error {
	return db.db.WithContext(ctx).Delete(&models.Session{}, "token = ?", token).Error
}

// CreateOrganization creates a new organization
func (db *DB) CreateOrganization(ctx context.Context, org *models.Organization) error {
	return db.db.WithContext(ctx).Create(org).Error
}

// GetOrganization retrieves an organization by ID
func (db *DB) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	var org models.Organization
	if err := db.db.WithContext(ctx).First(&org, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &org, nil
}

// Transaction wraps operations in a database transaction
func (db *DB) Transaction(fn func(*gorm.DB) error) error {
	return db.db.Transaction(fn)
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Dialector gorm.Dialector `json:"-"`        // GORM dialector (not serializable)
	Config    *gorm.Config   `json:"-"`        // GORM config (not serializable)
	Type      string         `json:"type"`     // Database type for backwards compatibility
	FilePath  string         `json:"filePath"` // For SQLite backwards compatibility
}

// NewGormDBFromConfig creates a new GORM database from configuration
func NewGormDBFromConfig(config *DatabaseConfig) (*DB, error) {
	if config.Dialector == nil {
		return nil, fmt.Errorf("dialector is required")
	}

	return NewGormDB(config.Dialector, config.Config)
}

// HealthCheck performs a database health check
func (db *DB) HealthCheck(ctx context.Context) error {
	sqlDB, err := db.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetStats returns database connection statistics
func (db *DB) GetStats() map[string]any {
	sqlDB, err := db.db.DB()
	if err != nil {
		return map[string]any{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]any{
		"openConnections":   stats.OpenConnections,
		"inUse":             stats.InUse,
		"idle":              stats.Idle,
		"waitCount":         stats.WaitCount,
		"waitDuration":      stats.WaitDuration.String(),
		"maxIdleClosed":     stats.MaxIdleClosed,
		"maxIdleTimeClosed": stats.MaxIdleTimeClosed,
		"maxLifetimeClosed": stats.MaxLifetimeClosed,
	}
}
