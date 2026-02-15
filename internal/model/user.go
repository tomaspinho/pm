package model

import "time"

// User represents a registered user in the system.
type User struct {
	ID                  int64      `db:"id"                    json:"id"`
	Email               string     `db:"email"                 json:"email"`
	PasswordHash        string     `db:"password_hash"         json:"-"`
	LastViewedProjectID *int64     `db:"last_viewed_project_id" json:"last_viewed_project_id,omitempty"`
	LastViewedAt        *time.Time `db:"last_viewed_at"        json:"last_viewed_at,omitempty"`
	CreatedAt           time.Time  `db:"created_at"            json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"            json:"updated_at"`
	DeletedAt           *time.Time `db:"deleted_at"            json:"deleted_at,omitempty"`
}
