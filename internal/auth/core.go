package auth

import (
	"better-auth/internal/config"
	"better-auth/internal/database"
	"better-auth/internal/models"
	"better-auth/pkg/plugins/core"
	"better-auth/pkg/transport"
	"context"
	"fmt"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// AuthCore is the main authentication engine
type AuthCore struct {
	config         *config.Config
	database       *database.DB
	transport      transport.Transport
	router         *http.ServeMux
	pluginRegistry *core.PluginRegistry
	handlers       *Handlers
	jwt            *JWTService
	oauth          *OAuthService
	session        *SessionService
	middleware     *AuthMiddleware
}

// New creates a new authentication system with plugin support
func New(db *database.DB, cfg *config.Config) (*AuthCore, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	auth := &AuthCore{
		config:    cfg,
		database:  db,
		router:    http.NewServeMux(),
		transport: transport.NewDefault(),
	}

	// Initialize plugin registry
	auth.pluginRegistry = core.NewPluginRegistry(auth, db.GetGormDB())
	auth.pluginRegistry.SetTransportProvider(auth)

	// Set plugin registry on database for migrations
	db.SetPluginRegistry(auth.pluginRegistry)

	auth.initializeServices()
	auth.registerRoutes()

	return auth, nil
}

// NewFromDatabaseConfig creates a new authentication system from database config
func NewFromDatabaseConfig(
	dbConfig *database.DatabaseConfig,
	cfg *config.Config,
) (*AuthCore, error) {
	// If no dialector is provided, use SQLite as default
	if dbConfig.Dialector == nil {
		filePath := dbConfig.FilePath
		if filePath == "" {
			filePath = "./better-auth.db"
		}
		dbConfig.Dialector = sqlite.Open(filePath)
	}

	db, err := database.NewGormDBFromConfig(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	return New(db, cfg)
}

// ServeHTTP implements http.Handler
func (c *AuthCore) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply global middleware
	r = c.applyCORS(w, r)
	if r == nil {
		return
	}

	r = c.applyRateLimit(w, r)
	if r == nil {
		return
	}

	// Apply plugin middleware
	handler := c.withPluginMiddleware(c.router)
	handler.ServeHTTP(w, r)
}

// RegisterPlugin registers a plugin but doesn't enable it
func (c *AuthCore) RegisterPlugin(plugin core.Plugin) error {
	return c.pluginRegistry.Register(plugin)
}

// EnablePlugin enables a plugin with the given configuration
func (c *AuthCore) EnablePlugin(
	ctx context.Context,
	pluginName string,
	config map[string]any,
) error {
	err := c.pluginRegistry.Enable(ctx, pluginName, config)
	if err != nil {
		return err
	}

	// Migrate plugin tables
	if err := c.database.MigratePluginModels(); err != nil {
		return fmt.Errorf("failed to migrate plugin models: %w", err)
	}

	// Apply plugin routes
	c.pluginRegistry.ApplyRoutes()

	return nil
}

// DisablePlugin disables a plugin
func (c *AuthCore) DisablePlugin(pluginName string) error {
	return c.pluginRegistry.Disable(pluginName)
}

// GetPlugin returns a plugin by name
func (c *AuthCore) GetPlugin(name string) (core.Plugin, bool) {
	return c.pluginRegistry.GetEnabled(name)
}

// GetPluginService returns a plugin service by name
func (c *AuthCore) GetPluginService(name string) (core.PluginService, bool) {
	return c.pluginRegistry.GetServiceRegistry().Get(name)
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

// GetDatabase returns the database instance
func (c *AuthCore) GetDatabase() *gorm.DB {
	return c.database.GetGormDB()
}

// GetGormDB returns the GORM database instance
func (c *AuthCore) GetGormDB() *database.DB {
	return c.database
}

// GetConfig returns the configuration
func (c *AuthCore) GetConfig() *config.Config {
	return c.config
}

func (c *AuthCore) initializeServices() {
	c.handlers = NewHandlers(c)
	c.jwt = NewJWTService([]byte(c.config.SecretKey), "better-auth", c.config.JWTExpiry)
	c.oauth = NewOAuthService(c.config)
	c.session = NewSessionService(c.database.GetGormDB(), SessionOptions{
		Expiry: c.config.SessionExpiry,
		Secure: false, // TODO: Add to config
		Domain: "",    // TODO: Add to config
	})
	c.middleware = NewAuthMiddleware(&MiddlewareConfig{
		SessionService: c.session,
		JWTService:     c.jwt,
		Transport:      c.transport,
		SkipPaths: []string{
			c.config.PathPrefix + "/sign-in",
			c.config.PathPrefix + "/sign-up",
		},
	})
}

func (c *AuthCore) registerRoutes() {
	prefix := c.config.PathPrefix

	// Authentication routes
	c.router.HandleFunc(fmt.Sprintf("POST %s/sign-up", prefix), c.handlers.SignUp)
	c.router.HandleFunc(fmt.Sprintf("POST %s/sign-in", prefix), c.handlers.SignIn)
	c.router.HandleFunc(fmt.Sprintf("POST %s/sign-out", prefix), c.handlers.SignOut)
	c.router.HandleFunc(fmt.Sprintf("GET %s/session", prefix), c.handlers.GetSession)
	c.router.HandleFunc(fmt.Sprintf("POST %s/reset-password", prefix), c.handlers.ResetPassword)
	c.router.HandleFunc(fmt.Sprintf("POST %s/verify-email", prefix), c.handlers.VerifyEmail)

	// Two-factor authentication routes
	setupTwoFactorHandler := c.middleware.RequireAuth(http.HandlerFunc(c.handlers.SetupTwoFactor))
	c.router.Handle(fmt.Sprintf("POST %s/two-factor/setup", prefix), setupTwoFactorHandler)

	verifyTwoFactorHandler := c.middleware.RequireAuth(http.HandlerFunc(c.handlers.VerifyTwoFactor))
	c.router.Handle(fmt.Sprintf("POST %s/two-factor/verify", prefix), verifyTwoFactorHandler)

	// OAuth routes
	c.router.HandleFunc(fmt.Sprintf("GET %s/oauth/{provider}", prefix), c.handlers.OAuthRedirect)
	c.router.HandleFunc(
		fmt.Sprintf("GET %s/oauth/{provider}/callback", prefix),
		c.handlers.OAuthCallback,
	)

	// Organization routes
	createOrgHandler := c.middleware.RequireAuth(http.HandlerFunc(c.handlers.CreateOrganization))
	c.router.Handle(fmt.Sprintf("POST %s/organization/create", prefix), createOrgHandler)

	getOrgHandler := c.middleware.RequireAuth(http.HandlerFunc(c.handlers.GetOrganization))
	c.router.Handle(fmt.Sprintf("GET %s/organization/{id}", prefix), getOrgHandler)

	inviteHandler := c.middleware.RequireAuth(http.HandlerFunc(c.handlers.InviteToOrganization))
	c.router.Handle(fmt.Sprintf("POST %s/organization/{id}/invite", prefix), inviteHandler)

	updateRoleHandler := c.middleware.RequireAuth(http.HandlerFunc(c.handlers.UpdateMemberRole))
	c.router.Handle(
		fmt.Sprintf("POST %s/organization/{id}/members/{userId}/role", prefix),
		updateRoleHandler,
	)
}

func (c *AuthCore) applyCORS(w http.ResponseWriter, r *http.Request) *http.Request {
	if c.config.CORSConfig == nil {
		return r
	}

	cors := c.config.CORSConfig
	origin := r.Header.Get("Origin")

	if len(cors.AllowedOrigins) > 0 {
		allowed := false
		for _, allowedOrigin := range cors.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}

	if len(cors.AllowedMethods) > 0 {
		methods := ""
		for i, method := range cors.AllowedMethods {
			if i > 0 {
				methods += ", "
			}
			methods += method
		}
		w.Header().Set("Access-Control-Allow-Methods", methods)
	}

	if len(cors.AllowedHeaders) > 0 {
		headers := ""
		for i, header := range cors.AllowedHeaders {
			if i > 0 {
				headers += ", "
			}
			headers += header
		}
		w.Header().Set("Access-Control-Allow-Headers", headers)
	}

	if cors.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	return r
}

func (c *AuthCore) applyRateLimit(_ http.ResponseWriter, r *http.Request) *http.Request {
	if !c.config.RateLimitEnabled {
		return r
	}
	// Rate limiting logic would be implemented here
	return r
}

func (c *AuthCore) withPluginMiddleware(handler http.Handler) http.Handler {
	for _, plugin := range c.pluginRegistry.GetAllEnabled() {
		if middleware, ok := plugin.(interface {
			Middleware(http.Handler) http.Handler
		}); ok {
			handler = middleware.Middleware(handler)
		}
	}
	return handler
}

// AddRoute adds a custom route to the router
func (c *AuthCore) AddRoute(path string, handler http.Handler) {
	c.router.Handle(path, handler)
}

// GetUser retrieves a user by ID
func (c *AuthCore) GetUser(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	if err := c.database.GetGormDB().WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (c *AuthCore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := c.database.GetGormDB().WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
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
	userID, ipAddress, userAgent string,
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
