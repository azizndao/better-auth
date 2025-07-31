package betterauth

import (
	"better-auth/internal/auth"
	"better-auth/internal/config"
	"better-auth/internal/database"
	"better-auth/pkg/plugins/core"
	"better-auth/pkg/transport"
)

// BetterAuth is the main authentication system
type BetterAuth struct {
	core *auth.AuthCore
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
type Plugin = core.Plugin

// Transport interface for customizing request/response handling
type Transport = transport.Transport

// Database interface for database operations
// Database is the GORM database instance
type Database = database.DB
