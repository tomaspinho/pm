package model

import "time"

// Organization represents a group of users who share projects.
type Organization struct {
	ID          int64      `db:"id"            json:"id"`
	Name        string     `db:"name"          json:"name"`
	OwnerUserID int64      `db:"owner_user_id" json:"owner_user_id"`
	CreatedAt   time.Time  `db:"created_at"    json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"    json:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"    json:"deleted_at,omitempty"`
}

// OrganizationMember represents a user's membership in an organization.
type OrganizationMember struct {
	OrganizationID int64      `db:"organization_id" json:"organization_id"`
	UserID         int64      `db:"user_id"         json:"user_id"`
	JoinedAt       time.Time  `db:"joined_at"       json:"joined_at"`
	CreatedAt      time.Time  `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"      json:"updated_at"`
	DeletedAt      *time.Time `db:"deleted_at"       json:"deleted_at,omitempty"`
}
