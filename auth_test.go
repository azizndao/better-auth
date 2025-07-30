package betterauth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"better-auth/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBetterAuth_SignUp(t *testing.T) {
	config := &Config{
		DatabaseURL:   "memory",
		SecretKey:     "test-secret-key-for-testing-purposes-only",
		SessionExpiry: 24 * time.Hour,
		JWTExpiry:     1 * time.Hour,
		BaseURL:       "http://localhost:8080",
		PathPrefix:    "/auth",
	}

	auth, err := New(config)
	require.NoError(t, err)

	t.Run("Successful sign up", func(t *testing.T) {
		reqBody := models.SignUpRequest{
			Email:    "test@example.com",
			Password: "TestPassword123!",
			Name:     "Test User",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, reqBody.Email, response.User.Email)
		assert.Equal(t, reqBody.Name, response.User.Name)
		assert.NotEmpty(t, response.Token)
		assert.NotNil(t, response.Session)
	})

	t.Run("Duplicate email", func(t *testing.T) {
		reqBody := models.SignUpRequest{
			Email:    "test@example.com",
			Password: "TestPassword123!",
			Name:     "Test User 2",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Invalid email", func(t *testing.T) {
		reqBody := models.SignUpRequest{
			Email:    "invalid-email",
			Password: "TestPassword123!",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestBetterAuth_SignIn(t *testing.T) {
	// Create a user that will persist across subtests
	createUser := func() *BetterAuth {
		config := &Config{
			DatabaseURL:   "memory",
			SecretKey:     "test-secret-key-for-testing-purposes-only",
			SessionExpiry: 24 * time.Hour,
			JWTExpiry:     1 * time.Hour,
			BaseURL:       "http://localhost:8080",
			PathPrefix:    "/auth",
		}

		auth, err := New(config)
		require.NoError(t, err)

		// Create a user first
		signUpBody := models.SignUpRequest{
			Email:    "signin@example.com",
			Password: "TestPassword123!",
			Name:     "Sign In User",
		}
		body, _ := json.Marshal(signUpBody)
		req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		auth.ServeHTTP(w, req)
		
		// Verify user was created successfully
		require.Equal(t, http.StatusCreated, w.Code, "User creation should succeed before testing sign-in")
		
		return auth
	}

	auth := createUser()

	t.Run("Successful sign in", func(t *testing.T) {
		reqBody := models.SignInRequest{
			Email:    "signin@example.com",
			Password: "TestPassword123!",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/sign-in", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, reqBody.Email, response.User.Email)
		assert.NotEmpty(t, response.Token)
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		reqBody := models.SignInRequest{
			Email:    "signin@example.com",
			Password: "WrongPassword123!",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/sign-in", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Non-existent user", func(t *testing.T) {
		reqBody := models.SignInRequest{
			Email:    "nonexistent@example.com",
			Password: "TestPassword123!",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/auth/sign-in", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestBetterAuth_GetSession(t *testing.T) {
	config := &Config{
		DatabaseURL:   "memory",
		SecretKey:     "test-secret-key-for-testing-purposes-only",
		SessionExpiry: 24 * time.Hour,
		JWTExpiry:     1 * time.Hour,
		BaseURL:       "http://localhost:8080",
		PathPrefix:    "/auth",
	}

	auth, err := New(config)
	require.NoError(t, err)

	// Create a user and get session token
	signUpBody := models.SignUpRequest{
		Email:    "session@example.com",
		Password: "TestPassword123!",
		Name:     "Session User",
	}
	body, _ := json.Marshal(signUpBody)
	req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	auth.ServeHTTP(w, req)

	var signUpResponse models.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &signUpResponse)
	require.NoError(t, err)
	require.NotNil(t, signUpResponse.Session, "Session should not be nil")

	t.Run("Valid session", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/session", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: signUpResponse.Session.Token,
		})
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.SessionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response.User)
		assert.NotNil(t, response.Session)
	})

	t.Run("No token provided", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/session", nil)
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/session", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "invalid-token",
		})
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestBetterAuth_Middleware(t *testing.T) {
	config := &Config{
		DatabaseURL:   "memory",
		SecretKey:     "test-secret-key-for-testing-purposes-only",
		SessionExpiry: 24 * time.Hour,
		JWTExpiry:     1 * time.Hour,
		BaseURL:       "http://localhost:8080",
		PathPrefix:    "/auth",
	}

	auth, err := New(config)
	require.NoError(t, err)

	// Create a user and get session token
	signUpBody := models.SignUpRequest{
		Email:    "middleware@example.com",
		Password: "TestPassword123!",
		Name:     "Middleware User",
	}
	body, _ := json.Marshal(signUpBody)
	req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	auth.ServeHTTP(w, req)

	var signUpResponse models.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &signUpResponse)
	require.NoError(t, err)
	require.NotNil(t, signUpResponse.Session, "Session should not be nil")

	// Create a protected handler using middleware
	protectedHandler := auth.Middleware().SessionAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Protected endpoint accessed",
		})
	}))

	t.Run("Valid session token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: signUpResponse.Session.Token,
		})
		w := httptest.NewRecorder()

		protectedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Protected endpoint accessed", response["message"])
	})

	t.Run("No token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		protectedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestBetterAuth_JWT(t *testing.T) {
	config := &Config{
		DatabaseURL:   "memory",
		SecretKey:     "test-secret-key-for-testing-purposes-only",
		SessionExpiry: 24 * time.Hour,
		JWTExpiry:     1 * time.Hour,
		BaseURL:       "http://localhost:8080",
	}

	auth, err := New(config)
	require.NoError(t, err)

	user := &models.User{
		ID:            "test-user-id",
		Email:         "jwt@example.com",
		Name:          "JWT User",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = auth.GetDatabase().CreateUser(context.Background(), user)
	require.NoError(t, err)

	t.Run("Generate and validate JWT", func(t *testing.T) {
		token, err := auth.GenerateJWT(user)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := auth.ValidateJWT(token)
		require.NoError(t, err)

		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Email, claims.Email)
		assert.Equal(t, user.Name, claims.Name)
	})

	t.Run("Invalid JWT signature", func(t *testing.T) {
		token, err := auth.GenerateJWT(user)
		require.NoError(t, err)

		// Tamper with the token
		tamperedToken := token[:len(token)-10] + "tampered123"

		_, err = auth.ValidateJWT(tamperedToken)
		assert.Error(t, err)
	})
}

func TestBetterAuth_Database(t *testing.T) {
	config := &Config{
		DatabaseURL:   "memory",
		SecretKey:     "test-secret-key-for-testing-purposes-only",
		SessionExpiry: 24 * time.Hour,
		JWTExpiry:     1 * time.Hour,
		BaseURL:       "http://localhost:8080",
	}

	auth, err := New(config)
	require.NoError(t, err)

	db := auth.GetDatabase()

	user := &models.User{
		ID:        "test-id",
		Email:     "db@example.com",
		Name:      "DB User",
		Password:  "hashed-password",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("Create and get user", func(t *testing.T) {
		err := db.CreateUser(context.Background(), user)
		require.NoError(t, err)

		retrievedUser, err := db.GetUser(context.Background(), user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Email, retrievedUser.Email)
		assert.Equal(t, user.Name, retrievedUser.Name)
	})

	t.Run("Get user by email", func(t *testing.T) {
		retrievedUser, err := db.GetUserByEmail(context.Background(), user.Email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrievedUser.ID)
	})

	t.Run("Update user", func(t *testing.T) {
		user.Name = "Updated Name"
		err := db.UpdateUser(context.Background(), user)
		require.NoError(t, err)

		retrievedUser, err := db.GetUser(context.Background(), user.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", retrievedUser.Name)
	})

	t.Run("Session operations", func(t *testing.T) {
		session := &models.Session{
			ID:        "session-id",
			UserID:    "user-id",
			Token:     "session-token",
			ExpiresAt: time.Now().Add(time.Hour),
			CreatedAt: time.Now(),
			Active:    true,
		}

		err := db.CreateSession(context.Background(), session)
		require.NoError(t, err)

		retrievedSession, err := db.GetSession(context.Background(), session.Token)
		require.NoError(t, err)
		assert.Equal(t, session.UserID, retrievedSession.UserID)

		err = db.DeleteSession(context.Background(), session.Token)
		require.NoError(t, err)

		_, err = db.GetSession(context.Background(), session.Token)
		assert.Error(t, err)
	})

	t.Run("Delete user", func(t *testing.T) {
		err := db.DeleteUser(context.Background(), user.ID)
		require.NoError(t, err)

		_, err = db.GetUser(context.Background(), user.ID)
		assert.Error(t, err)
	})
}

func TestBetterAuth_Transport(t *testing.T) {
	config := &Config{
		DatabaseURL:   "memory",
		SecretKey:     "test-secret-key-for-testing-purposes-only",
		SessionExpiry: 24 * time.Hour,
		JWTExpiry:     1 * time.Hour,
		BaseURL:       "http://localhost:8080",
	}

	auth, err := New(config)
	require.NoError(t, err)

	transport := auth.GetTransport()

	t.Run("ExtractToken from Authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		token := transport.ExtractToken(req)
		assert.Equal(t, "test-token", token)
	})

	t.Run("ExtractToken from cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "better-auth.session_token",
			Value: "cookie-token",
		})

		token := transport.ExtractToken(req)
		assert.Equal(t, "cookie-token", token)
	})

	t.Run("ExtractToken from query parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?token=query-token", nil)

		token := transport.ExtractToken(req)
		assert.Equal(t, "query-token", token)
	})

	t.Run("UserContext operations", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)

		_, ok := transport.GetUserContext(req)
		assert.False(t, ok)

		userCtx := &UserContext{
			User: &models.User{ID: "test-user"},
		}

		req = transport.SetUserContext(req, userCtx)

		retrievedCtx, ok := transport.GetUserContext(req)
		assert.True(t, ok)
		assert.Equal(t, "test-user", retrievedCtx.User.ID)
	})
}

func TestBetterAuth_Config(t *testing.T) {
	t.Run("Default config", func(t *testing.T) {
		config := DefaultConfig()
		assert.NotNil(t, config)
		assert.Equal(t, "/auth", config.PathPrefix)
		assert.Equal(t, 24*time.Hour, config.SessionExpiry)
		assert.Equal(t, 1*time.Hour, config.JWTExpiry)
	})

	t.Run("CORS config", func(t *testing.T) {
		corsConfig := DefaultCORSConfig()
		assert.NotNil(t, corsConfig)
		assert.Contains(t, corsConfig.AllowedOrigins, "*")
		assert.Contains(t, corsConfig.AllowedMethods, "GET")
		assert.Contains(t, corsConfig.AllowedMethods, "POST")
	})
}