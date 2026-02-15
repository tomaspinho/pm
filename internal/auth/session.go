package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	// SessionCookieName is the name of the session cookie.
	SessionCookieName = "session_id"

	// SessionDuration is the default session duration (1 day).
	SessionDuration = 24 * time.Hour

	// SessionDurationRemember is the extended session duration (30 days).
	SessionDurationRemember = 30 * 24 * time.Hour
)

// GenerateSessionID creates a cryptographically secure random session ID.
func GenerateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generating session ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
