package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	betterauth "better-auth"
	"better-auth/internal/config"
	"better-auth/pkg/plugins/admin"
	"better-auth/pkg/plugins/jwt"
	"better-auth/pkg/plugins/organizations"
)

func main() {
	ctx := context.Background()

	// Create auth configuration
	cfg := &config.Config{
		SecretKey:        "your-secret-key-at-least-32-characters-long",
		SessionExpiry:    24 * time.Hour,
		JWTExpiry:        1 * time.Hour,
		RateLimitEnabled: false,
		TwoFactorEnabled: true,
		BaseURL:          "http://localhost:8080",
		PathPrefix:       "/auth",
		CORSConfig:       config.DefaultCORSConfig(),
	}

	// Create Better Auth instance
	auth, err := betterauth.New(cfg)
	if err != nil {
		log.Fatal("Failed to create auth system:", err)
	}

	// Create and register plugins
	jwtPlugin := jwt.NewJWTPlugin(&jwt.JWTConfig{
		SecretKey:               []byte(cfg.SecretKey),
		AccessTokenExpiry:       15 * time.Minute,
		RefreshTokenExpiry:      7 * 24 * time.Hour,
		Issuer:                  "better-auth-example",
		EnableRefreshTokens:     true,
		EnableTokenBlacklist:    true,
		CleanupInterval:         1 * time.Hour,
		MaxRefreshTokensPerUser: 5,
	})

	orgPlugin := organizations.NewOrganizationPlugin()
	adminPlugin := admin.NewAdminPlugin()

	// Register plugins
	if err := auth.RegisterPlugin(jwtPlugin); err != nil {
		log.Fatal("Failed to register JWT plugin:", err)
	}

	if err := auth.RegisterPlugin(orgPlugin); err != nil {
		log.Fatal("Failed to register organization plugin:", err)
	}

	if err := auth.RegisterPlugin(adminPlugin); err != nil {
		log.Fatal("Failed to register admin plugin:", err)
	}

	// Enable plugins with configuration
	if err := auth.EnablePlugin(ctx, "jwt", map[string]any{
		"enabled":              true,
		"secretKey":            cfg.SecretKey,
		"accessTokenExpiry":    15 * time.Minute,
		"refreshTokenExpiry":   7 * 24 * time.Hour,
		"enableRefreshTokens":  true,
		"enableTokenBlacklist": true,
	}); err != nil {
		log.Fatal("Failed to enable JWT plugin:", err)
	}

	if err := auth.EnablePlugin(ctx, "organizations", map[string]any{
		"enabled": true,
	}); err != nil {
		log.Fatal("Failed to enable organizations plugin:", err)
	}

	if err := auth.EnablePlugin(ctx, "admin", map[string]any{
		"enabled": true,
	}); err != nil {
		log.Fatal("Failed to enable admin plugin:", err)
	}

	log.Println("Better Auth with plugins initialized successfully!")
	log.Println("Available routes:")
	log.Println("- POST /auth/sign-up")
	log.Println("- POST /auth/sign-in")
	log.Println("- POST /auth/sign-out")
	log.Println("- GET /auth/session")
	log.Println("- POST /auth/jwt/refresh (JWT plugin)")
	log.Println("- POST /auth/organizations (Organizations plugin)")
	log.Println("- GET /auth/admin/system/stats (Admin plugin)")

	// Create HTTP server
	mux := http.NewServeMux()
	mux.Handle("/auth/", auth)

	// Add a health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "healthy", "plugins": ["jwt", "organizations", "admin"]}`)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Server starting on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed:", err)
	}
}

