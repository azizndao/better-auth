package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"better-auth/internal/models"
)

// JWTService handles JWT token generation and validation
type JWTService struct {
	secret []byte
	issuer string
	expiry time.Duration
}

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID    string         `json:"sub"`
	Email     string         `json:"email"`
	Name      string         `json:"name"`
	Role      string         `json:"role,omitempty"`
	IssuedAt  int64          `json:"iat"`
	ExpiresAt int64          `json:"exp"`
	Issuer    string         `json:"iss"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// NewJWTService creates a new JWT service
func NewJWTService(secret []byte, issuer string, expiry time.Duration) *JWTService {
	if len(secret) == 0 {
		secret = make([]byte, 32)
		rand.Read(secret)
	}

	return &JWTService{
		secret: secret,
		issuer: issuer,
		expiry: expiry,
	}
}

// GenerateToken generates a JWT token for a user
func (j *JWTService) GenerateToken(user *models.User) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Name:      user.Name,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(j.expiry).Unix(),
		Issuer:    j.issuer,
		Metadata:  user.Metadata,
	}

	header := map[string]any{
		"typ": "JWT",
		"alg": "HS256",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %v", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %v", err)
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	message := headerEncoded + "." + claimsEncoded
	signature := j.sign(message)

	return message + "." + signature, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerEncoded, claimsEncoded, signature := parts[0], parts[1], parts[2]
	message := headerEncoded + "." + claimsEncoded

	if !j.verify(message, signature) {
		return nil, fmt.Errorf("invalid token signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %v", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %v", err)
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token has expired")
	}

	return &claims, nil
}

// RefreshToken generates a new token with extended expiry
func (j *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("invalid token for refresh: %v", err)
	}

	now := time.Now()
	claims.IssuedAt = now.Unix()
	claims.ExpiresAt = now.Add(j.expiry).Unix()

	header := map[string]any{
		"typ": "JWT",
		"alg": "HS256",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %v", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %v", err)
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	message := headerEncoded + "." + claimsEncoded
	signature := j.sign(message)

	return message + "." + signature, nil
}

// sign creates a signature for the message
func (j *JWTService) sign(message string) string {
	h := hmac.New(sha256.New, j.secret)
	h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// verify checks if the signature is valid for the message
func (j *JWTService) verify(message, signature string) bool {
	expectedSignature := j.sign(message)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// ExtractUserID extracts user ID from token without full validation
func (j *JWTService) ExtractUserID(tokenString string) (string, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode claims: %v", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return "", fmt.Errorf("failed to unmarshal claims: %v", err)
	}

	return claims.UserID, nil
}

