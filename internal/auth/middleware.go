package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"better-auth/internal/models"
	"better-auth/internal/transport"
)

// MiddlewareConfig configures authentication middleware
type MiddlewareConfig struct {
	SessionService *SessionService
	JWTService     *JWTService
	Transport      transport.Transport
	SkipPaths      []string
	Optional       bool
}

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	config *MiddlewareConfig
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config *MiddlewareConfig) *AuthMiddleware {
	return &AuthMiddleware{
		config: config,
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
			m.config.Transport.WriteError(w, http.StatusUnauthorized, err.Error())
			return
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth middleware that optionally authenticates users
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := m.authenticateRequest(r)
		
		// Add user to request context (can be nil)
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SessionAuth middleware that uses session-based authentication
func (m *AuthMiddleware) SessionAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.config.SessionService == nil {
			m.config.Transport.WriteError(w, http.StatusInternalServerError, "session service not configured")
			return
		}

		// Get session token from request
		token := m.config.SessionService.GetSessionFromRequest(r)
		if token == "" {
			if !m.config.Optional {
				m.config.Transport.WriteError(w, http.StatusUnauthorized, "session token required")
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// Validate session
		session, err := m.config.SessionService.ValidateSession(r.Context(), token)
		if err != nil {
			if !m.config.Optional {
				m.config.Transport.WriteError(w, http.StatusUnauthorized, "invalid session")
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// Add session to request context
		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// JWTAuth middleware that uses JWT-based authentication
func (m *AuthMiddleware) JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.config.JWTService == nil {
			m.config.Transport.WriteError(w, http.StatusInternalServerError, "JWT service not configured")
			return
		}

		// Get JWT token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			if !m.config.Optional {
				m.config.Transport.WriteError(w, http.StatusUnauthorized, "authorization header required")
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from "Bearer <token>" format
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			if !m.config.Optional {
				m.config.Transport.WriteError(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		token := authHeader[len(bearerPrefix):]

		// Validate JWT token
		claims, err := m.config.JWTService.ValidateToken(token)
		if err != nil {
			if !m.config.Optional {
				m.config.Transport.WriteError(w, http.StatusUnauthorized, "invalid JWT token")
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), "jwt_claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleAuth middleware that requires specific roles
func (m *AuthMiddleware) RoleAuth(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First authenticate the user
			user, err := m.authenticateRequest(r)
			if err != nil {
				m.config.Transport.WriteError(w, http.StatusUnauthorized, err.Error())
				return
			}

			// Check if user has required role
			userRole := ""
			if user.Metadata != nil {
				if role, ok := user.Metadata["role"].(string); ok {
					userRole = role
				}
			}

			hasRole := false
			for _, role := range requiredRoles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				m.config.Transport.WriteError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			// Add user to request context
			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RateLimitMiddleware provides rate limiting
func (m *AuthMiddleware) RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
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
				m.config.Transport.WriteError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			
			// Add current request
			clients[clientIP] = append(clients[clientIP], now)
			
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware provides CORS support
func (m *AuthMiddleware) CORSMiddleware(allowedOrigins []string, allowedMethods []string, allowedHeaders []string) func(http.Handler) http.Handler {
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
	// Try session authentication first
	if m.config.SessionService != nil {
		if token := m.config.SessionService.GetSessionFromRequest(r); token != "" {
			if session, err := m.config.SessionService.ValidateSession(r.Context(), token); err == nil {
				// Need to get user from session - this would require a user service
				// For now, return a placeholder
				return &models.User{ID: session.UserID}, nil
			}
		}
	}

	// Try JWT authentication
	if m.config.JWTService != nil {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := authHeader[7:]
			if claims, err := m.config.JWTService.ValidateToken(token); err == nil {
				return &models.User{
					ID:       claims.UserID,
					Email:    claims.Email,
					Name:     claims.Name,
					Metadata: claims.Metadata,
				}, nil
			}
		}
	}

	return nil, http.ErrNotSupported
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

// GetSessionFromContext extracts session from request context
func GetSessionFromContext(ctx context.Context) (*models.Session, bool) {
	session, ok := ctx.Value("session").(*models.Session)
	return session, ok
}

// GetJWTClaimsFromContext extracts JWT claims from request context
func GetJWTClaimsFromContext(ctx context.Context) (*JWTClaims, bool) {
	claims, ok := ctx.Value("jwt_claims").(*JWTClaims)
	return claims, ok
}