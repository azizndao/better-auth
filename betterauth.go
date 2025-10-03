// Package betterauth provides a simplified interface for the Better Auth authentication system.
package betterauth

import (
	"context"
	"net/http"
	"time"

	"better-auth/internal/auth"
	"better-auth/internal/config"
	"better-auth/pkg/plugins/core"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// New creates a new Better Auth instance
func New(
	cfg *Config,
	db *gorm.DB,
	plugins []core.Plugin,
) (*BetterAuth, error) {
	core, err := auth.New(cfg, db, plugins)
	if err != nil {
		return nil, err
	}

	return &BetterAuth{core: core}, nil
}

// ServeHTTP implements http.Handler
func (ba *BetterAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ba.core.ServeHTTP(w, r)
}

// Handler returns the http.Handler
func (ba *BetterAuth) Handler() http.Handler {
	return ba.core
}

// GetUser retrieves a user by ID
func (ba *BetterAuth) GetUser(ctx context.Context, userID uuid.UUID) (*User, error) {
	return ba.core.GetUser(ctx, userID)
}

// GetUserByEmail retrieves a user by email
func (ba *BetterAuth) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return ba.core.GetUserByEmail(ctx, email)
}

// CreateSession creates a new session for a user
func (ba *BetterAuth) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	ipAddress, userAgent string,
	tx *gorm.DB,
) (*Session, error) {
	return ba.core.CreateSession(ctx, userID, ipAddress, userAgent, tx)
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
func (ba *BetterAuth) GetDatabase() *gorm.DB {
	return ba.core.GetGormDB()
}

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
