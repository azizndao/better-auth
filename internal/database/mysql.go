package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"better-auth/internal/models"
	_ "github.com/go-sql-driver/mysql"
)

// MySQLDB implements Database interface for MySQL
type MySQLDB struct {
	connectionURL string
	db            *sql.DB
}

func (db *MySQLDB) migrate() error {
	var err error
	db.db, err = sql.Open("mysql", db.connectionURL)
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
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			metadata JSON,
			last_sign_in TIMESTAMP NULL,
			sign_in_count INTEGER DEFAULT 0,
			blocked BOOLEAN DEFAULT FALSE,
			email_verify_token VARCHAR(255),
			password_reset_token VARCHAR(255)
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			ip_address VARCHAR(45),
			user_agent TEXT,
			active BOOLEAN DEFAULT TRUE,
			data JSON,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS organizations (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) UNIQUE NOT NULL,
			logo VARCHAR(512),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			metadata JSON
		)`,
	}

	for _, query := range queries {
		if _, err := db.db.Exec(query); err != nil {
			return fmt.Errorf("migration failed: %v", err)
		}
	}
	return nil
}

func (db *MySQLDB) CreateUser(ctx context.Context, user *models.User) error {
	metadataJSON, _ := json.Marshal(user.Metadata)

	_, err := db.db.ExecContext(ctx, `
		INSERT INTO users (id, email, email_verified, name, image, password, two_factor_enabled, 
			two_factor_secret, created_at, updated_at, metadata, last_sign_in, sign_in_count, 
			blocked, email_verify_token, password_reset_token)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.Email, user.EmailVerified, user.Name, user.Image, user.Password,
		user.TwoFactorEnabled, user.TwoFactorSecret, user.CreatedAt, user.UpdatedAt,
		metadataJSON, user.LastSignIn, user.SignInCount, user.Blocked,
		user.EmailVerifyToken, user.PasswordResetToken)

	return err
}

func (db *MySQLDB) GetUser(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	var metadataJSON []byte

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

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &user.Metadata)
	}

	return user, nil
}

func (db *MySQLDB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	var metadataJSON []byte

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

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &user.Metadata)
	}

	return user, nil
}

func (db *MySQLDB) UpdateUser(ctx context.Context, user *models.User) error {
	metadataJSON, _ := json.Marshal(user.Metadata)

	_, err := db.db.ExecContext(ctx, `
		UPDATE users SET email = ?, email_verified = ?, name = ?, image = ?, 
			password = ?, two_factor_enabled = ?, two_factor_secret = ?, 
			updated_at = ?, metadata = ?, last_sign_in = ?, sign_in_count = ?,
			blocked = ?, email_verify_token = ?, password_reset_token = ?
		WHERE id = ?`,
		user.Email, user.EmailVerified, user.Name, user.Image, user.Password,
		user.TwoFactorEnabled, user.TwoFactorSecret, user.UpdatedAt, metadataJSON,
		user.LastSignIn, user.SignInCount, user.Blocked, user.EmailVerifyToken,
		user.PasswordResetToken, user.ID)

	return err
}

func (db *MySQLDB) DeleteUser(ctx context.Context, id string) error {
	_, err := db.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	return err
}

func (db *MySQLDB) CreateSession(ctx context.Context, session *models.Session) error {
	dataJSON, _ := json.Marshal(session.Data)

	_, err := db.db.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, token, expires_at, created_at, ip_address, user_agent, active, data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		session.ID, session.UserID, session.Token, session.ExpiresAt, session.CreatedAt,
		session.IPAddress, session.UserAgent, session.Active, dataJSON)

	return err
}

func (db *MySQLDB) GetSession(ctx context.Context, token string) (*models.Session, error) {
	session := &models.Session{}
	var dataJSON []byte

	err := db.db.QueryRowContext(ctx, `
		SELECT id, user_id, token, expires_at, created_at, ip_address, user_agent, active, data
		FROM sessions WHERE token = ?`, token).Scan(
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

func (db *MySQLDB) UpdateSession(ctx context.Context, session *models.Session) error {
	dataJSON, _ := json.Marshal(session.Data)

	_, err := db.db.ExecContext(ctx, `
		UPDATE sessions SET expires_at = ?, ip_address = ?, user_agent = ?, active = ?, data = ?
		WHERE token = ?`,
		session.ExpiresAt, session.IPAddress, session.UserAgent, session.Active, dataJSON, session.Token)

	return err
}

func (db *MySQLDB) DeleteSession(ctx context.Context, token string) error {
	_, err := db.db.ExecContext(ctx, "DELETE FROM sessions WHERE token = ?", token)
	return err
}

func (db *MySQLDB) CreateOrganization(ctx context.Context, org *models.Organization) error {
	metadataJSON, _ := json.Marshal(org.Metadata)

	_, err := db.db.ExecContext(ctx, `
		INSERT INTO organizations (id, name, slug, logo, created_at, updated_at, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		org.ID, org.Name, org.Slug, org.Logo, org.CreatedAt, org.UpdatedAt, metadataJSON)

	return err
}

func (db *MySQLDB) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	org := &models.Organization{}
	var metadataJSON []byte

	err := db.db.QueryRowContext(ctx, `
		SELECT id, name, slug, logo, created_at, updated_at, metadata
		FROM organizations WHERE id = ?`, id).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Logo, &org.CreatedAt, &org.UpdatedAt, &metadataJSON)

	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &org.Metadata)
	}

	return org, nil
}