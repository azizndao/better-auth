package config

import "time"

// Config holds the configuration for Better Auth
type Config struct {
	DatabaseURL      string        `json:"databaseUrl"`
	SecretKey        string        `json:"secretKey"`
	SessionExpiry    time.Duration `json:"sessionExpiry"`
	JWTExpiry        time.Duration `json:"jwtExpiry"`
	RateLimitEnabled bool          `json:"rateLimitEnabled"`
	TwoFactorEnabled bool          `json:"twoFactorEnabled"`
	Providers        []OAuthProvider `json:"providers"`
	BaseURL          string        `json:"baseUrl"`
	EmailConfig      *EmailConfig  `json:"emailConfig"`
	CORSConfig       *CORSConfig   `json:"corsConfig"`
	PathPrefix       string        `json:"pathPrefix"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `json:"allowedOrigins"`
	AllowedMethods   []string `json:"allowedMethods"`
	AllowedHeaders   []string `json:"allowedHeaders"`
	AllowCredentials bool     `json:"allowCredentials"`
}

// OAuthProvider configuration
type OAuthProvider struct {
	Name         string   `json:"name"`
	ClientID     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	RedirectURL  string   `json:"redirectUrl"`
	Scopes       []string `json:"scopes"`
}

// EmailConfig holds email configuration
type EmailConfig struct {
	Provider string            `json:"provider"`
	From     string            `json:"from"`
	SMTP     *SMTPConfig       `json:"smtp,omitempty"`
	Config   map[string]string `json:"config"`
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	TLS      bool   `json:"tls"`
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.SecretKey == "" {
		return ErrMissingSecretKey
	}
	if c.PathPrefix == "" {
		c.PathPrefix = "/auth"
	}
	if c.SessionExpiry == 0 {
		c.SessionExpiry = 24 * time.Hour
	}
	if c.JWTExpiry == 0 {
		c.JWTExpiry = 1 * time.Hour
	}
	return nil
}