package main

import (
	"context"
	"log"
	"time"

	betterauth "better-auth"
	"better-auth/internal/config"
	"better-auth/pkg/plugins/jwt"
	"better-auth/pkg/plugins/organizations"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	ctx := context.Background()

	// Example 1: Using PostgreSQL with custom GORM config
	_ = func() { // PostgreSQL example (commented out, requires database)
		// Create GORM dialector for PostgreSQL
		dialector := postgres.Open("host=localhost user=username password=password dbname=better_auth port=5432 sslmode=disable")
		
		// Create custom GORM config
		gormConfig := &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
			// Add other GORM configurations as needed
		}

		// Create database configuration
		dbConfig := betterauth.NewDatabaseConfig(dialector, gormConfig)

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

		// Create Better Auth instance with custom database
		auth, err := betterauth.NewWithDatabase(cfg, dbConfig)
		if err != nil {
			log.Fatal("Failed to create auth system with PostgreSQL:", err)
		}

		log.Println("PostgreSQL example initialized successfully!")
		_ = auth // Use the auth instance
	}

	// Example 2: Using SQLite with custom configuration
	sqliteExample := func() {
		// Create GORM dialector for SQLite
		dialector := sqlite.Open("./custom-better-auth.db")
		
		// Create custom GORM config with different settings
		gormConfig := &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent), // Silent logging
		}

		// Create database configuration
		dbConfig := betterauth.NewDatabaseConfig(dialector, gormConfig)

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

		// Create Better Auth instance with custom database
		auth, err := betterauth.NewWithDatabase(cfg, dbConfig)
		if err != nil {
			log.Fatal("Failed to create auth system with SQLite:", err)
		}

		// Register and enable plugins
		jwtPlugin := jwt.NewJWTPlugin(&jwt.JWTConfig{
			SecretKey:              []byte(cfg.SecretKey),
			AccessTokenExpiry:      15 * time.Minute,
			RefreshTokenExpiry:     7 * 24 * time.Hour,
			Issuer:                 "better-auth-custom-db",
			EnableRefreshTokens:    true,
			EnableTokenBlacklist:   true,
			CleanupInterval:        1 * time.Hour,
			MaxRefreshTokensPerUser: 5,
		})

		orgPlugin := organizations.NewOrganizationPlugin()

		if err := auth.RegisterPlugin(jwtPlugin); err != nil {
			log.Fatal("Failed to register JWT plugin:", err)
		}

		if err := auth.RegisterPlugin(orgPlugin); err != nil {
			log.Fatal("Failed to register organization plugin:", err)
		}

		// Enable plugins
		if err := auth.EnablePlugin(ctx, "jwt", map[string]interface{}{
			"enabled":              true,
			"secretKey":            cfg.SecretKey,
			"accessTokenExpiry":    15 * time.Minute,
			"refreshTokenExpiry":   7 * 24 * time.Hour,
			"enableRefreshTokens":  true,
			"enableTokenBlacklist": true,
		}); err != nil {
			log.Fatal("Failed to enable JWT plugin:", err)
		}

		if err := auth.EnablePlugin(ctx, "organizations", map[string]interface{}{
			"enabled": true,
		}); err != nil {
			log.Fatal("Failed to enable organizations plugin:", err)
		}

		log.Println("SQLite example with plugins initialized successfully!")

		// Create HTTP server (commented out to avoid conflict)
		/*
		mux := http.NewServeMux()
		mux.Handle("/auth/", auth)
		
		server := &http.Server{
			Addr:    ":8080",
			Handler: mux,
		}

		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil {
			log.Fatal("Server failed:", err)
		}
		*/

		_ = auth // Use the auth instance
	}

	// Example 3: Using default SQLite (backwards compatibility)
	defaultExample := func() {
		// Create auth configuration (uses default SQLite)
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

		// Create Better Auth instance (uses default SQLite database)
		auth, err := betterauth.New(cfg)
		if err != nil {
			log.Fatal("Failed to create auth system with default config:", err)
		}

		log.Println("Default SQLite example initialized successfully!")
		_ = auth // Use the auth instance
	}

	// Run examples
	log.Println("=== Custom Database Configuration Examples ===")
	log.Println()

	log.Println("1. PostgreSQL Example:")
	// postgresExample() // Uncomment to test PostgreSQL (requires database)

	log.Println("2. SQLite with Custom Config Example:")
	sqliteExample()

	log.Println("3. Default Configuration Example:")
	defaultExample()

	log.Println()
	log.Println("All examples completed successfully!")
	log.Println("You can now use any GORM dialector (PostgreSQL, MySQL, SQLite, etc.) with Better Auth!")
}