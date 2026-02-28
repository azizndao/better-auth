package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"better-auth/internal/dto"
	"better-auth/internal/models"
	"better-auth/pkg/transport"

	"github.com/azizndao/grouter"
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

func (h *handlers) registerRoutes(r grouter.RouteGroup) {
	r.Post("/signup", h.signUp)
	r.Post("/signin", h.signIn)
	r.Post("/signout", h.signOut)
	r.Get("/session", h.GetSession)
	r.Post("/reset-password", h.resetPassword)
	r.Post("/verify-email", h.verifyEmail)
}

func (h *handlers) signUp(c *grouter.Ctx) error {
	var req dto.SignUpPayload
	if err := h.DecodeJSON(c, &req); err != nil {
		return err
	}

	existingUser, _ := h.core.GetUserByEmail(c.Context(), req.Email)
	if existingUser != nil {
		return grouter.ErrorConflict(nil, fmt.Errorf("user already exists"))
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

	ctx := c.Context()

	return h.core.database.Transaction(func(tx *gorm.DB) error {
		err := gorm.G[models.User](tx).Create(ctx, user)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return grouter.ErrorNotFound(nil, fmt.Errorf("user not found"))
			}
			return grouter.ErrorInternalServerError(nil, fmt.Errorf("failed to create user"))
		}

		session, err := h.core.session.CreateSession(
			ctx,
			user.ID,
			c.IP(),
			c.Request.UserAgent(),
			tx,
		)
		if err != nil {
			return grouter.ErrorInternalServerError("Failed to create session", err)
		}

		h.core.session.SetSessionCookie(c.Response, session.Token)

		return c.JSON(dto.NewAuthResponse(*user, *session))
	})
}

func (h *handlers) signIn(c *grouter.Ctx) error {
	var req dto.SignInPayload
	if err := h.DecodeJSON(c, &req); err != nil {
		return err
	}

	user, err := h.core.GetUserByEmail(c.Context(), req.Email)
	if err != nil || user.BannedUtil.Time.After(time.Now()) {
		return grouter.ErrorUnauthorized("Invalid credentials", err)
	}

	if err := user.CheckPassword(req.Password); err != nil {
		return grouter.ErrorUnauthorized("Invalid credentials", err)
	}

	session, err := h.core.session.CreateSession(
		c.Context(),
		user.ID,
		c.IP(),
		c.Request.UserAgent(),
		nil,
	)
	if err != nil {
		return grouter.ErrorInternalServerError("Failed to create session", err)
	}

	h.core.session.SetSessionCookie(c.Response, session.Token)

	// Create a copy of the user for the response to avoid modifying the stored user
	return c.JSON(dto.NewAuthResponse(*user, *session))
}

func (h *handlers) signOut(c *grouter.Ctx) error {
	token := h.core.session.GetSessionFromRequest(c)
	if token == "" {
		return grouter.ErrorBadRequest("No session token provided", nil)
	}

	if err := h.core.session.DeleteSession(c.Context(), token); err != nil {
		return grouter.ErrorInternalServerError("Failed to delete session", err)
	}

	h.core.session.ClearSessionCookie(c.Response)
	return c.JSON(map[string]string{"message": "Signed out successfully"})
}

func (h *handlers) GetSession(c *grouter.Ctx) error {
	token := h.core.session.GetSessionFromRequest(c)
	if token == "" {
		return grouter.ErrorUnauthorized("No session token provided", nil)
	}

	session, err := h.core.session.ValidateSession(c.Context(), token)
	if err != nil {
		return grouter.ErrorUnauthorized("Invalid session", err)
	}

	user, err := h.core.GetUser(c.Context(), session.UserID)
	if err != nil {
		return grouter.ErrorInternalServerError("Failed to get user", err)
	}

	return c.JSON(dto.NewSessionResponse(*user, *session))
}

func (h *handlers) resetPassword(c *grouter.Ctx) error {
	var req dto.ResetPasswordPayload
	if err := h.DecodeJSON(c, &req); err != nil {
		return err
	}

	// TODO: Implement password reset logic
	return c.JSON(map[string]string{"message": "Password reset email sent"})
}

func (h *handlers) verifyEmail(c *grouter.Ctx) error {
	var req dto.VerifyEmailPayload
	if err := h.DecodeJSON(c, &req); err != nil {
		return err
	}

	// TODO: Implement email verification logic
	return c.JSON(map[string]string{"message": "Email verified successfully"})
}
