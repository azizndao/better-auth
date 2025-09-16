// Package middleware provides middleware utilities for better-auth
package middleware

import (
	"net/http"

	"better-auth/internal/auth"
)

// Middleware provides public middleware utilities for better-auth
type Middleware struct {
	authMiddleware *auth.AuthMiddleware
}

// New creates a new middleware instance
func New(authMiddleware *auth.AuthMiddleware) *Middleware {
	return &Middleware{
		authMiddleware: authMiddleware,
	}
}

// RequireAuth returns middleware that requires authentication
func (m *Middleware) RequireAuth() func(http.Handler) http.Handler {
	return m.authMiddleware.RequireAuth
}

// OptionalAuth returns middleware that optionally authenticates users
func (m *Middleware) OptionalAuth() func(http.Handler) http.Handler {
	return m.authMiddleware.OptionalAuth
}

// SessionAuth returns middleware that uses session-based authentication
func (m *Middleware) SessionAuth() func(http.Handler) http.Handler {
	return m.authMiddleware.SessionAuth
}

// JWTAuth returns middleware that uses JWT-based authentication
func (m *Middleware) JWTAuth() func(http.Handler) http.Handler {
	return m.authMiddleware.JWTAuth
}

// RoleAuth returns middleware that requires specific roles
func (m *Middleware) RoleAuth(roles ...string) func(http.Handler) http.Handler {
	return m.authMiddleware.RoleAuth(roles...)
}

// RateLimit returns middleware that applies rate limiting
func (m *Middleware) RateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	return m.authMiddleware.RateLimitMiddleware(requestsPerMinute)
}

// CORS returns middleware that handles CORS
func (m *Middleware) CORS(origins, methods, headers []string) func(http.Handler) http.Handler {
	return m.authMiddleware.CORSMiddleware(origins, methods, headers)
}

// Common middleware configurations
var (
	// DefaultCORSOrigins provides default CORS origins
	DefaultCORSOrigins = []string{"http://localhost:3000", "http://localhost:8080"}

	// DefaultCORSMethods provides default CORS methods
	DefaultCORSMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}

	// DefaultCORSHeaders provides default CORS headers
	DefaultCORSHeaders = []string{"Content-Type", "Authorization", "X-Requested-With"}
)

