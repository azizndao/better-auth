package auth

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"better-auth/internal/dto"
	"better-auth/internal/models"
	"better-auth/pkg/router"
	"better-auth/pkg/transport"

	"gorm.io/gorm"
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
	var req dto.SignUpPayload
	if err := h.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, err)
		return
	}

	existingUser, _ := h.core.GetUserByEmail(r.Context(), req.Email)
	if existingUser != nil {
		h.RespondError(w, transport.NewAPIError(http.StatusConflict, "User already exists", nil))
		return
	}

	user := &models.User{
		Email:         req.Email,
		LastName:      req.LastName,
		EmailVerified: false,
		Metadata:      req.Metadata,
	}

	if req.FirstName != nil {
		user.FirstName = sql.NullString{String: *req.FirstName, Valid: true}
	}

	ctx := r.Context()

	_ = h.core.database.Transaction(func(tx *gorm.DB) error {
		err := gorm.G[models.User](tx).Create(ctx, user)
		if err != nil {
			h.RespondError(w, transport.NewAPIError(http.StatusInternalServerError, "Failed to create user", err))
			return err
		}

		session, err := h.core.session.CreateSession(
			ctx,
			user.ID,
			getClientIP(r),
			r.UserAgent(),
			tx,
		)
		if err != nil {
			h.RespondError(w, transport.NewAPIError(http.StatusInternalServerError, "Failed to create session", err))
			return err
		}

		h.core.session.SetSessionCookie(w, session.Token)

		h.RespondJSON(w, http.StatusCreated, dto.NewAuthResponse(*user, *session))
		return nil
	})
}

func (h *handlers) signIn(w http.ResponseWriter, r *http.Request) {
	var req dto.SignInPayload
	if err := h.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, err)
		return
	}

	user, err := h.core.GetUserByEmail(r.Context(), req.Email)
	if err != nil || user.BannedUtil.Time.After(time.Now()) {
		h.RespondError(w, transport.NewUnauthorizedError("Invalid credentials", err))
		return
	}

	if err := user.CheckPassword(req.Password); err != nil {
		h.RespondError(w, transport.NewUnauthorizedError("Invalid credentials", err))
		return
	}

	session, err := h.core.session.CreateSession(
		r.Context(),
		user.ID,
		getClientIP(r),
		r.UserAgent(),
		nil,
	)
	if err != nil {
		h.RespondError(w, transport.NewInternalServerError("Failed to create session", err))
		return
	}

	h.core.session.SetSessionCookie(w, session.Token)

	// Create a copy of the user for the response to avoid modifying the stored user
	h.RespondJSON(w, http.StatusOK, dto.NewAuthResponse(*user, *session))
}

func (h *handlers) signOut(w http.ResponseWriter, r *http.Request) {
	token := h.core.session.GetSessionFromRequest(r)
	if token == "" {
		h.RespondError(w, transport.NewBadRequestError("No session token provided", nil))
		return
	}

	if err := h.core.session.DeleteSession(r.Context(), token); err != nil {
		h.RespondError(w, transport.NewInternalServerError("Failed to delete session", err))
		return
	}

	h.core.session.ClearSessionCookie(w)
	h.RespondJSON(w, http.StatusOK, map[string]string{"message": "Signed out successfully"})
}

func (h *handlers) GetSession(w http.ResponseWriter, r *http.Request) {
	token := h.core.session.GetSessionFromRequest(r)
	if token == "" {
		h.RespondError(w, transport.NewUnauthorizedError("No session token provided", nil))
		return
	}

	session, err := h.core.session.ValidateSession(r.Context(), token)
	if err != nil {
		h.RespondError(w, transport.NewUnauthorizedError("Invalid session", err))
		return
	}

	user, err := h.core.GetUser(r.Context(), session.UserID)
	if err != nil {
		h.RespondError(w, transport.NewInternalServerError("Failed to get user", err))
		return
	}

	h.RespondJSON(w, http.StatusOK, dto.NewSessionResponse(*user, *session))
}

func (h *handlers) resetPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ResetPasswordPayload
	if err := h.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, err)
		return
	}

	// TODO: Implement password reset logic
	h.RespondJSON(w, http.StatusOK, map[string]string{"message": "Password reset email sent"})
}

func (h *handlers) verifyEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyEmailPayload
	if err := h.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, err)
		return
	}

	// TODO: Implement email verification logic
	h.RespondJSON(w, http.StatusOK, map[string]string{"message": "Email verified successfully"})
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
