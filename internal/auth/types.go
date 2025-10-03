package auth

import "better-auth/internal/models"

// Type aliases for backward compatibility

type (
	User         = models.User
	Session      = models.Session
	Organization = models.Organization
	OAuthAccount = models.Account
	UserContext  = models.AuthData
)

// Request/Response types

type (
	SignUpRequest               = models.SignUpPayload
	SignInRequest               = models.SignInPayload
	ResetPasswordRequest        = models.ResetPasswordPayload
	ChangePasswordRequest       = models.ChangePasswordPayload
	VerifyEmailRequest          = models.VerifyEmailPayload
	TwoFactorSetupRequest       = models.TwoFactorSetupPayload
	TwoFactorVerifyRequest      = models.TwoFactorVerifyPayload
	UpdateUserRequest           = models.UpdateUserPayload
	CreateOrganizationRequest   = models.CreateOrganizationPayload
	InviteToOrganizationRequest = models.InviteToOrganizationRequest
	UpdateMemberRoleRequest     = models.UpdateMemberRoleRequest
)

// Response types

type (
	AuthResponse      = models.AuthResponse
	TwoFactorResponse = models.TwoFactorResponse
	SessionResponse   = models.SessionResponse
	ErrorResponse     = models.ErrorResponse
)
