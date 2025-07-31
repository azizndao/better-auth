package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// JWTRefreshToken represents a JWT refresh token stored in the database
type JWTRefreshToken struct {
	ID        string         `gorm:"primaryKey"           json:"id"`
	UserID    string         `gorm:"not null;index"       json:"userId"`
	Token     string         `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time      `gorm:"not null"             json:"expiresAt"`
	IssuedAt  time.Time      `gorm:"not null"             json:"issuedAt"`
	IPAddress string         `                            json:"ipAddress,omitempty"`
	UserAgent string         `                            json:"userAgent,omitempty"`
	Revoked   bool           `gorm:"default:false"        json:"revoked"`
	RevokedAt *time.Time     `                            json:"revokedAt,omitempty"`
	CreatedAt time.Time      `                            json:"createdAt"`
	UpdatedAt time.Time      `                            json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"                json:"-"`
}

// JWTBlacklist represents blacklisted JWT tokens
type JWTBlacklist struct {
	ID        string         `gorm:"primaryKey"           json:"id"`
	JTI       string         `gorm:"uniqueIndex;not null" json:"jti"` // JWT ID
	ExpiresAt time.Time      `gorm:"not null;index"       json:"expiresAt"`
	CreatedAt time.Time      `                            json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"                json:"-"`
}

// JWTClaims represents the claims stored in a JWT token
type JWTClaims struct {
	UserID           string         `json:"userId"`
	Email            string         `json:"email"`
	Name             string         `json:"name"`
	Image            string         `json:"image,omitempty"`
	EmailVerified    bool           `json:"emailVerified"`
	TwoFactorEnabled bool           `json:"twoFactorEnabled"`
	OrganizationID   string         `json:"organizationId,omitempty"`
	TeamID           string         `json:"teamId,omitempty"`
	Role             string         `json:"role,omitempty"`
	Permissions      []string       `json:"permissions,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`

	// Standard JWT claims
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	Audience  string `json:"aud,omitempty"`
	ExpiresAt int64  `json:"exp"`
	NotBefore int64  `json:"nbf,omitempty"`
	IssuedAt  int64  `json:"iat"`
	JWTID     string `json:"jti"`
}

// GetExpirationTime implements jwt.Claims interface
func (c *JWTClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	if c.ExpiresAt == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(c.ExpiresAt, 0)), nil
}

// GetIssuedAt implements jwt.Claims interface
func (c *JWTClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	if c.IssuedAt == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(c.IssuedAt, 0)), nil
}

// GetNotBefore implements jwt.Claims interface
func (c *JWTClaims) GetNotBefore() (*jwt.NumericDate, error) {
	if c.NotBefore == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(c.NotBefore, 0)), nil
}

// GetIssuer implements jwt.Claims interface
func (c *JWTClaims) GetIssuer() (string, error) {
	return c.Issuer, nil
}

// GetSubject implements jwt.Claims interface
func (c *JWTClaims) GetSubject() (string, error) {
	return c.Subject, nil
}

// GetAudience implements jwt.Claims interface
func (c *JWTClaims) GetAudience() (jwt.ClaimStrings, error) {
	if c.Audience == "" {
		return nil, nil
	}
	return jwt.ClaimStrings{c.Audience}, nil
}

// JWTTokenPair represents an access and refresh token pair
type JWTTokenPair struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	TokenType    string    `json:"tokenType"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	SecretKey               []byte        `json:"-"`
	AccessTokenExpiry       time.Duration `json:"accessTokenExpiry"`
	RefreshTokenExpiry      time.Duration `json:"refreshTokenExpiry"`
	Issuer                  string        `json:"issuer"`
	Audience                string        `json:"audience,omitempty"`
	EnableRefreshTokens     bool          `json:"enableRefreshTokens"`
	EnableTokenBlacklist    bool          `json:"enableTokenBlacklist"`
	CleanupInterval         time.Duration `json:"cleanupInterval"`
	MaxRefreshTokensPerUser int           `json:"maxRefreshTokensPerUser"`
}

// TableName overrides the table name for JWTRefreshToken
func (JWTRefreshToken) TableName() string {
	return "jwt_refresh_tokens"
}

// TableName overrides the table name for JWTBlacklist
func (JWTBlacklist) TableName() string {
	return "jwt_blacklist"
}

