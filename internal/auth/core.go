package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"better-auth/internal/config"
	"better-auth/internal/models"
	"better-auth/pkg/plugins/core"
	"better-auth/pkg/router"
	"better-auth/pkg/transport"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthCore is the main authentication engine
type AuthCore struct {
	config     *config.Config
	database   *gorm.DB
	plugins    []core.Plugin // List of plugins for extensibility
	transport  transport.Transport
	router     router.Router
	jwt        *JWTService
	oauth      *OAuthService
	session    *SessionService
	middleware *AuthMiddleware
}

// New creates a new authentication system with plugin support
func New(cfg *config.Config, db *gorm.DB, plugins []core.Plugin) (*AuthCore, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Create router with options suitable for auth
	routerOptions := router.RouterOptions{
		AutoOPTIONS:           true,
		AutoHEAD:              true,
		TrailingSlashRedirect: false, // Don't redirect for auth endpoints
		MethodNotAllowed:      true,
		EnableLogging:         false, // We'll handle logging separately
		EnableRecovery:        true,
	}

	auth := &AuthCore{
		config:    cfg,
		database:  db,
		plugins:   plugins,
		router:    router.NewRouterWithOptions(routerOptions),
		transport: transport.NewDefault(),
	}

	auth.registerRoutes()

	for _, p := range plugins {
		if p == nil {
			continue
		}

		// Initialize plugin with the database
		if err := p.Initialize(context.Background(), db); err != nil {
			return nil, fmt.Errorf("failed to initialize plugin %s: %w", p.Name(), err)
		}

		auth.plugins = append(auth.plugins, p)

		// Register plugin routes if available
		if err := p.RegisterRoutes(auth.router); err != nil {
			return nil, fmt.Errorf("failed to register routes for plugin %s: %w", p.Name(), err)
		}
	}

	return auth, nil
}

// ServeHTTP implements http.Handler
func (c *AuthCore) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply global middleware that isn't handled by router
	r = c.applyRateLimit(w, r)
	if r == nil {
		return
	}

	c.router.ServeHTTP(w, r)
}

// SetTransport sets a custom transport
func (c *AuthCore) SetTransport(t transport.Transport) {
	c.transport = t
}

// GetTransport returns the current transport
func (c *AuthCore) GetTransport() transport.Transport {
	return c.transport
}

// GetTransportInterface returns the transport as any for plugin system
func (c *AuthCore) GetTransportInterface() any {
	return c.transport
}

// GetGormDB returns the GORM database instance
func (c *AuthCore) GetGormDB() *gorm.DB {
	return c.database
}

// GetConfig returns the configuration
func (c *AuthCore) GetConfig() *config.Config {
	return c.config
}

func (c *AuthCore) registerRoutes() {
	// Create route group with the configured prefix
	authGroup := c.router.Group(c.config.PathPrefix)

	// Add CORS middleware if configured
	if c.config.CORSConfig != nil {
		corsOptions := router.CORSOptions{
			AllowedOrigins:   c.config.CORSConfig.AllowedOrigins,
			AllowedMethods:   c.config.CORSConfig.AllowedMethods,
			AllowedHeaders:   c.config.CORSConfig.AllowedHeaders,
			AllowCredentials: c.config.CORSConfig.AllowCredentials,
			MaxAge:           24 * time.Hour, // Default 24 hours
		}
		authGroup.Use(router.CORS(corsOptions))
	}

	handler := newHandlers(c)
	handler.registerRoutes(authGroup)
}

func (c *AuthCore) applyRateLimit(_ http.ResponseWriter, r *http.Request) *http.Request {
	if !c.config.RateLimitEnabled {
		return r
	}
	// Rate limiting logic would be implemented here
	return r
}

// GetUser retrieves a user by ID
func (c *AuthCore) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := c.database.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (c *AuthCore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := c.database.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ValidateJWT validates a JWT token and returns claims
func (c *AuthCore) ValidateJWT(tokenString string) (*JWTClaims, error) {
	return c.jwt.ValidateToken(tokenString)
}

// GenerateJWT generates a JWT token for a user
func (c *AuthCore) GenerateJWT(user *models.User) (string, error) {
	return c.jwt.GenerateToken(user)
}

// CreateSession creates a new session for a user
func (c *AuthCore) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	ipAddress, userAgent string,
) (*models.Session, error) {
	return c.session.CreateSession(ctx, userID, ipAddress, userAgent)
}

// ValidateSession validates a session token
func (c *AuthCore) ValidateSession(ctx context.Context, token string) (*models.Session, error) {
	return c.session.ValidateSession(ctx, token)
}

// DeleteSession deletes a session
func (c *AuthCore) DeleteSession(ctx context.Context, token string) error {
	return c.session.DeleteSession(ctx, token)
}

// Middleware returns the middleware manager
func (c *AuthCore) Middleware() *AuthMiddleware {
	return c.middleware
}
