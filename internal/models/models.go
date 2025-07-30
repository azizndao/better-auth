package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID                 string                 `json:"id"`
	Email              string                 `json:"email"`
	EmailVerified      bool                   `json:"emailVerified"`
	Name               string                 `json:"name"`
	Image              string                 `json:"image,omitempty"`
	Password           string                 `json:"-"`
	TwoFactorEnabled   bool                   `json:"twoFactorEnabled"`
	TwoFactorSecret    string                 `json:"-"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	LastSignIn         *time.Time             `json:"lastSignIn,omitempty"`
	SignInCount        int                    `json:"signInCount"`
	Blocked            bool                   `json:"blocked"`
	EmailVerifyToken   string                 `json:"-"`
	PasswordResetToken string                 `json:"-"`
}

// Session represents a user session
type Session struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"userId"`
	Token     string                 `json:"token"`
	ExpiresAt time.Time              `json:"expiresAt"`
	CreatedAt time.Time              `json:"createdAt"`
	IPAddress string                 `json:"ipAddress,omitempty"`
	UserAgent string                 `json:"userAgent,omitempty"`
	Active    bool                   `json:"active"`
	Device    string                 `json:"device,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Organization represents an organization
type Organization struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	Logo      string                 `json:"logo,omitempty"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// OAuthAccount represents an OAuth account link
type OAuthAccount struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"userId"`
	Provider      string                 `json:"provider"`
	ProviderID    string                 `json:"providerId"`
	Email         string                 `json:"email,omitempty"`
	AccessToken   string                 `json:"-"`
	RefreshToken  string                 `json:"-"`
	ExpiresAt     *time.Time             `json:"expiresAt,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
	TokenType     string                 `json:"tokenType,omitempty"`
	Scope         string                 `json:"scope,omitempty"`
	AccountData   map[string]interface{} `json:"accountData,omitempty"`
}

// UserContext contains user information for request context
type UserContext struct {
	User         *User         `json:"user"`
	Session      *Session      `json:"session,omitempty"`
	Organization *Organization `json:"organization,omitempty"`
}

// SignUpRequest represents a sign-up request
type SignUpRequest struct {
	Email                string                 `json:"email" validate:"required,email"`
	Password             string                 `json:"password" validate:"required,min=8"`
	Name                 string                 `json:"name"`
	Image                string                 `json:"image,omitempty"`
	EmailVerifyCallback  string                 `json:"callbackURL,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// SignInRequest represents a sign-in request
type SignInRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	RememberMe bool   `json:"rememberMe"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	CallbackURL string `json:"callbackURL,omitempty"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

// VerifyEmailRequest represents an email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// TwoFactorSetupRequest represents a 2FA setup request
type TwoFactorSetupRequest struct {
	Password string `json:"password" validate:"required"`
}

// TwoFactorVerifyRequest represents a 2FA verification request
type TwoFactorVerifyRequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

// UpdateUserRequest represents a user profile update request
type UpdateUserRequest struct {
	Name     string                 `json:"name,omitempty"`
	Image    string                 `json:"image,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CreateOrganizationRequest represents an organization creation request
type CreateOrganizationRequest struct {
	Name     string                 `json:"name" validate:"required"`
	Slug     string                 `json:"slug"`
	Logo     string                 `json:"logo,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
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
	User    *User     `json:"user"`
	Session *Session  `json:"session,omitempty"`
	Token   string    `json:"token,omitempty"`
	Error   string    `json:"error,omitempty"`
}

// TwoFactorResponse represents a 2FA setup response
type TwoFactorResponse struct {
	Secret   string `json:"secret"`
	QRCode   string `json:"qrCode"`
	BackupCodes []string `json:"backupCodes,omitempty"`
}

// SessionResponse represents a session response
type SessionResponse struct {
	User    *User    `json:"user"`
	Session *Session `json:"session"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}