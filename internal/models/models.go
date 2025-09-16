// Package models defines the data models for the authentication system.
package models

import (
	"database/sql"
	"time"

	"better-auth/pkg/plugins/core"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	core.Model
	Email         string         `gorm:"uniqueIndex;not null" json:"email"`
	EmailVerified bool           `gorm:"default:false"        json:"emailVerified"`
	FirstName     string         `                            json:"firstName,omitempty"`
	LastName      string         `                            json:"lastName,omitempty"`
	Image         sql.NullString `                            json:"image"`
	Password      string         `                            json:"-"`
	Metadata      map[string]any `gorm:"serializer:json"      json:"metadata,omitempty"`
	BannedAt      sql.NullTime   `                            json:"bannedAt"`
	BannedUtil    sql.NullTime   `                            json:"bannedUntil"`
	BanReason     sql.NullString `                            json:"banReason"`
}

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	if u.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

// Session represents a user session
type Session struct {
	core.Model
	UserID    uuid.UUID      `gorm:"not null;index"       json:"userId"`
	Token     string         `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time      `gorm:"not null"             json:"expiresAt"`
	IPAddress string         `                            json:"ipAddress,omitempty"`
	UserAgent string         `                            json:"userAgent,omitempty"`
	Active    bool           `gorm:"default:true"         json:"active"`
	Device    string         `                            json:"device,omitempty"`
	Data      map[string]any `gorm:"serializer:json"      json:"data,omitempty"`
}

// Organization represents an organization
type Organization struct {
	core.Model
	Name     string         `gorm:"not null"             json:"name"`
	Slug     string         `gorm:"uniqueIndex;not null" json:"slug"`
	Logo     string         `                            json:"logo,omitempty"`
	Metadata map[string]any `gorm:"serializer:json"      json:"metadata,omitempty"`
}

// OAuthAccount represents an OAuth account link
type OAuthAccount struct {
	core.Model
	UserID       string         `gorm:"not null;index"  json:"userId"`
	Provider     string         `gorm:"not null"        json:"provider"`
	ProviderID   string         `gorm:"not null"        json:"providerId"`
	Email        string         `                       json:"email,omitempty"`
	AccessToken  string         `                       json:"-"`
	RefreshToken string         `                       json:"-"`
	TokenType    string         `                       json:"tokenType,omitempty"`
	Scope        string         `                       json:"scope,omitempty"`
	AccountData  map[string]any `gorm:"serializer:json" json:"accountData,omitempty"`
}

// AuthData contains user information for request context
type AuthData struct {
	*Session
	User         *User         `json:"user"`
	Organization *Organization `json:"organization"`
}

// SignUpPayload represents a sign-up request
type SignUpPayload struct {
	Email               string         `json:"email"                 validate:"required,email"`
	Password            string         `json:"password"              validate:"required,min=8"`
	FirstName           string         `json:"firstName"`
	LastName            string         `json:"lastName"`
	EmailVerifyCallback *string        `json:"callbackURL"`
	Metadata            map[string]any `json:"metadata"`
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

// TwoFactorSetupPayload represents a 2FA setup request
type TwoFactorSetupPayload struct {
	Password string `json:"password" validate:"required"`
}

// TwoFactorVerifyPayload represents a 2FA verification request
type TwoFactorVerifyPayload struct {
	Code string `json:"code" validate:"required,len=6"`
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

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
	Value   any    `json:"value,omitempty"`
}

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Code    int               `json:"code"`
	Type    string            `json:"type"`
	Errors  []ValidationError `json:"errors"`
	Message string            `json:"message"`
}
