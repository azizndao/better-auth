package database

import (
	"context"
	"fmt"
	"sync"

	"better-auth/internal/models"
)

// InMemoryDB is an in-memory database implementation
type InMemoryDB struct {
	users         map[string]*models.User
	sessions      map[string]*models.Session
	organizations map[string]*models.Organization
	oauthAccounts map[string]*models.OAuthAccount
	mutex         sync.RWMutex
}

func (db *InMemoryDB) CreateUser(ctx context.Context, user *models.User) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, exists := db.users[user.ID]; exists {
		return fmt.Errorf("user already exists")
	}

	for _, u := range db.users {
		if u.Email == user.Email {
			return fmt.Errorf("email already exists")
		}
	}

	db.users[user.ID] = user
	return nil
}

func (db *InMemoryDB) GetUser(ctx context.Context, id string) (*models.User, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	user, exists := db.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (db *InMemoryDB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	for _, user := range db.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func (db *InMemoryDB) UpdateUser(ctx context.Context, user *models.User) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, exists := db.users[user.ID]; !exists {
		return fmt.Errorf("user not found")
	}

	db.users[user.ID] = user
	return nil
}

func (db *InMemoryDB) DeleteUser(ctx context.Context, id string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	delete(db.users, id)
	return nil
}

func (db *InMemoryDB) CreateSession(ctx context.Context, session *models.Session) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.sessions[session.Token] = session
	return nil
}

func (db *InMemoryDB) GetSession(ctx context.Context, token string) (*models.Session, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	session, exists := db.sessions[token]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}

func (db *InMemoryDB) UpdateSession(ctx context.Context, session *models.Session) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, exists := db.sessions[session.Token]; !exists {
		return fmt.Errorf("session not found")
	}

	db.sessions[session.Token] = session
	return nil
}

func (db *InMemoryDB) DeleteSession(ctx context.Context, token string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	delete(db.sessions, token)
	return nil
}

func (db *InMemoryDB) CreateOrganization(ctx context.Context, org *models.Organization) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.organizations[org.ID] = org
	return nil
}

func (db *InMemoryDB) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	org, exists := db.organizations[id]
	if !exists {
		return nil, fmt.Errorf("organization not found")
	}

	return org, nil
}