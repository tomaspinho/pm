package auth

import (
	"fmt"
	"regexp"
	"strings"
)

// emailRegex is a simple but effective email format validator.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks the email format and returns a normalized (lowercased, trimmed) version.
func ValidateEmail(email string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(email))
	if normalized == "" {
		return "", fmt.Errorf("email is required")
	}
	if !emailRegex.MatchString(normalized) {
		return "", fmt.Errorf("invalid email format")
	}
	return normalized, nil
}
