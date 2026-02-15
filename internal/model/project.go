package model

import "time"

// Project represents a project in the system.
type Project struct {
	ID             int64      `db:"id"              json:"id"`
	Name           string     `db:"name"            json:"name"`
	Description    string     `db:"description"     json:"description"`
	OrganizationID int64      `db:"organization_id" json:"organization_id"`
	CreatedAt      time.Time  `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"      json:"updated_at"`
	DeletedAt      *time.Time `db:"deleted_at"       json:"deleted_at,omitempty"`
}
