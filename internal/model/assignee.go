package model

import "time"

// TaskAssignee represents the assignment of a user to a task
type TaskAssignee struct {
	TaskID     int64     `db:"task_id"`
	UserID     int64     `db:"user_id"`
	AssignedAt time.Time `db:"assigned_at"`
}

// AssigneeInfo contains user information for display in assignee lists
type AssigneeInfo struct {
	UserID      int64  `db:"user_id" json:"user_id"`
	Email       string `db:"email" json:"email"`
	DisplayName string `db:"display_name" json:"display_name"`
}

// Initials returns the first two characters of the display name, uppercased
func (a *AssigneeInfo) Initials() string {
	if len(a.DisplayName) == 0 {
		return "?"
	}
	runes := []rune(a.DisplayName)
	if len(runes) == 1 {
		return string([]rune{runes[0]})
	}
	return string([]rune{runes[0], runes[1]})
}

// AvatarColor returns a consistent hex color based on the user's email
func (a *AssigneeInfo) AvatarColor() string {
	// Use the same User logic
	u := &User{Email: a.Email, DisplayName: a.DisplayName}
	return u.AvatarColor()
}
