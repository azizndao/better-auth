// Package dto ..
package dto

import (
	"time"

	"better-auth/internal/models"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	FirstName *string   `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Image     *string   `json:"image"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewUser(u models.User) *User {
	user := &User{
		ID:        u.ID,
		LastName:  u.LastName,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}

	if u.FirstName.Valid {
		user.FirstName = &u.FirstName.String
	}

	if u.Image.Valid {
		user.Image = &u.Image.String
	}

	return user
}

type Session struct {
	ID        uuid.UUID      `json:"id"`
	UserID    uuid.UUID      `json:"userId"`
	Token     string         `json:"token"`
	ExpiresAt time.Time      `json:"expiresAt"`
	IPAddress string         `json:"ipAddress"`
	UserAgent string         `json:"userAgent"`
	Active    bool           `json:"active"`
	Device    string         `json:"device"`
	Data      map[string]any `json:"data"`
}

func NewSession(s models.Session) *Session {
	return &Session{
		ID:        s.ID,
		UserID:    s.UserID,
		Token:     s.Token,
		ExpiresAt: s.ExpiresAt,
		IPAddress: s.IPAddress,
		UserAgent: s.UserAgent,
		Active:    s.Active,
		Device:    s.Device,
		Data:      s.Data,
	}
}

// AuthData contains user information for request context
type AuthData struct {
	*Session
	User *User `json:"user"`
}

func NewAuthData(s models.Session, u models.User) *AuthData {
	return &AuthData{
		Session: NewSession(s),
		User:    NewUser(u),
	}
}

// SignUpPayload represents a sign-up request
type SignUpPayload struct {
	Email               string          `json:"email"                 validate:"required,email"`
	Password            string          `json:"password"              validate:"required,min=8"`
	FirstName           *string         `json:"firstName"`
	LastName            string          `json:"lastName"              validate:"required"`
	EmailVerifyCallback *string         `json:"callbackURL"`
	Metadata            *map[string]any `json:"metadata"`
}

// SignInPayload represents a sign-in request
type SignInPayload struct {
	Email      string `json:"email"      validate:"required,email"`
	Password   string `json:"password"   validate:"required"`
	RememberMe bool   `json:"rememberMe"`
}

// ResetPasswordPayload represents a password reset request
type ResetPasswordPayload struct {
	Email       string `json:"email"                 validate:"required,email"`
	CallbackURL string `json:"callbackURL,omitempty"`
}

// ChangePasswordPayload represents a password change request
type ChangePasswordPayload struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword"     validate:"required,min=8"`
}

// VerifyEmailPayload represents an email verification request
type VerifyEmailPayload struct {
	Token string `json:"token" validate:"required"`
}

// UpdateUserPayload represents a user profile update request
type UpdateUserPayload struct {
	Name     string         `json:"name,omitempty"`
	Image    string         `json:"image,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// CreateOrganizationPayload represents an organization creation request
type CreateOrganizationPayload struct {
	Name     string         `json:"name"               validate:"required"`
	Slug     string         `json:"slug"`
	Logo     string         `json:"logo,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// InviteToOrganizationRequest represents an organization invitation request
type InviteToOrganizationRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role"`
}

// UpdateMemberRoleRequest represents a member role update request
type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	User    *User    `json:"user"`
	Session *Session `json:"session,omitempty"`
}

func NewAuthResponse(u models.User, s models.Session) *AuthResponse {
	return &AuthResponse{
		User:    NewUser(u),
		Session: NewSession(s),
	}
}

// TwoFactorResponse represents a 2FA setup response
type TwoFactorResponse struct {
	Secret      string   `json:"secret"`
	QRCode      string   `json:"qrCode"`
	BackupCodes []string `json:"backupCodes,omitempty"`
}

// SessionResponse represents a session response
type SessionResponse struct {
	User    *User    `json:"user"`
	Session *Session `json:"session"`
}

func NewSessionResponse(u models.User, s models.Session) *SessionResponse {
	return &SessionResponse{
		User:    NewUser(u),
		Session: NewSession(s),
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}
