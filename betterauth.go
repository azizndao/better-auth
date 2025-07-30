package betterauth

import (
	"context"
	"net/http"
	"time"

	"better-auth/internal/auth"
	"better-auth/internal/config"
	"better-auth/internal/database"
	"better-auth/internal/transport"
	"better-auth/pkg/middleware"
	"better-auth/pkg/plugins"
)

// BetterAuth is the main authentication system
type BetterAuth struct {
	core *auth.BetterAuth
}

// Config holds the configuration for Better Auth
type Config = config.Config

// User represents an authenticated user
type User = auth.User

// Session represents a user session
type Session = auth.Session

// Organization represents an organization
type Organization = auth.Organization

// UserContext contains user information in request context
type UserContext = auth.UserContext

// Plugin interface for extending functionality
type Plugin = plugins.Plugin

// Transport interface for customizing request/response handling
type Transport = transport.Transport

// New creates a new Better Auth instance
func New(cfg *Config) (*BetterAuth, error) {
	core, err := auth.New(cfg)
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

// Use adds a plugin to the authentication system
func (ba *BetterAuth) Use(plugin Plugin) error {
	return ba.core.Use(plugin)
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
func (ba *BetterAuth) CreateSession(ctx context.Context, userID, ipAddress, userAgent string) (*Session, error) {
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
func (ba *BetterAuth) GetDatabase() Database {
	return ba.core.GetDatabase()
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

// Database types for convenience
type Database = database.Database

// Database constructors
func NewPostgresDB(connectionURL string) Database {
	return database.NewPostgresDB(connectionURL)
}

func NewMySQLDB(connectionURL string) Database {
	return database.NewMySQLDB(connectionURL)
}

func NewSQLiteDB(connectionURL string) Database {
	return database.NewSQLiteDB(connectionURL)
}

func NewInMemoryDB() Database {
	return database.NewInMemoryDB()
}