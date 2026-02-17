package model

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// User represents a registered user in the system.
type User struct {
	ID                  int64      `db:"id"                    json:"id"`
	Email               string     `db:"email"                 json:"email"`
	DisplayName         string     `db:"display_name"          json:"display_name"`
	PasswordHash        string     `db:"password_hash"         json:"-"`
	LastViewedProjectID *int64     `db:"last_viewed_project_id" json:"last_viewed_project_id,omitempty"`
	LastViewedAt        *time.Time `db:"last_viewed_at"        json:"last_viewed_at,omitempty"`
	CreatedAt           time.Time  `db:"created_at"            json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"            json:"updated_at"`
	DeletedAt           *time.Time `db:"deleted_at"            json:"deleted_at,omitempty"`
}

// Initials returns the first two characters of the display name, uppercased
// Used for avatar badges
func (u *User) Initials() string {
	if len(u.DisplayName) == 0 {
		return "?"
	}
	runes := []rune(u.DisplayName)
	if len(runes) == 1 {
		return string([]rune{runes[0]})
	}
	return string([]rune{runes[0], runes[1]})
}

// AvatarColor returns a consistent hex color based on the user's email
// Used for avatar background colors
func (u *User) AvatarColor() string {
	hash := sha256.Sum256([]byte(u.Email))
	hashStr := hex.EncodeToString(hash[:])

	// Use first 6 characters of hash as RGB color
	// Ensure colors are vibrant by forcing minimum brightness
	r := parseInt(hashStr[0:2], 16)
	g := parseInt(hashStr[2:4], 16)
	b := parseInt(hashStr[4:6], 16)

	// Ensure minimum brightness of 60 to avoid very dark colors
	if r < 60 {
		r = 60 + (r % 100)
	}
	if g < 60 {
		g = 60 + (g % 100)
	}
	if b < 60 {
		b = 60 + (b % 100)
	}

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func parseInt(s string, base int) int {
	var result int
	_, _ = fmt.Sscanf(s, "%x", &result) // Ignore error, will return 0 if invalid
	return result
}
