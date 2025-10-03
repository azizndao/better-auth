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
	plugins    []core.Plugin
	transport  transport.Transport
	router     router.Router
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
		EnableLogging:         false, // We'll handle logging separately
	}

	auth := &AuthCore{
		config:    cfg,
		database:  db,
		plugins:   plugins,
		router:    router.NewRouterWithOptions(routerOptions),
		transport: transport.NewDefault(),
		session:   NewSessionService(db, SessionOptions{}),
	}

	db.AutoMigrate(models.User{}, models.Session{}, models.Organization{}, models.Account{})

	auth.registerRoutes()

	for _, p := range plugins {
		if p == nil {
			continue
		}

		// Initialize plugin with the database
		if err := p.Init(context.Background(), db); err != nil {
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

func (c *AuthCore) SetTransport(t transport.Transport) {
	c.transport = t
}

func (c *AuthCore) GetTransport() transport.Transport {
	return c.transport
}

func (c *AuthCore) GetTransportInterface() any {
	return c.transport
}

func (c *AuthCore) GetGormDB() *gorm.DB {
	return c.database
}

func (c *AuthCore) GetConfig() *config.Config {
	return c.config
}

func (c *AuthCore) registerRoutes() {
	// Create route group with the configured prefix
	authGroup := c.router

	// Add CORS middleware if configured
	if c.config.CORSConfig != nil {
		corsOptions := router.CORSOptions{
			AllowedOrigins:   c.config.CORSConfig.AllowedOrigins,
			AllowedMethods:   c.config.CORSConfig.AllowedMethods,
			AllowedHeaders:   c.config.CORSConfig.AllowedHeaders,
			AllowCredentials: c.config.CORSConfig.AllowCredentials,
			MaxAge:           24 * time.Hour,
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
	// TODO: Rate limiting logic would be implemented here
	return r
}

func (c *AuthCore) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := c.database.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *AuthCore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := c.database.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *AuthCore) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	ipAddress, userAgent string,
	tx *gorm.DB,
) (*models.Session, error) {
	return c.session.CreateSession(ctx, userID, ipAddress, userAgent, tx)
}

func (c *AuthCore) ValidateSession(ctx context.Context, token string) (*models.Session, error) {
	return c.session.ValidateSession(ctx, token)
}

func (c *AuthCore) DeleteSession(ctx context.Context, token string) error {
	return c.session.DeleteSession(ctx, token)
}

func (c *AuthCore) Middleware() *AuthMiddleware {
	return c.middleware
}
