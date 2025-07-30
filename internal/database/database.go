package database

import (
	"context"

	"better-auth/internal/models"
)

// Database interface defines database operations
type Database interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
	CreateSession(ctx context.Context, session *models.Session) error
	GetSession(ctx context.Context, token string) (*models.Session, error)
	UpdateSession(ctx context.Context, session *models.Session) error
	DeleteSession(ctx context.Context, token string) error
	CreateOrganization(ctx context.Context, org *models.Organization) error
	GetOrganization(ctx context.Context, id string) (*models.Organization, error)
}

// Factory functions
func NewPostgresDB(connectionURL string) Database {
	db := &PostgresDB{connectionURL: connectionURL}
	db.migrate()
	return db
}

func NewMySQLDB(connectionURL string) Database {
	db := &MySQLDB{connectionURL: connectionURL}
	db.migrate()
	return db
}

func NewSQLiteDB(connectionURL string) Database {
	db := &SQLiteDB{connectionURL: connectionURL}
	db.migrate()
	return db
}

func NewInMemoryDB() Database {
	return &InMemoryDB{
		users:         make(map[string]*models.User),
		sessions:      make(map[string]*models.Session),
		organizations: make(map[string]*models.Organization),
		oauthAccounts: make(map[string]*models.OAuthAccount),
	}
}