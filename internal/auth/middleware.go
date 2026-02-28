// Package auth provides authentication middleware and services.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"better-auth/internal/dto"
	"better-auth/internal/models"
	"better-auth/pkg/plugins/core"
	"better-auth/pkg/transport"
)

// MiddlewareConfig configures authentication middleware
type MiddlewareConfig struct {
	SkipPaths []string
	Optional  bool
}

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	transport.Transport
	sessionService *SessionService
	config         *MiddlewareConfig
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config *MiddlewareConfig, transport transport.Transport, sessionService *SessionService) *AuthMiddleware {
	return &AuthMiddleware{
		Transport:      transport,
		sessionService: sessionService,
		config:         config,
	}
}

// OptionalAuth middleware that optionally authenticates users
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := m.authenticateRequest(r)

		// Add user to request context using transport (can be nil)
		if user != nil {
			userCtx := &dto.AuthData{User: dto.NewUser(*user)}
			r = SetUserContext(r, userCtx)
		}
		next.ServeHTTP(w, r)
	})
}

// authenticateRequest attempts to authenticate a request using available methods
func (m *AuthMiddleware) authenticateRequest(r *http.Request) (*models.User, error) {
	// Extract token using transport (supports Authorization header, cookies, query params)
	token := m.ExtractToken(r)
	if token == "" {
		return nil, fmt.Errorf("no authentication token found")
	}

	if session, err := m.sessionService.ValidateSession(r.Context(), token); err == nil {
		// Need to get user from session - this would require a user service
		// For now, return a placeholder
		baseModel, err := core.NewModelFromID(session.UserID)
		if err != nil {
			return nil, err
		}
		return &models.User{Model: *baseModel}, nil
	}

	return nil, fmt.Errorf("authentication failed")
}

// shouldSkipPath checks if the path should skip authentication
func (m *AuthMiddleware) shouldSkipPath(path string) bool {
	for _, skipPath := range m.config.SkipPaths {
		if path == skipPath || strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

type contextKey string

const UserContextKey contextKey = "betterauth.user"

// GetUserContext retrieves user context from the request
func GetUserContext(req *http.Request) (*dto.AuthData, bool) {
	ctx := req.Context().Value(UserContextKey)
	if ctx == nil {
		return nil, false
	}
	userCtx, ok := ctx.(*dto.AuthData)
	return userCtx, ok
}

// SetUserContext sets user context in the request
func SetUserContext(req *http.Request, ctx *dto.AuthData) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), UserContextKey, ctx))
}
