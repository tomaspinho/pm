package model

import "time"

// Session represents an authenticated user session.
type Session struct {
	ID        string    `db:"id"         json:"id"`
	UserID    int64     `db:"user_id"    json:"user_id"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// IsExpired returns true if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
