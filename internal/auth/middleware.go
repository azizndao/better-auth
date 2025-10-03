// Package auth provides authentication middleware and services.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

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

// RequireAuth middleware that requires authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path should be skipped
		if m.shouldSkipPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.authenticateRequest(r)
		if err != nil {
			m.RespondError(w, err)
			return
		}

		// Add user to request context using transport
		userCtx := &models.AuthData{User: user}
		r = m.SetUserContext(r, userCtx)
		next.ServeHTTP(w, r)
	})
}

// OptionalAuth middleware that optionally authenticates users
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := m.authenticateRequest(r)

		// Add user to request context using transport (can be nil)
		if user != nil {
			userCtx := &models.AuthData{User: user}
			r = m.SetUserContext(r, userCtx)
		}
		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware provides rate limiting
func (m *AuthMiddleware) RateLimitMiddleware(
	requestsPerMinute int,
) func(http.Handler) http.Handler {
	// Simple in-memory rate limiter (in production, use Redis or similar)
	clients := make(map[string][]time.Time)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			now := time.Now()

			// Clean old entries
			if requests, exists := clients[clientIP]; exists {
				validRequests := []time.Time{}
				for _, requestTime := range requests {
					if now.Sub(requestTime) < time.Minute {
						validRequests = append(validRequests, requestTime)
					}
				}
				clients[clientIP] = validRequests
			}

			// Check rate limit
			if len(clients[clientIP]) >= requestsPerMinute {
				m.RespondError(w, transport.NewForbiddenError("Rate limit exceeded", nil))
				return
			}

			// Add current request
			clients[clientIP] = append(clients[clientIP], now)

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware provides CORS support
func (m *AuthMiddleware) CORSMiddleware(
	allowedOrigins []string,
	allowedMethods []string,
	allowedHeaders []string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
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

// GetUserFromContext extracts user from request context
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value("user").(*models.User)
	return user, ok
}
