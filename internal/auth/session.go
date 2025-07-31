package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"maps"
	"net/http"
	"time"

	"better-auth/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SessionService handles session management
type SessionService struct {
	db     *gorm.DB
	expiry time.Duration
	secure bool
	domain string
}

// SessionOptions configures session behavior
type SessionOptions struct {
	Expiry time.Duration
	Secure bool
	Domain string
}

// NewSessionService creates a new session service
func NewSessionService(db *gorm.DB, opts SessionOptions) *SessionService {
	if opts.Expiry == 0 {
		opts.Expiry = 24 * time.Hour // Default 24 hours
	}

	return &SessionService{
		db:     db,
		expiry: opts.Expiry,
		secure: opts.Secure,
		domain: opts.Domain,
	}
}

// CreateSession creates a new session for a user
func (s *SessionService) CreateSession(
	ctx context.Context,
	userID, ipAddress, userAgent string,
) (*models.Session, error) {
	token, err := s.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %v", err)
	}

	now := time.Now()
	session := &models.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: now.Add(s.expiry),
		CreatedAt: now,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Active:    true,
		Data:      make(map[string]any),
	}

	if err := s.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	return session, nil
}

// GetSession retrieves a session by token
func (s *SessionService) GetSession(ctx context.Context, token string) (*models.Session, error) {
	var session models.Session
	if err := s.db.WithContext(ctx).First(&session, "token = ?", token).Error; err != nil {
		return nil, fmt.Errorf("session not found: %v", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		s.db.WithContext(ctx).
			Delete(&models.Session{}, "token = ?", token)
			// Clean up expired session
		return nil, fmt.Errorf("session has expired")
	}

	// Check if session is active
	if !session.Active {
		return nil, fmt.Errorf("session is inactive")
	}

	return &session, nil
}

// ValidateSession validates a session token and returns the session
func (s *SessionService) ValidateSession(
	ctx context.Context,
	token string,
) (*models.Session, error) {
	if token == "" {
		return nil, fmt.Errorf("session token is required")
	}

	return s.GetSession(ctx, token)
}

// RefreshSession extends the session expiry time
func (s *SessionService) RefreshSession(
	ctx context.Context,
	token string,
) (*models.Session, error) {
	session, err := s.GetSession(ctx, token)
	if err != nil {
		return nil, err
	}

	// Extend expiry time
	session.ExpiresAt = time.Now().Add(s.expiry)

	// Update session in database
	if err := s.db.WithContext(ctx).Save(session).Error; err != nil {
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}

	return session, nil
}

// DeactivateSession marks a session as inactive
func (s *SessionService) DeactivateSession(ctx context.Context, token string) error {
	var session models.Session
	if err := s.db.WithContext(ctx).First(&session, "token = ?", token).Error; err != nil {
		return fmt.Errorf("session not found: %v", err)
	}

	session.Active = false

	if err := s.db.WithContext(ctx).Save(&session).Error; err != nil {
		return fmt.Errorf("failed to deactivate session: %v", err)
	}

	return nil
}

// DeleteSession removes a session
func (s *SessionService) DeleteSession(ctx context.Context, token string) error {
	if err := s.db.WithContext(ctx).Delete(&models.Session{}, "token = ?", token).Error; err != nil {
		return fmt.Errorf("failed to delete session: %v", err)
	}

	return nil
}

// DeleteAllUserSessions removes all sessions for a user
func (s *SessionService) DeleteAllUserSessions(ctx context.Context, userID string) error {
	// This would require a new database method to delete by user ID
	// For now, we'll implement it in a simple way
	return fmt.Errorf("not implemented: delete all user sessions")
}

// SetSessionCookie sets the session cookie in HTTP response
func (s *SessionService) SetSessionCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		Domain:   s.domain,
		Expires:  time.Now().Add(s.expiry),
		Secure:   s.secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the session cookie
func (s *SessionService) ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Domain:   s.domain,
		Expires:  time.Unix(0, 0),
		Secure:   s.secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
}

// GetSessionFromRequest extracts session token from HTTP request
func (s *SessionService) GetSessionFromRequest(r *http.Request) string {
	// Try to get from cookie first
	if cookie, err := r.Cookie("session_token"); err == nil {
		return cookie.Value
	}

	// Try to get from Authorization header
	if auth := r.Header.Get("Authorization"); auth != "" {
		const prefix = "Bearer "
		if len(auth) > len(prefix) && auth[:len(prefix)] == prefix {
			return auth[len(prefix):]
		}
	}

	return ""
}

// UpdateSessionData updates session data
func (s *SessionService) UpdateSessionData(
	ctx context.Context,
	token string,
	data map[string]any,
) error {
	session, err := s.GetSession(ctx, token)
	if err != nil {
		return err
	}

	// Merge new data with existing data
	if session.Data == nil {
		session.Data = make(map[string]any)
	}

	maps.Copy(session.Data, data)

	if err := s.db.WithContext(ctx).Save(session).Error; err != nil {
		return fmt.Errorf("failed to update session data: %v", err)
	}

	return nil
}

// GetSessionData retrieves specific data from session
func (s *SessionService) GetSessionData(ctx context.Context, token, key string) (any, error) {
	session, err := s.GetSession(ctx, token)
	if err != nil {
		return nil, err
	}

	if session.Data == nil {
		return nil, fmt.Errorf("no data found for key: %s", key)
	}

	value, exists := session.Data[key]
	if !exists {
		return nil, fmt.Errorf("no data found for key: %s", key)
	}

	return value, nil
}

// CleanupExpiredSessions removes expired sessions from database
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	// This would require a new database method to delete expired sessions
	// For now, we'll return not implemented
	return fmt.Errorf("not implemented: cleanup expired sessions")
}

// generateSecureToken generates a cryptographically secure random token
func (s *SessionService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// IsValidToken checks if a token format is valid
func (s *SessionService) IsValidToken(token string) bool {
	if len(token) == 0 {
		return false
	}

	// Decode to check if it's valid base64
	_, err := base64.URLEncoding.DecodeString(token)
	return err == nil
}
