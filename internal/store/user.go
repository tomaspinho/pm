package store

import (
	"context"
	"fmt"
	"time"

	"pm/internal/model"
)

// CreateUser inserts a new user with email, password hash, and display name.
func (s *Store) CreateUser(ctx context.Context, email, passwordHash, displayName string) (*model.User, error) {
	var user model.User
	err := s.db.GetContext(ctx, &user,
		`INSERT INTO users (email, password_hash, display_name) VALUES ($1, $2, $3) RETURNING *`,
		email, passwordHash, displayName)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return &user, nil
}

// GetUserByEmail fetches a user by email address.
func (s *Store) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := s.db.GetContext(ctx, &user,
		`SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`, email)
	if err != nil {
		return nil, fmt.Errorf("getting user by email: %w", err)
	}
	return &user, nil
}

// GetUserByID fetches a user by ID.
func (s *Store) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := s.db.GetContext(ctx, &user,
		`SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return nil, fmt.Errorf("getting user %d: %w", id, err)
	}
	return &user, nil
}

// GetUser fetches a user by ID (convenience alias for GetUserByID).
func (s *Store) GetUser(ctx context.Context, id int64) (*model.User, error) {
	return s.GetUserByID(ctx, id)
}

// UpdateLastViewedProject updates the user's last viewed project.
func (s *Store) UpdateLastViewedProject(ctx context.Context, userID, projectID int64) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET last_viewed_project_id = $1, last_viewed_at = $2, updated_at = NOW() WHERE id = $3`,
		projectID, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("updating last viewed project: %w", err)
	}
	return nil
}

// UpdateUserDisplayName updates a user's display name.
func (s *Store) UpdateUserDisplayName(ctx context.Context, userID int64, displayName string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET display_name = $1, updated_at = NOW() WHERE id = $2`,
		displayName, userID)
	if err != nil {
		return fmt.Errorf("updating user display name: %w", err)
	}
	return nil
}
