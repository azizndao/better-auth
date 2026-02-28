package main

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	betterauth "better-auth"

	"github.com/azizndao/grouter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Database file sqlite file
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	auth := setupAuth(db)

	router := grouter.NewRouter()
	router.Use(grouter.Logger())
	router.Route("/auth", auth.Handler())

	slog.Default().Info("Listening on port 8080")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func setupAuth(db *gorm.DB) *betterauth.BetterAuth {
	cfg := &betterauth.Config{
		SecretKey:     "test-secret-key-for-example-only",
		BaseURL:       "http://localhost:8080",
		PathPrefix:    "/auth",
		SessionExpiry: 24 * time.Hour,
		JWTExpiry:     1 * time.Hour,
		CORSConfig:    betterauth.DefaultCORSConfig(),
	}

	auth, err := betterauth.New(cfg, db, []betterauth.Plugin{})
	if err != nil {
		log.Fatalf("Failed to create auth: %v", err)
	}
	return auth
}
