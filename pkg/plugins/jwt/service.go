package jwt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"better-auth/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// JWTService provides JWT token management functionality
type JWTService struct {
	db     *gorm.DB
	config *JWTConfig
}

// NewJWTService creates a new JWT service
func NewJWTService(db *gorm.DB, config *JWTConfig) *JWTService {
	return &JWTService{
		db:     db,
		config: config,
	}
}

// Initialize implements the PluginService interface
func (s *JWTService) Initialize(db *gorm.DB) error {
	s.db = db

	// Start cleanup routine if enabled
	if s.config.CleanupInterval > 0 {
		go s.startCleanupRoutine()
	}

	return nil
}

// Name implements the PluginService interface
func (s *JWTService) Name() string {
	return "jwt"
}

// GenerateTokenPair generates a new access and refresh token pair
func (s *JWTService) GenerateTokenPair(
	ctx context.Context,
	user *models.User,
	metadata map[string]any,
) (*JWTTokenPair, error) {
	now := time.Now()

	// Generate JWT ID
	jti, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT ID: %w", err)
	}

	// Create claims
	claims := &JWTClaims{
		UserID:           user.ID,
		Email:            user.Email,
		Name:             user.Name,
		Image:            user.Image,
		EmailVerified:    user.EmailVerified,
		TwoFactorEnabled: user.TwoFactorEnabled,
		Metadata:         metadata,
		Issuer:           s.config.Issuer,
		Subject:          user.ID,
		ExpiresAt:        now.Add(s.config.AccessTokenExpiry).Unix(),
		IssuedAt:         now.Unix(),
		JWTID:            jti,
	}

	if s.config.Audience != "" {
		claims.Audience = s.config.Audience
	}

	// Generate access token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(s.config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	tokenPair := &JWTTokenPair{
		AccessToken: accessToken,
		ExpiresAt:   now.Add(s.config.AccessTokenExpiry),
		TokenType:   "Bearer",
	}

	// Generate refresh token if enabled
	if s.config.EnableRefreshTokens {
		refreshToken, err := s.generateRefreshToken(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate refresh token: %w", err)
		}
		tokenPair.RefreshToken = refreshToken
	}

	return tokenPair, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.config.SecretKey, nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is blacklisted
	if s.config.EnableTokenBlacklist {
		if blacklisted, err := s.isTokenBlacklisted(context.Background(), claims.JWTID); err != nil {
			return nil, fmt.Errorf("failed to check blacklist: %w", err)
		} else if blacklisted {
			return nil, fmt.Errorf("token is blacklisted")
		}
	}

	return claims, nil
}

// RefreshToken generates a new access token using a refresh token
func (s *JWTService) RefreshToken(ctx context.Context, refreshToken string) (*JWTTokenPair, error) {
	if !s.config.EnableRefreshTokens {
		return nil, fmt.Errorf("refresh tokens are not enabled")
	}

	// Validate refresh token
	var dbToken JWTRefreshToken
	if err := s.db.WithContext(ctx).First(&dbToken, "token = ? AND revoked = false", refreshToken).Error; err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if token is expired
	if time.Now().After(dbToken.ExpiresAt) {
		return nil, fmt.Errorf("refresh token has expired")
	}

	// Get user
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", dbToken.UserID).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Generate new token pair
	tokenPair, err := s.GenerateTokenPair(ctx, &user, nil)
	if err != nil {
		return nil, err
	}

	// Revoke old refresh token
	dbToken.Revoked = true
	now := time.Now()
	dbToken.RevokedAt = &now
	s.db.WithContext(ctx).Save(&dbToken)

	return tokenPair, nil
}

// RevokeToken revokes a JWT token by adding it to the blacklist
func (s *JWTService) RevokeToken(ctx context.Context, tokenString string) error {
	if !s.config.EnableTokenBlacklist {
		return fmt.Errorf("token blacklist is not enabled")
	}

	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	blacklistEntry := &JWTBlacklist{
		ID:        generateSimpleID(),
		JTI:       claims.JWTID,
		ExpiresAt: time.Unix(claims.ExpiresAt, 0),
		CreatedAt: time.Now(),
	}

	return s.db.WithContext(ctx).Create(blacklistEntry).Error
}

// RevokeRefreshToken revokes a refresh token
func (s *JWTService) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	if !s.config.EnableRefreshTokens {
		return fmt.Errorf("refresh tokens are not enabled")
	}

	var dbToken JWTRefreshToken
	if err := s.db.WithContext(ctx).First(&dbToken, "token = ?", refreshToken).Error; err != nil {
		return fmt.Errorf("refresh token not found")
	}

	dbToken.Revoked = true
	now := time.Now()
	dbToken.RevokedAt = &now

	return s.db.WithContext(ctx).Save(&dbToken).Error
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (s *JWTService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	if !s.config.EnableRefreshTokens {
		return nil
	}

	now := time.Now()
	return s.db.WithContext(ctx).
		Model(&JWTRefreshToken{}).
		Where("user_id = ? AND revoked = false", userID).
		Updates(map[string]any{
			"revoked":    true,
			"revoked_at": now,
			"updated_at": now,
		}).Error
}

// generateRefreshToken generates a new refresh token
func (s *JWTService) generateRefreshToken(ctx context.Context, userID string) (string, error) {
	// Clean up old tokens if limit exceeded
	if s.config.MaxRefreshTokensPerUser > 0 {
		if err := s.cleanupUserRefreshTokens(ctx, userID); err != nil {
			return "", err
		}
	}

	token, err := generateSecureToken()
	if err != nil {
		return "", err
	}

	refreshToken := &JWTRefreshToken{
		ID:        generateSimpleID(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(s.config.RefreshTokenExpiry),
		IssuedAt:  time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(refreshToken).Error; err != nil {
		return "", err
	}

	return token, nil
}

// cleanupUserRefreshTokens removes old refresh tokens if limit exceeded
func (s *JWTService) cleanupUserRefreshTokens(ctx context.Context, userID string) error {
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&JWTRefreshToken{}).
		Where("user_id = ? AND revoked = false", userID).
		Count(&count).Error; err != nil {
		return err
	}

	if count >= int64(s.config.MaxRefreshTokensPerUser) {
		// Delete oldest tokens
		tokensToDelete := count - int64(s.config.MaxRefreshTokensPerUser) + 1
		return s.db.WithContext(ctx).
			Where("user_id = ? AND revoked = false", userID).
			Order("created_at ASC").
			Limit(int(tokensToDelete)).
			Delete(&JWTRefreshToken{}).Error
	}

	return nil
}

// isTokenBlacklisted checks if a token is blacklisted
func (s *JWTService) isTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&JWTBlacklist{}).
		Where("jti = ? AND expires_at > ?", jti, time.Now()).
		Count(&count).Error

	return count > 0, err
}

// startCleanupRoutine starts a background routine to clean up expired tokens
func (s *JWTService) startCleanupRoutine() {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)

		// Clean up expired refresh tokens
		s.db.WithContext(ctx).
			Where("expires_at < ? OR (revoked = true AND revoked_at < ?)",
				time.Now(), time.Now().Add(-24*time.Hour)).
			Delete(&JWTRefreshToken{})

		// Clean up expired blacklist entries
		s.db.WithContext(ctx).
			Where("expires_at < ?", time.Now()).
			Delete(&JWTBlacklist{})

		cancel()
	}
}

// Helper functions

func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateSimpleID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

