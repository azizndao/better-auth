package jwt

import (
	"errors"
	"net/http"

	"better-auth/pkg/transport"
)

// JWTHandlers provides HTTP handlers for JWT operations
type JWTHandlers struct {
	service *JWTService
	transport.Transport
}

// NewJWTHandlers creates new JWT handlers
func NewJWTHandlers(service *JWTService, transport transport.Transport) *JWTHandlers {
	return &JWTHandlers{
		service:   service,
		Transport: transport,
	}
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// RevokeTokenRequest represents a token revocation request
type RevokeTokenRequest struct {
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	RevokeAll    bool   `json:"revokeAll,omitempty"`
}

// RefreshToken handles token refresh requests
func (h *JWTHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		var validationErr *transport.ValidationError
		if errors.As(err, &validationErr) {
			h.RespondValidationError(w, validationErr)
		} else {
			h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		}
		return
	}

	if req.RefreshToken == "" {
		h.RespondError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	tokenPair, err := h.service.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		h.RespondError(w, http.StatusUnauthorized, "Failed to refresh token")
		return
	}

	h.RespondJSON(w, http.StatusOK, tokenPair)
}

// RevokeToken handles token revocation requests
func (h *JWTHandlers) RevokeToken(w http.ResponseWriter, r *http.Request) {
	var req RevokeTokenRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		var validationErr *transport.ValidationError
		if errors.As(err, &validationErr) {
			h.RespondValidationError(w, validationErr)
		} else {
			h.RespondError(w, http.StatusBadRequest, "Invalid request body")
		}
		return
	}

	// Get user from context
	userCtx, ok := h.GetUserContext(r)
	if !ok {
		h.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user := userCtx.User

	if req.RevokeAll {
		// Revoke all user tokens
		if err := h.service.RevokeAllUserTokens(r.Context(), user.ID); err != nil {
			h.RespondError(w, http.StatusInternalServerError, "Failed to revoke tokens")
			return
		}
	} else {
		// Revoke specific tokens
		if req.Token != "" {
			if err := h.service.RevokeToken(r.Context(), req.Token); err != nil {
				h.RespondError(w, http.StatusBadRequest, "Failed to revoke access token")
				return
			}
		}

		if req.RefreshToken != "" {
			if err := h.service.RevokeRefreshToken(r.Context(), req.RefreshToken); err != nil {
				h.RespondError(w, http.StatusBadRequest, "Failed to revoke refresh token")
				return
			}
		}
	}

	h.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Token(s) revoked successfully",
	})
}

// ValidateToken handles token validation requests
func (h *JWTHandlers) ValidateToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.RespondError(w, http.StatusBadRequest, "Authorization header required")
		return
	}

	// Extract token from "Bearer <token>"
	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		h.RespondError(w, http.StatusBadRequest, "Invalid authorization header format")
		return
	}

	tokenString := authHeader[len(bearerPrefix):]
	claims, err := h.service.ValidateToken(tokenString)
	if err != nil {
		h.RespondError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	h.RespondJSON(w, http.StatusOK, map[string]any{
		"valid":  true,
		"claims": claims,
	})
}

// GetTokenInfo returns information about the current token
func (h *JWTHandlers) GetTokenInfo(w http.ResponseWriter, r *http.Request) {
	// Get user from context (assuming JWT middleware sets this)
	_, ok := h.GetUserContext(r)
	if !ok {
		h.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	claimsCtx := r.Context().Value("claims")
	if claimsCtx == nil {
		h.RespondError(w, http.StatusBadRequest, "No token claims found")
		return
	}

	claims := claimsCtx.(*JWTClaims)

	h.RespondJSON(w, http.StatusOK, map[string]any{
		"userId":    claims.UserID,
		"email":     claims.Email,
		"name":      claims.Name,
		"expiresAt": claims.ExpiresAt,
		"issuedAt":  claims.IssuedAt,
		"issuer":    claims.Issuer,
		"jti":       claims.JWTID,
		"metadata":  claims.Metadata,
	})
}
