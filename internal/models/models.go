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

type User struct {
	core.Model
	Email         string          `gorm:"uniqueIndex;not null" json:"email"`
	EmailVerified bool            `gorm:"default:false"        json:"emailVerified"`
	FirstName     sql.NullString  `                            json:"firstName"`
	LastName      string          `                            json:"lastName,omitempty"`
	Image         sql.NullString  `                            json:"image"`
	Password      sql.NullString  `                       json:"-"`
	Metadata      *map[string]any `gorm:"serializer:json"      json:"metadata,omitempty"`
	BannedAt      sql.NullTime    `                            json:"bannedAt"`
	BannedUtil    sql.NullTime    `                            json:"bannedUntil"`
	BanReason     sql.NullString  `                            json:"banReason"`
}

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	if u.Password.String == "" {
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password.String), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = sql.NullString{String: string(hashedPassword), Valid: true}

	return nil
}

func (u *User) CheckPassword(password string) error {
	if !u.Password.Valid {
		return nil
	}
	return bcrypt.CompareHashAndPassword([]byte(u.Password.String), []byte(password))
}

// Session represents a user session
type Session struct {
	core.Model
	UserID    uuid.UUID      `gorm:"not null;index" json:"userId"`
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

// Account represents an OAuth account link
type Account struct {
	core.Model
	UserID       uuid.UUID      `gorm:"not null;index"  json:"userId"`
	Provider     string         `gorm:"not null"        json:"provider"`
	ProviderID   string         `gorm:"not null"        json:"providerId"`
	Email        string         `                       json:"email,omitempty"`
	AccessToken  sql.NullString `                       json:"-"`
	RefreshToken sql.NullString `                       json:"-"`
	TokenType    string         `                       json:"tokenType,omitempty"`
	Scope        string         `                       json:"scope,omitempty"`
	AccountData  map[string]any `gorm:"serializer:json" json:"accountData,omitempty"`
}
