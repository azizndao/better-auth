package betterauth

import (
	"context"
	"net/http"
	"time"

	"better-auth/internal/auth"
	"better-auth/internal/config"
	"better-auth/internal/database"
	"better-auth/pkg/middleware"

	"gorm.io/gorm"
)

// New creates a new Better Auth instance
func New(cfg *Config) (*BetterAuth, error) {
	// Create database configuration with SQLite as default
	dbConfig := &database.DatabaseConfig{
		Dialector: nil, // Will be set to SQLite default in auth.NewFromDatabaseConfig
		Config:    nil, // Use default GORM config
		Type:      "sqlite",
		FilePath:  "./better-auth.db",
	}

	core, err := auth.NewFromDatabaseConfig(dbConfig, cfg)
	if err != nil {
		return nil, err
	}

	return &BetterAuth{
		core: core,
	}, nil
}

// ServeHTTP implements http.Handler
func (ba *BetterAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ba.core.ServeHTTP(w, r)
}

// Handler returns the http.Handler
func (ba *BetterAuth) Handler() http.Handler {
	return ba.core
}

// RegisterPlugin registers a plugin but doesn't enable it
func (ba *BetterAuth) RegisterPlugin(plugin Plugin) error {
	return ba.core.RegisterPlugin(plugin)
}

// EnablePlugin enables a plugin with configuration
func (ba *BetterAuth) EnablePlugin(
	ctx context.Context,
	pluginName string,
	config map[string]any,
) error {
	return ba.core.EnablePlugin(ctx, pluginName, config)
}

// DisablePlugin disables a plugin
func (ba *BetterAuth) DisablePlugin(pluginName string) error {
	return ba.core.DisablePlugin(pluginName)
}

// Use adds a plugin to the authentication system (deprecated - use RegisterPlugin and EnablePlugin)
func (ba *BetterAuth) Use(plugin Plugin) error {
	if err := ba.core.RegisterPlugin(plugin); err != nil {
		return err
	}
	return ba.core.EnablePlugin(context.Background(), plugin.Name(), map[string]any{
		"enabled": true,
	})
}

// SetTransport sets a custom transport
func (ba *BetterAuth) SetTransport(t Transport) {
	ba.core.SetTransport(t)
}

// GetTransport returns the current transport
func (ba *BetterAuth) GetTransport() Transport {
	return ba.core.GetTransport()
}

// Middleware returns the middleware manager
func (ba *BetterAuth) Middleware() *middleware.Middleware {
	return middleware.New(ba.core.Middleware())
}

// GetUser retrieves a user by ID
func (ba *BetterAuth) GetUser(ctx context.Context, userID string) (*User, error) {
	return ba.core.GetUser(ctx, userID)
}

// GetUserByEmail retrieves a user by email
func (ba *BetterAuth) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return ba.core.GetUserByEmail(ctx, email)
}

// ValidateJWT validates a JWT token and returns claims
func (ba *BetterAuth) ValidateJWT(tokenString string) (*auth.JWTClaims, error) {
	return ba.core.ValidateJWT(tokenString)
}

// GenerateJWT generates a JWT token for a user
func (ba *BetterAuth) GenerateJWT(user *User) (string, error) {
	return ba.core.GenerateJWT(user)
}

// CreateSession creates a new session for a user
func (ba *BetterAuth) CreateSession(
	ctx context.Context,
	userID, ipAddress, userAgent string,
) (*Session, error) {
	return ba.core.CreateSession(ctx, userID, ipAddress, userAgent)
}

// ValidateSession validates a session token
func (ba *BetterAuth) ValidateSession(ctx context.Context, token string) (*Session, error) {
	return ba.core.ValidateSession(ctx, token)
}

// DeleteSession deletes a session
func (ba *BetterAuth) DeleteSession(ctx context.Context, token string) error {
	return ba.core.DeleteSession(ctx, token)
}

// GetDatabase returns the database instance
func (ba *BetterAuth) GetDatabase() *database.DB {
	return ba.core.GetGormDB()
}

// Convenience functions for common configurations
func DefaultCORSConfig() *config.CORSConfig {
	return config.DefaultCORSConfig()
}

func DefaultConfig() *Config {
	return &Config{
		PathPrefix:    "/auth",
		SessionExpiry: 24 * time.Hour,
		JWTExpiry:     1 * time.Hour,
		CORSConfig:    DefaultCORSConfig(),
	}
}

// NewDatabaseConfig creates a new database configuration with GORM dialector
func NewDatabaseConfig(dialector gorm.Dialector, config *gorm.Config) *database.DatabaseConfig {
	return &database.DatabaseConfig{
		Dialector: dialector,
		Config:    config,
	}
}

// NewWithDatabase creates a new Better Auth instance with custom database
func NewWithDatabase(cfg *Config, dbConfig *database.DatabaseConfig) (*BetterAuth, error) {
	core, err := auth.NewFromDatabaseConfig(dbConfig, cfg)
	if err != nil {
		return nil, err
	}

	return &BetterAuth{
		core: core,
	}, nil
}
