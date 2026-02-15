package store

import (
	"context"
	"fmt"
	"time"

	"pm/internal/model"
)

// CreateSession inserts a new session.
func (s *Store) CreateSession(ctx context.Context, id string, userID int64, expiresAt time.Time) (*model.Session, error) {
	var session model.Session
	err := s.db.GetContext(ctx, &session,
		`INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3) RETURNING *`,
		id, userID, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}
	return &session, nil
}

// GetSession fetches a session by ID (only if not expired).
func (s *Store) GetSession(ctx context.Context, id string) (*model.Session, error) {
	var session model.Session
	err := s.db.GetContext(ctx, &session,
		`SELECT * FROM sessions WHERE id = $1 AND expires_at > $2`, id, time.Now())
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}
	return &session, nil
}

// DeleteSession removes a session by ID.
func (s *Store) DeleteSession(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}

// DeleteUserSessions removes all sessions for a user.
func (s *Store) DeleteUserSessions(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("deleting user sessions: %w", err)
	}
	return nil
}

// DeleteExpiredSessions removes all expired sessions. Returns the number deleted.
func (s *Store) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < $1`, time.Now())
	if err != nil {
		return 0, fmt.Errorf("deleting expired sessions: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("getting rows affected: %w", err)
	}
	return count, nil
}
