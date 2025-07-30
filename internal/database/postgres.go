package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"better-auth/internal/models"
	_ "github.com/lib/pq"
)

// PostgresDB implements Database interface for PostgreSQL
type PostgresDB struct {
	connectionURL string
	db            *sql.DB
}

func (db *PostgresDB) migrate() error {
	var err error
	db.db, err = sql.Open("postgres", db.connectionURL)
	if err != nil {
		return err
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			email_verified BOOLEAN DEFAULT FALSE,
			name VARCHAR(255),
			image VARCHAR(512),
			password VARCHAR(255),
			two_factor_enabled BOOLEAN DEFAULT FALSE,
			two_factor_secret VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			metadata JSONB,
			last_sign_in TIMESTAMP,
			sign_in_count INTEGER DEFAULT 0,
			blocked BOOLEAN DEFAULT FALSE,
			email_verify_token VARCHAR(255),
			password_reset_token VARCHAR(255)
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			ip_address VARCHAR(45),
			user_agent TEXT,
			active BOOLEAN DEFAULT TRUE,
			data JSONB
		)`,
		`CREATE TABLE IF NOT EXISTS organizations (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) UNIQUE NOT NULL,
			logo VARCHAR(512),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			metadata JSONB
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
	}

	for _, query := range queries {
		if _, err := db.db.Exec(query); err != nil {
			return fmt.Errorf("migration failed: %v", err)
		}
	}
	return nil
}

func (db *PostgresDB) CreateUser(ctx context.Context, user *models.User) error {
	metadataJSON, _ := json.Marshal(user.Metadata)

	_, err := db.db.ExecContext(ctx, `
		INSERT INTO users (id, email, email_verified, name, image, password, two_factor_enabled, 
			two_factor_secret, created_at, updated_at, metadata, last_sign_in, sign_in_count, 
			blocked, email_verify_token, password_reset_token)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
		user.ID, user.Email, user.EmailVerified, user.Name, user.Image, user.Password,
		user.TwoFactorEnabled, user.TwoFactorSecret, user.CreatedAt, user.UpdatedAt,
		metadataJSON, user.LastSignIn, user.SignInCount, user.Blocked,
		user.EmailVerifyToken, user.PasswordResetToken)

	return err
}

func (db *PostgresDB) GetUser(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	var metadataJSON []byte

	err := db.db.QueryRowContext(ctx, `
		SELECT id, email, email_verified, name, image, password, two_factor_enabled,
			two_factor_secret, created_at, updated_at, metadata, last_sign_in, 
			sign_in_count, blocked, email_verify_token, password_reset_token
		FROM users WHERE id = $1`, id).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.Name, &user.Image,
		&user.Password, &user.TwoFactorEnabled, &user.TwoFactorSecret,
		&user.CreatedAt, &user.UpdatedAt, &metadataJSON, &user.LastSignIn,
		&user.SignInCount, &user.Blocked, &user.EmailVerifyToken, &user.PasswordResetToken)

	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &user.Metadata)
	}

	return user, nil
}

func (db *PostgresDB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	var metadataJSON []byte

	err := db.db.QueryRowContext(ctx, `
		SELECT id, email, email_verified, name, image, password, two_factor_enabled,
			two_factor_secret, created_at, updated_at, metadata, last_sign_in, 
			sign_in_count, blocked, email_verify_token, password_reset_token
		FROM users WHERE email = $1`, email).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.Name, &user.Image,
		&user.Password, &user.TwoFactorEnabled, &user.TwoFactorSecret,
		&user.CreatedAt, &user.UpdatedAt, &metadataJSON, &user.LastSignIn,
		&user.SignInCount, &user.Blocked, &user.EmailVerifyToken, &user.PasswordResetToken)

	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &user.Metadata)
	}

	return user, nil
}

func (db *PostgresDB) UpdateUser(ctx context.Context, user *models.User) error {
	metadataJSON, _ := json.Marshal(user.Metadata)

	_, err := db.db.ExecContext(ctx, `
		UPDATE users SET email = $2, email_verified = $3, name = $4, image = $5, 
			password = $6, two_factor_enabled = $7, two_factor_secret = $8, 
			updated_at = $9, metadata = $10, last_sign_in = $11, sign_in_count = $12,
			blocked = $13, email_verify_token = $14, password_reset_token = $15
		WHERE id = $1`,
		user.ID, user.Email, user.EmailVerified, user.Name, user.Image, user.Password,
		user.TwoFactorEnabled, user.TwoFactorSecret, user.UpdatedAt, metadataJSON,
		user.LastSignIn, user.SignInCount, user.Blocked, user.EmailVerifyToken,
		user.PasswordResetToken)

	return err
}

func (db *PostgresDB) DeleteUser(ctx context.Context, id string) error {
	_, err := db.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}

func (db *PostgresDB) CreateSession(ctx context.Context, session *models.Session) error {
	dataJSON, _ := json.Marshal(session.Data)

	_, err := db.db.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, token, expires_at, created_at, ip_address, user_agent, active, data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		session.ID, session.UserID, session.Token, session.ExpiresAt, session.CreatedAt,
		session.IPAddress, session.UserAgent, session.Active, dataJSON)

	return err
}

func (db *PostgresDB) GetSession(ctx context.Context, token string) (*models.Session, error) {
	session := &models.Session{}
	var dataJSON []byte

	err := db.db.QueryRowContext(ctx, `
		SELECT id, user_id, token, expires_at, created_at, ip_address, user_agent, active, data
		FROM sessions WHERE token = $1`, token).Scan(
		&session.ID, &session.UserID, &session.Token, &session.ExpiresAt,
		&session.CreatedAt, &session.IPAddress, &session.UserAgent,
		&session.Active, &dataJSON)

	if err != nil {
		return nil, err
	}

	if len(dataJSON) > 0 {
		json.Unmarshal(dataJSON, &session.Data)
	}

	return session, nil
}

func (db *PostgresDB) UpdateSession(ctx context.Context, session *models.Session) error {
	dataJSON, _ := json.Marshal(session.Data)

	_, err := db.db.ExecContext(ctx, `
		UPDATE sessions SET expires_at = $2, ip_address = $3, user_agent = $4, active = $5, data = $6
		WHERE token = $1`,
		session.Token, session.ExpiresAt, session.IPAddress, session.UserAgent, session.Active, dataJSON)

	return err
}

func (db *PostgresDB) DeleteSession(ctx context.Context, token string) error {
	_, err := db.db.ExecContext(ctx, "DELETE FROM sessions WHERE token = $1", token)
	return err
}

func (db *PostgresDB) CreateOrganization(ctx context.Context, org *models.Organization) error {
	metadataJSON, _ := json.Marshal(org.Metadata)

	_, err := db.db.ExecContext(ctx, `
		INSERT INTO organizations (id, name, slug, logo, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		org.ID, org.Name, org.Slug, org.Logo, org.CreatedAt, org.UpdatedAt, metadataJSON)

	return err
}

func (db *PostgresDB) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	org := &models.Organization{}
	var metadataJSON []byte

	err := db.db.QueryRowContext(ctx, `
		SELECT id, name, slug, logo, created_at, updated_at, metadata
		FROM organizations WHERE id = $1`, id).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Logo, &org.CreatedAt, &org.UpdatedAt, &metadataJSON)

	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &org.Metadata)
	}

	return org, nil
}