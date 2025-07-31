package jwt

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	"better-auth/pkg/plugins/core"
	"better-auth/pkg/transport"

	"gorm.io/gorm"
)

// JWTPlugin provides JWT token management functionality
type JWTPlugin struct {
	*core.BasePlugin
	service   *JWTService
	handlers  *JWTHandlers
	config    *JWTConfig
	transport transport.Transport
}

// NewJWTPlugin creates a new JWT plugin
func NewJWTPlugin(config *JWTConfig) *JWTPlugin {
	if config == nil {
		config = &JWTConfig{
			AccessTokenExpiry:       15 * time.Minute,
			RefreshTokenExpiry:      7 * 24 * time.Hour,
			Issuer:                  "better-auth",
			EnableRefreshTokens:     true,
			EnableTokenBlacklist:    true,
			CleanupInterval:         1 * time.Hour,
			MaxRefreshTokensPerUser: 5,
		}
	}

	plugin := &JWTPlugin{
		BasePlugin: core.NewBasePlugin("jwt"),
		config:     config,
	}

	// Add database models
	plugin.AddModel(&JWTRefreshToken{})
	plugin.AddModel(&JWTBlacklist{})

	// Set required configuration
	plugin.SetRequiredConfig(map[string]any{
		"enabled":              true,
		"secretKey":            nil, // Will be validated
		"accessTokenExpiry":    config.AccessTokenExpiry,
		"refreshTokenExpiry":   config.RefreshTokenExpiry,
		"enableRefreshTokens":  config.EnableRefreshTokens,
		"enableTokenBlacklist": config.EnableTokenBlacklist,
	})

	return plugin
}

// SetTransport sets the transport for the plugin
func (p *JWTPlugin) SetTransport(transportInterface any) {
	if t, ok := transportInterface.(transport.Transport); ok {
		p.transport = t
	}
}

// Initialize initializes the JWT plugin
func (p *JWTPlugin) Initialize(ctx context.Context, db *gorm.DB) error {
	// Initialize service
	p.service = NewJWTService(db, p.config)
	
	// Initialize handlers with transport (use default if not set)
	if p.transport == nil {
		p.transport = transport.NewDefault()
	}
	p.handlers = NewJWTHandlers(p.service, p.transport)

	// Add service to plugin
	p.AddService(p.service)

	// Setup routes
	p.setupRoutes()

	// Call base initialization
	return p.BasePlugin.Initialize(ctx, db)
}

// setupRoutes configures the plugin routes
func (p *JWTPlugin) setupRoutes() {
	// Token management routes
	p.AddRoute("POST /jwt/refresh", http.HandlerFunc(p.handlers.RefreshToken))
	p.AddRoute("POST /jwt/revoke", http.HandlerFunc(p.handlers.RevokeToken))
	p.AddRoute("POST /jwt/validate", http.HandlerFunc(p.handlers.ValidateToken))
	p.AddRoute("GET /jwt/info", http.HandlerFunc(p.handlers.GetTokenInfo))
}

// ValidateConfig validates the plugin configuration
func (p *JWTPlugin) ValidateConfig(config map[string]any) error {
	enabled, ok := config["enabled"]
	if !ok {
		return fmt.Errorf("JWT plugin requires 'enabled' configuration")
	}

	if enabledBool, ok := enabled.(bool); !ok || !enabledBool {
		return fmt.Errorf("JWT plugin must be enabled")
	}

	secretKey, ok := config["secretKey"]
	if !ok || secretKey == nil {
		return fmt.Errorf("JWT plugin requires 'secretKey' configuration")
	}

	switch key := secretKey.(type) {
	case string:
		if len(key) < 32 {
			return fmt.Errorf("JWT secret key must be at least 32 characters long")
		}
		p.config.SecretKey = []byte(key)
	case []byte:
		if len(key) < 32 {
			return fmt.Errorf("JWT secret key must be at least 32 bytes long")
		}
		p.config.SecretKey = key
	default:
		return fmt.Errorf("JWT secret key must be a string or byte slice")
	}

	// Update config with provided values
	if val, ok := config["accessTokenExpiry"]; ok {
		if duration, ok := val.(time.Duration); ok {
			p.config.AccessTokenExpiry = duration
		}
	}

	if val, ok := config["refreshTokenExpiry"]; ok {
		if duration, ok := val.(time.Duration); ok {
			p.config.RefreshTokenExpiry = duration
		}
	}

	if val, ok := config["issuer"]; ok {
		if issuer, ok := val.(string); ok {
			p.config.Issuer = issuer
		}
	}

	if val, ok := config["audience"]; ok {
		if audience, ok := val.(string); ok {
			p.config.Audience = audience
		}
	}

	if val, ok := config["enableRefreshTokens"]; ok {
		if enabled, ok := val.(bool); ok {
			p.config.EnableRefreshTokens = enabled
		}
	}

	if val, ok := config["enableTokenBlacklist"]; ok {
		if enabled, ok := val.(bool); ok {
			p.config.EnableTokenBlacklist = enabled
		}
	}

	if val, ok := config["cleanupInterval"]; ok {
		if interval, ok := val.(time.Duration); ok {
			p.config.CleanupInterval = interval
		}
	}

	if val, ok := config["maxRefreshTokensPerUser"]; ok {
		if max, ok := val.(int); ok {
			p.config.MaxRefreshTokensPerUser = max
		}
	}

	return nil
}

// OnEnabled is called when the plugin is enabled
func (p *JWTPlugin) OnEnabled() error {
	return p.BasePlugin.OnEnabled()
}

// OnDisabled is called when the plugin is disabled
func (p *JWTPlugin) OnDisabled() error {
	return p.BasePlugin.OnDisabled()
}

// GetService returns the JWT service
func (p *JWTPlugin) GetService() *JWTService {
	return p.service
}

// GetHandlers returns the JWT handlers
func (p *JWTPlugin) GetHandlers() *JWTHandlers {
	return p.handlers
}

// GetConfig returns the JWT configuration
func (p *JWTPlugin) GetConfig() *JWTConfig {
	return p.config
}

// Middleware provides JWT authentication middleware
func (p *JWTPlugin) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip JWT validation for certain paths
		skipPaths := []string{
			"/auth/sign-in",
			"/auth/sign-up",
			"/auth/reset-password",
			"/jwt/refresh",
			"/jwt/validate",
		}

		if slices.Contains(skipPaths, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract JWT token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		const bearerPrefix = "Bearer "
		if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			next.ServeHTTP(w, r)
			return
		}

		tokenString := authHeader[len(bearerPrefix):]
		claims, err := p.service.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), "claims", claims)
		ctx = context.WithValue(ctx, "userId", claims.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

