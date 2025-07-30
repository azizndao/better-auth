package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"better-auth/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDB implements Database interface for SQLite
type SQLiteDB struct {
	connectionURL string
	db            *sql.DB
}

func (db *SQLiteDB) migrate() error {
	var err error
	db.db, err = sql.Open("sqlite3", db.connectionURL)
	if err != nil {
		return err
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			email_verified BOOLEAN DEFAULT 0,
			name TEXT,
			image TEXT,
			password TEXT,
			two_factor_enabled BOOLEAN DEFAULT 0,
			two_factor_secret TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			metadata TEXT,
			last_sign_in DATETIME,
			sign_in_count INTEGER DEFAULT 0,
			blocked BOOLEAN DEFAULT 0,
			email_verify_token TEXT,
			password_reset_token TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token TEXT UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			ip_address TEXT,
			user_agent TEXT,
			active BOOLEAN DEFAULT 1,
			data TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS organizations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			logo TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			metadata TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
	}

	for _, query := range queries {
		if _, err := db.db.Exec(query); err != nil {
			return fmt.Errorf("migration failed: %v", err)
		}
	}
	return nil
}

func (db *SQLiteDB) CreateUser(ctx context.Context, user *models.User) error {
	metadataJSON, _ := json.Marshal(user.Metadata)

	_, err := db.db.ExecContext(ctx, `
		INSERT INTO users (id, email, email_verified, name, image, password, two_factor_enabled, 
			two_factor_secret, created_at, updated_at, metadata, last_sign_in, sign_in_count, 
			blocked, email_verify_token, password_reset_token)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Email, user.EmailVerified, user.Name, user.Image, user.Password,
		user.TwoFactorEnabled, user.TwoFactorSecret, user.CreatedAt, user.UpdatedAt,
		string(metadataJSON), user.LastSignIn, user.SignInCount, user.Blocked,
		user.EmailVerifyToken, user.PasswordResetToken)

	return err
}

func (db *SQLiteDB) GetUser(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	var metadataJSON string

	err := db.db.QueryRowContext(ctx, `
		SELECT id, email, email_verified, name, image, password, two_factor_enabled,
			two_factor_secret, created_at, updated_at, metadata, last_sign_in, 
			sign_in_count, blocked, email_verify_token, password_reset_token
		FROM users WHERE id = ?`, id).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.Name, &user.Image,
		&user.Password, &user.TwoFactorEnabled, &user.TwoFactorSecret,
		&user.CreatedAt, &user.UpdatedAt, &metadataJSON, &user.LastSignIn,
		&user.SignInCount, &user.Blocked, &user.EmailVerifyToken, &user.PasswordResetToken)

	if err != nil {
		return nil, err
	}

	if metadataJSON != "" {
		json.Unmarshal([]byte(metadataJSON), &user.Metadata)
	}

	return user, nil
}

func (db *SQLiteDB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	var metadataJSON string

	err := db.db.QueryRowContext(ctx, `
		SELECT id, email, email_verified, name, image, password, two_factor_enabled,
			two_factor_secret, created_at, updated_at, metadata, last_sign_in, 
			sign_in_count, blocked, email_verify_token, password_reset_token
		FROM users WHERE email = ?`, email).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.Name, &user.Image,
		&user.Password, &user.TwoFactorEnabled, &user.TwoFactorSecret,
		&user.CreatedAt, &user.UpdatedAt, &metadataJSON, &user.LastSignIn,
		&user.SignInCount, &user.Blocked, &user.EmailVerifyToken, &user.PasswordResetToken)

	if err != nil {
		return nil, err
	}

	if metadataJSON != "" {
		json.Unmarshal([]byte(metadataJSON), &user.Metadata)
	}

	return user, nil
}

func (db *SQLiteDB) UpdateUser(ctx context.Context, user *models.User) error {
	metadataJSON, _ := json.Marshal(user.Metadata)

	_, err := db.db.ExecContext(ctx, `
		UPDATE users SET email = ?, email_verified = ?, name = ?, image = ?, 
			password = ?, two_factor_enabled = ?, two_factor_secret = ?, 
			updated_at = ?, metadata = ?, last_sign_in = ?, sign_in_count = ?,
			blocked = ?, email_verify_token = ?, password_reset_token = ?
		WHERE id = ?`,
		user.Email, user.EmailVerified, user.Name, user.Image, user.Password,
		user.TwoFactorEnabled, user.TwoFactorSecret, user.UpdatedAt, string(metadataJSON),
		user.LastSignIn, user.SignInCount, user.Blocked, user.EmailVerifyToken,
		user.PasswordResetToken, user.ID)

	return err
}

func (db *SQLiteDB) DeleteUser(ctx context.Context, id string) error {
	_, err := db.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	return err
}

func (db *SQLiteDB) CreateSession(ctx context.Context, session *models.Session) error {
	dataJSON, _ := json.Marshal(session.Data)

	_, err := db.db.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, token, expires_at, created_at, ip_address, user_agent, active, data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		session.ID, session.UserID, session.Token, session.ExpiresAt, session.CreatedAt,
		session.IPAddress, session.UserAgent, session.Active, string(dataJSON))

	return err
}

func (db *SQLiteDB) GetSession(ctx context.Context, token string) (*models.Session, error) {
	session := &models.Session{}
	var dataJSON string

	err := db.db.QueryRowContext(ctx, `
		SELECT id, user_id, token, expires_at, created_at, ip_address, user_agent, active, data
		FROM sessions WHERE token = ?`, token).Scan(
		&session.ID, &session.UserID, &session.Token, &session.ExpiresAt,
		&session.CreatedAt, &session.IPAddress, &session.UserAgent,
		&session.Active, &dataJSON)

	if err != nil {
		return nil, err
	}

	if dataJSON != "" {
		json.Unmarshal([]byte(dataJSON), &session.Data)
	}

	return session, nil
}

func (db *SQLiteDB) UpdateSession(ctx context.Context, session *models.Session) error {
	dataJSON, _ := json.Marshal(session.Data)

	_, err := db.db.ExecContext(ctx, `
		UPDATE sessions SET expires_at = ?, ip_address = ?, user_agent = ?, active = ?, data = ?
		WHERE token = ?`,
		session.ExpiresAt, session.IPAddress, session.UserAgent, session.Active, string(dataJSON), session.Token)

	return err
}

func (db *SQLiteDB) DeleteSession(ctx context.Context, token string) error {
	_, err := db.db.ExecContext(ctx, "DELETE FROM sessions WHERE token = ?", token)
	return err
}

func (db *SQLiteDB) CreateOrganization(ctx context.Context, org *models.Organization) error {
	metadataJSON, _ := json.Marshal(org.Metadata)

	_, err := db.db.ExecContext(ctx, `
		INSERT INTO organizations (id, name, slug, logo, created_at, updated_at, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		org.ID, org.Name, org.Slug, org.Logo, org.CreatedAt, org.UpdatedAt, string(metadataJSON))

	return err
}

func (db *SQLiteDB) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	org := &models.Organization{}
	var metadataJSON string

	err := db.db.QueryRowContext(ctx, `
		SELECT id, name, slug, logo, created_at, updated_at, metadata
		FROM organizations WHERE id = ?`, id).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Logo, &org.CreatedAt, &org.UpdatedAt, &metadataJSON)

	if err != nil {
		return nil, err
	}

	if metadataJSON != "" {
		json.Unmarshal([]byte(metadataJSON), &org.Metadata)
	}

	return org, nil
}