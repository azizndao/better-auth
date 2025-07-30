package auth

import (
	"context"
	"fmt"
	"net/http"

	"better-auth/internal/config"
	"better-auth/internal/database"
	"better-auth/internal/models"
	"better-auth/internal/transport"
	"better-auth/pkg/plugins"
)

// BetterAuth is the main authentication engine
type BetterAuth struct {
	config     *config.Config
	db         database.Database
	transport  transport.Transport
	router     *http.ServeMux
	plugins    []plugins.Plugin
	handlers   *Handlers
	jwt        *JWTService
	oauth      *OAuthService
	session    *SessionService
	middleware *AuthMiddleware
}

// New creates a new authentication system
func New(cfg *config.Config) (*BetterAuth, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	core := &BetterAuth{
		config:    cfg,
		router:    http.NewServeMux(),
		transport: transport.NewDefault(),
		plugins:   make([]plugins.Plugin, 0),
	}

	if err := core.initializeDatabase(); err != nil {
		return nil, err
	}

	core.initializeServices()
	core.registerRoutes()

	return core, nil
}

// ServeHTTP implements http.Handler
func (c *BetterAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

// Use adds a plugin
func (c *BetterAuth) Use(plugin plugins.Plugin) error {
	if err := plugin.Initialize(); err != nil {
		return err
	}
	c.plugins = append(c.plugins, plugin)
	return nil
}

// SetTransport sets a custom transport
func (c *BetterAuth) SetTransport(t transport.Transport) {
	c.transport = t
}

// GetTransport returns the current transport
func (c *BetterAuth) GetTransport() transport.Transport {
	return c.transport
}

// GetDatabase returns the database instance
func (c *BetterAuth) GetDatabase() database.Database {
	return c.db
}

// GetConfig returns the configuration
func (c *BetterAuth) GetConfig() *config.Config {
	return c.config
}

func (c *BetterAuth) initializeDatabase() error {
	switch {
	case len(c.config.DatabaseURL) >= 8 && c.config.DatabaseURL[:8] == "postgres":
		c.db = database.NewPostgresDB(c.config.DatabaseURL)
	case len(c.config.DatabaseURL) >= 5 && c.config.DatabaseURL[:5] == "mysql":
		c.db = database.NewMySQLDB(c.config.DatabaseURL)
	case len(c.config.DatabaseURL) >= 6 && c.config.DatabaseURL[:6] == "sqlite":
		c.db = database.NewSQLiteDB(c.config.DatabaseURL)
	default:
		c.db = database.NewInMemoryDB()
	}
	return nil
}

func (c *BetterAuth) initializeServices() {
	c.handlers = NewHandlers(c)
	c.jwt = NewJWTService([]byte(c.config.SecretKey), "better-auth", c.config.JWTExpiry)
	c.oauth = NewOAuthService(c.config)
	c.session = NewSessionService(c.db, SessionOptions{
		Expiry: c.config.SessionExpiry,
		Secure: false, // TODO: Add to config
		Domain: "",    // TODO: Add to config
	})
	c.middleware = NewAuthMiddleware(&MiddlewareConfig{
		SessionService: c.session,
		JWTService:     c.jwt,
		Transport:      c.transport,
		SkipPaths:      []string{c.config.PathPrefix + "/sign-in", c.config.PathPrefix + "/sign-up"},
	})
}

func (c *BetterAuth) registerRoutes() {
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
	c.router.HandleFunc(fmt.Sprintf("GET %s/oauth/{provider}/callback", prefix), c.handlers.OAuthCallback)

	// Organization routes
	createOrgHandler := c.middleware.RequireAuth(http.HandlerFunc(c.handlers.CreateOrganization))
	c.router.Handle(fmt.Sprintf("POST %s/organization/create", prefix), createOrgHandler)

	getOrgHandler := c.middleware.RequireAuth(http.HandlerFunc(c.handlers.GetOrganization))
	c.router.Handle(fmt.Sprintf("GET %s/organization/{id}", prefix), getOrgHandler)

	inviteHandler := c.middleware.RequireAuth(http.HandlerFunc(c.handlers.InviteToOrganization))
	c.router.Handle(fmt.Sprintf("POST %s/organization/{id}/invite", prefix), inviteHandler)

	updateRoleHandler := c.middleware.RequireAuth(http.HandlerFunc(c.handlers.UpdateMemberRole))
	c.router.Handle(fmt.Sprintf("POST %s/organization/{id}/members/{userId}/role", prefix), updateRoleHandler)
}

func (c *BetterAuth) applyCORS(w http.ResponseWriter, r *http.Request) *http.Request {
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

func (c *BetterAuth) applyRateLimit(w http.ResponseWriter, r *http.Request) *http.Request {
	if !c.config.RateLimitEnabled {
		return r
	}
	// Rate limiting logic would be implemented here
	return r
}

func (c *BetterAuth) withPluginMiddleware(handler http.Handler) http.Handler {
	for _, plugin := range c.plugins {
		if middleware, ok := plugin.(interface{ Middleware(http.Handler) http.Handler }); ok {
			handler = middleware.Middleware(handler)
		}
	}
	return handler
}

// AddRoute adds a custom route to the router
func (c *BetterAuth) AddRoute(path string, handler http.Handler) {
	c.router.Handle(path, handler)
}

// GetUser retrieves a user by ID
func (c *BetterAuth) GetUser(ctx context.Context, userID string) (*models.User, error) {
	return c.db.GetUser(ctx, userID)
}

// GetUserByEmail retrieves a user by email
func (c *BetterAuth) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return c.db.GetUserByEmail(ctx, email)
}

// ValidateJWT validates a JWT token and returns claims
func (c *BetterAuth) ValidateJWT(tokenString string) (*JWTClaims, error) {
	return c.jwt.ValidateToken(tokenString)
}

// GenerateJWT generates a JWT token for a user
func (c *BetterAuth) GenerateJWT(user *models.User) (string, error) {
	return c.jwt.GenerateToken(user)
}

// CreateSession creates a new session for a user
func (c *BetterAuth) CreateSession(ctx context.Context, userID, ipAddress, userAgent string) (*models.Session, error) {
	return c.session.CreateSession(ctx, userID, ipAddress, userAgent)
}

// ValidateSession validates a session token
func (c *BetterAuth) ValidateSession(ctx context.Context, token string) (*models.Session, error) {
	return c.session.ValidateSession(ctx, token)
}

// DeleteSession deletes a session
func (c *BetterAuth) DeleteSession(ctx context.Context, token string) error {
	return c.session.DeleteSession(ctx, token)
}

// Middleware returns the middleware manager
func (c *BetterAuth) Middleware() *AuthMiddleware {
	return c.middleware
}