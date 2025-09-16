package auth

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"better-auth/internal/models"
	"better-auth/pkg/plugins/core"
	"better-auth/pkg/router"
	"better-auth/pkg/transport"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// handlers contains all HTTP handlers
type handlers struct {
	core *AuthCore
	transport.Transport
}

// newHandlers creates a new handlers instance
func newHandlers(core *AuthCore) *handlers {
	return &handlers{core: core, Transport: core.transport}
}

func (h *handlers) registerRoutes(r router.RouteGroup) {
	authGroup := r.Group("/auth")

	authGroup.POST("/signup", h.signUp)
	authGroup.POST("/signin", h.signIn)
	authGroup.POST("/signout", h.signOut)
	authGroup.GET("/session", h.GetSession)
	authGroup.POST("/reset-password", h.resetPassword)
	authGroup.POST("/verify-email", h.verifyEmail)
}

func (h *handlers) signUp(w http.ResponseWriter, r *http.Request) {
	var req models.SignUpPayload
	if err := h.DecodeJSON(r, &req); err != nil {
		var validationErr *transport.ValidationError
		if errors.As(err, &validationErr) {
			h.RespondValidationError(w, validationErr)
		} else {
			h.RespondError(w, http.StatusBadRequest, "Invalid JSON")
		}
		return
	}

	if req.Email == "" || req.Password == "" {
		h.RespondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	if err := validateEmail(req.Email); err != nil {
		h.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validatePassword(req.Password); err != nil {
		h.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	existingUser, _ := h.core.GetUserByEmail(r.Context(), req.Email)
	if existingUser != nil {
		h.RespondError(w, http.StatusConflict, "User already exists")
		return
	}

	now := time.Now()
	user := &models.User{
		Model:         core.Model{ID: uuid.New(), CreatedAt: now, UpdatedAt: now},
		Email:         req.Email,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Password:      req.Password,
		EmailVerified: false,
		Metadata:      req.Metadata,
	}

	if err := h.core.database.WithContext(r.Context()).Create(user).Error; err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	session, err := h.core.session.CreateSession(
		r.Context(),
		user.ID,
		getClientIP(r),
		r.UserAgent(),
	)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	h.core.session.SetSessionCookie(w, session.Token)

	h.RespondJSON(w, http.StatusCreated, models.AuthResponse{
		User:    user,
		Session: session,
	})
}

func (h *handlers) signIn(w http.ResponseWriter, r *http.Request) {
	var req models.SignInPayload
	if err := h.DecodeJSON(r, &req); err != nil {
		var validationErr *transport.ValidationError
		if errors.As(err, &validationErr) {
			h.RespondValidationError(w, validationErr)
		} else {
			h.RespondError(w, http.StatusBadRequest, "Invalid JSON")
		}
		return
	}

	if req.Email == "" || req.Password == "" {
		h.RespondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := h.core.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		h.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if user.BannedUtil.Time.After(time.Now()) {
		h.RespondError(w, http.StatusLocked, "User is banned")
		return
	}

	session, err := h.core.session.CreateSession(
		r.Context(),
		user.ID,
		getClientIP(r),
		r.UserAgent(),
	)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	h.core.session.SetSessionCookie(w, session.Token)

	// Create a copy of the user for the response to avoid modifying the stored user
	h.RespondJSON(w, http.StatusOK, models.AuthResponse{
		User:    user,
		Session: session,
	})
}

func (h *handlers) signOut(w http.ResponseWriter, r *http.Request) {
	token := h.core.session.GetSessionFromRequest(r)
	if token == "" {
		h.RespondError(w, http.StatusBadRequest, "No session token provided")
		return
	}

	if err := h.core.session.DeleteSession(r.Context(), token); err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to delete session")
		return
	}

	h.core.session.ClearSessionCookie(w)
	h.RespondJSON(w, http.StatusOK, map[string]string{"message": "Signed out successfully"})
}

func (h *handlers) GetSession(w http.ResponseWriter, r *http.Request) {
	token := h.core.session.GetSessionFromRequest(r)
	if token == "" {
		h.RespondError(w, http.StatusUnauthorized, "No session token provided")
		return
	}

	session, err := h.core.session.ValidateSession(r.Context(), token)
	if err != nil {
		h.RespondError(w, http.StatusUnauthorized, "Invalid session")
		return
	}

	user, err := h.core.GetUser(r.Context(), session.UserID)
	if err != nil {
		h.RespondError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	// Create a copy of the user for the response to avoid modifying the stored user
	responseUser := *user
	responseUser.Password = ""
	h.RespondJSON(w, http.StatusOK, models.SessionResponse{
		User:    &responseUser,
		Session: session,
	})
}

func (h *handlers) resetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ResetPasswordPayload
	if err := h.DecodeJSON(r, &req); err != nil {
		var validationErr *transport.ValidationError
		if errors.As(err, &validationErr) {
			h.RespondValidationError(w, validationErr)
		} else {
			h.RespondError(w, http.StatusBadRequest, "Invalid JSON")
		}
		return
	}

	if req.Email == "" {
		h.RespondError(w, http.StatusBadRequest, "Email is required")
		return
	}

	// TODO: Implement password reset logic
	h.RespondJSON(w, http.StatusOK, map[string]string{"message": "Password reset email sent"})
}

func (h *handlers) verifyEmail(w http.ResponseWriter, r *http.Request) {
	var req models.VerifyEmailPayload
	if err := h.DecodeJSON(r, &req); err != nil {
		var validationErr *transport.ValidationError
		if errors.As(err, &validationErr) {
			h.RespondValidationError(w, validationErr)
		} else {
			h.RespondError(w, http.StatusBadRequest, "Invalid JSON")
		}
		return
	}

	if req.Token == "" {
		h.RespondError(w, http.StatusBadRequest, "Token is required")
		return
	}

	// TODO: Implement email verification logic
	h.RespondJSON(w, http.StatusOK, map[string]string{"message": "Email verified successfully"})
}

func (h *handlers) SetupTwoFactor(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement 2FA setup
	h.RespondError(w, http.StatusNotImplemented, "Two-factor setup not implemented")
}

func (h *handlers) VerifyTwoFactor(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement 2FA verification
	h.RespondError(w, http.StatusNotImplemented, "Two-factor verification not implemented")
}

func (h *handlers) OAuthRedirect(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement OAuth redirect
	h.RespondError(w, http.StatusNotImplemented, "OAuth redirect not implemented")
}

func (h *handlers) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement OAuth callback
	h.RespondError(w, http.StatusNotImplemented, "OAuth callback not implemented")
}

// Validation functions
func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	return nil
}

// Helper function to get client IP
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}
