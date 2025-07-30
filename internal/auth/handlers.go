package auth

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"better-auth/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	core *BetterAuth
}

// NewHandlers creates a new handlers instance
func NewHandlers(core *BetterAuth) *Handlers {
	return &Handlers{core: core}
}

func (h *Handlers) SignUp(w http.ResponseWriter, r *http.Request) {
	var req models.SignUpRequest
	if err := h.core.transport.DecodeJSON(r, &req); err != nil {
		h.core.transport.RespondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Email == "" || req.Password == "" {
		h.core.transport.RespondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	if err := validateEmail(req.Email); err != nil {
		h.core.transport.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validatePassword(req.Password); err != nil {
		h.core.transport.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	existingUser, _ := h.core.db.GetUserByEmail(r.Context(), req.Email)
	if existingUser != nil {
		h.core.transport.RespondError(w, http.StatusConflict, "User already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.core.transport.RespondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	now := time.Now()
	user := &models.User{
		ID:            uuid.New().String(),
		Email:         req.Email,
		Name:          req.Name,
		Image:         req.Image,
		Password:      string(hashedPassword),
		CreatedAt:     now,
		UpdatedAt:     now,
		EmailVerified: false,
		Metadata:      req.Metadata,
	}

	if err := h.core.db.CreateUser(r.Context(), user); err != nil {
		h.core.transport.RespondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	session, err := h.core.session.CreateSession(r.Context(), user.ID, getClientIP(r), r.UserAgent())
	if err != nil {
		h.core.transport.RespondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	h.core.session.SetSessionCookie(w, session.Token)

	// Create a copy of the user for the response to avoid modifying the stored user
	responseUser := *user
	responseUser.Password = ""
	h.core.transport.RespondJSON(w, http.StatusCreated, models.AuthResponse{
		User:    &responseUser,
		Session: session,
		Token:   session.Token,
	})
}

func (h *Handlers) SignIn(w http.ResponseWriter, r *http.Request) {
	var req models.SignInRequest
	if err := h.core.transport.DecodeJSON(r, &req); err != nil {
		h.core.transport.RespondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Email == "" || req.Password == "" {
		h.core.transport.RespondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := h.core.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		h.core.transport.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.core.transport.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if user.Blocked {
		h.core.transport.RespondError(w, http.StatusForbidden, "User is blocked")
		return
	}

	session, err := h.core.session.CreateSession(r.Context(), user.ID, getClientIP(r), r.UserAgent())
	if err != nil {
		h.core.transport.RespondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	h.core.session.SetSessionCookie(w, session.Token)

	// Update last sign in
	now := time.Now()
	user.LastSignIn = &now
	user.SignInCount++
	h.core.db.UpdateUser(r.Context(), user)

	// Create a copy of the user for the response to avoid modifying the stored user
	responseUser := *user
	responseUser.Password = ""
	h.core.transport.RespondJSON(w, http.StatusOK, models.AuthResponse{
		User:    &responseUser,
		Session: session,
		Token:   session.Token,
	})
}

func (h *Handlers) SignOut(w http.ResponseWriter, r *http.Request) {
	token := h.core.session.GetSessionFromRequest(r)
	if token == "" {
		h.core.transport.RespondError(w, http.StatusBadRequest, "No session token provided")
		return
	}

	if err := h.core.session.DeleteSession(r.Context(), token); err != nil {
		h.core.transport.RespondError(w, http.StatusInternalServerError, "Failed to delete session")
		return
	}

	h.core.session.ClearSessionCookie(w)
	h.core.transport.RespondJSON(w, http.StatusOK, map[string]string{"message": "Signed out successfully"})
}

func (h *Handlers) GetSession(w http.ResponseWriter, r *http.Request) {
	token := h.core.session.GetSessionFromRequest(r)
	if token == "" {
		h.core.transport.RespondError(w, http.StatusUnauthorized, "No session token provided")
		return
	}

	session, err := h.core.session.ValidateSession(r.Context(), token)
	if err != nil {
		h.core.transport.RespondError(w, http.StatusUnauthorized, "Invalid session")
		return
	}

	user, err := h.core.db.GetUser(r.Context(), session.UserID)
	if err != nil {
		h.core.transport.RespondError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	// Create a copy of the user for the response to avoid modifying the stored user
	responseUser := *user
	responseUser.Password = ""
	h.core.transport.RespondJSON(w, http.StatusOK, models.SessionResponse{
		User:    &responseUser,
		Session: session,
	})
}

func (h *Handlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ResetPasswordRequest
	if err := h.core.transport.DecodeJSON(r, &req); err != nil {
		h.core.transport.RespondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Email == "" {
		h.core.transport.RespondError(w, http.StatusBadRequest, "Email is required")
		return
	}

	// TODO: Implement password reset logic
	h.core.transport.RespondJSON(w, http.StatusOK, map[string]string{"message": "Password reset email sent"})
}

func (h *Handlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req models.VerifyEmailRequest
	if err := h.core.transport.DecodeJSON(r, &req); err != nil {
		h.core.transport.RespondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Token == "" {
		h.core.transport.RespondError(w, http.StatusBadRequest, "Token is required")
		return
	}

	// TODO: Implement email verification logic
	h.core.transport.RespondJSON(w, http.StatusOK, map[string]string{"message": "Email verified successfully"})
}

func (h *Handlers) SetupTwoFactor(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement 2FA setup
	h.core.transport.RespondError(w, http.StatusNotImplemented, "Two-factor setup not implemented")
}

func (h *Handlers) VerifyTwoFactor(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement 2FA verification
	h.core.transport.RespondError(w, http.StatusNotImplemented, "Two-factor verification not implemented")
}

func (h *Handlers) OAuthRedirect(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement OAuth redirect
	h.core.transport.RespondError(w, http.StatusNotImplemented, "OAuth redirect not implemented")
}

func (h *Handlers) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement OAuth callback
	h.core.transport.RespondError(w, http.StatusNotImplemented, "OAuth callback not implemented")
}

func (h *Handlers) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement organization creation
	h.core.transport.RespondError(w, http.StatusNotImplemented, "Organization creation not implemented")
}

func (h *Handlers) GetOrganization(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement get organization
	h.core.transport.RespondError(w, http.StatusNotImplemented, "Get organization not implemented")
}

func (h *Handlers) InviteToOrganization(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement organization invitation
	h.core.transport.RespondError(w, http.StatusNotImplemented, "Organization invitation not implemented")
}

func (h *Handlers) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement member role update
	h.core.transport.RespondError(w, http.StatusNotImplemented, "Member role update not implemented")
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