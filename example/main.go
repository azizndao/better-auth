package main

import (
	"log"
	"net/http"
	"time"

	betterauth "better-auth"
	"better-auth/internal/config"
)

func main() {
	// Create configuration
	cfg := &betterauth.Config{
		DatabaseURL:      "sqlite:./auth.db",
		SecretKey:        "your-secret-key-here-make-it-long-and-secure",
		SessionExpiry:    24 * time.Hour,
		JWTExpiry:        1 * time.Hour,
		RateLimitEnabled: true,
		TwoFactorEnabled: true,
		BaseURL:          "http://localhost:8080",
		PathPrefix:       "/auth",
		CORSConfig:       betterauth.DefaultCORSConfig(),
		Providers: []config.OAuthProvider{
			{
				Name:         "google",
				ClientID:     "your-google-client-id",
				ClientSecret: "your-google-client-secret",
				RedirectURL:  "http://localhost:8080/auth/oauth/google/callback",
				Scopes:       []string{"openid", "email", "profile"},
			},
			{
				Name:         "github",
				ClientID:     "your-github-client-id",
				ClientSecret: "your-github-client-secret",
				RedirectURL:  "http://localhost:8080/auth/oauth/github/callback",
				Scopes:       []string{"user:email"},
			},
		},
		EmailConfig: &config.EmailConfig{
			Provider: "smtp",
			From:     "noreply@yourapp.com",
			SMTP: &config.SMTPConfig{
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "your-email@gmail.com",
				Password: "your-app-password",
			},
		},
	}

	// Initialize Better Auth
	auth, err := betterauth.New(cfg)
	if err != nil {
		log.Fatal("Failed to initialize Better Auth:", err)
	}

	// Create HTTP server with Better Auth handler
	mux := http.NewServeMux()
	
	// Mount Better Auth routes
	mux.Handle("/auth/", auth)

	// Example protected route
	protectedHandler := auth.Middleware().SessionAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "This is a protected route", "user": "authenticated"}`))
	}))
	mux.Handle("/api/profile", protectedHandler)

	// Example public route
	mux.HandleFunc("/api/public", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "This is a public route"}`))
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	})

	// Start server
	log.Println("Starting server on :8080")
	log.Println("Better Auth mounted at /auth")
	log.Println("Available endpoints:")
	log.Println("  POST /auth/sign-up")
	log.Println("  POST /auth/sign-in")
	log.Println("  POST /auth/sign-out")
	log.Println("  GET  /auth/session")
	log.Println("  POST /auth/reset-password")
	log.Println("  POST /auth/verify-email")
	log.Println("  GET  /auth/oauth/google")
	log.Println("  GET  /auth/oauth/github")
	log.Println("  GET  /api/profile (protected)")
	log.Println("  GET  /api/public")
	log.Println("  GET  /health")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}