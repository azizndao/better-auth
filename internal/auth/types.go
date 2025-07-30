package auth

import "better-auth/internal/models"

// Type aliases for backward compatibility
type User = models.User
type Session = models.Session
type Organization = models.Organization
type OAuthAccount = models.OAuthAccount
type UserContext = models.UserContext

// Request/Response types
type SignUpRequest = models.SignUpRequest
type SignInRequest = models.SignInRequest
type ResetPasswordRequest = models.ResetPasswordRequest
type ChangePasswordRequest = models.ChangePasswordRequest
type VerifyEmailRequest = models.VerifyEmailRequest
type TwoFactorSetupRequest = models.TwoFactorSetupRequest
type TwoFactorVerifyRequest = models.TwoFactorVerifyRequest
type UpdateUserRequest = models.UpdateUserRequest
type CreateOrganizationRequest = models.CreateOrganizationRequest
type InviteToOrganizationRequest = models.InviteToOrganizationRequest
type UpdateMemberRoleRequest = models.UpdateMemberRoleRequest

// Response types
type AuthResponse = models.AuthResponse
type TwoFactorResponse = models.TwoFactorResponse
type SessionResponse = models.SessionResponse
type ErrorResponse = models.ErrorResponse