package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloud-drive/user-service/internal/models"
	"github.com/google/uuid"
	"sync"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrUserExists is returned when a user already exists
	ErrUserExists = errors.New("user already exists")
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
	FindByEmail(email string) (*models.User, error)
}

// InMemoryUserRepository is an in-memory implementation of UserRepository
type InMemoryUserRepository struct {
	users map[string]*models.User
	mu    sync.RWMutex
}

// NewInMemoryUserRepository creates a new in-memory user repository
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*models.User),
	}
}

// Create creates a new user
func (r *InMemoryUserRepository) Create(ctx context.Context, user *models.User) error {
	// Check if user with same username or email already exists
	for _, u := range r.users {
		if u.Username == user.Username {
			return ErrUserExists
		}
		if u.Email == user.Email {
			return ErrUserExists
		}
	}

	// Generate a new ID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Store the user
	r.users[user.ID] = user
	return nil
}

// GetByID returns a user by ID
func (r *InMemoryUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetByUsername returns a user by username
func (r *InMemoryUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

// GetByEmail returns a user by email
func (r *InMemoryUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

// Update updates a user
func (r *InMemoryUserRepository) Update(ctx context.Context, user *models.User) error {
	_, ok := r.users[user.ID]
	if !ok {
		return ErrUserNotFound
	}
	r.users[user.ID] = user
	return nil
}

// Delete deletes a user
func (r *InMemoryUserRepository) Delete(ctx context.Context, id string) error {
	_, ok := r.users[id]
	if !ok {
		return ErrUserNotFound
	}
	delete(r.users, id)
	return nil
}

// List returns a list of users
func (r *InMemoryUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	users := make([]*models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	// Apply pagination
	if offset >= len(users) {
		return []*models.User{}, nil
	}

	end := offset + limit
	if end > len(users) {
		end = len(users)
	}

	return users[offset:end], nil
}

// FindByEmail tìm người dùng theo email
func (r *InMemoryUserRepository) FindByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found with email: %s", email)
}
