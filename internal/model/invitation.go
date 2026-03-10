package model

import "time"

type Invitation struct {
	ID         int64      `db:"id"          json:"id"`
	OrgID      int64      `db:"org_id"      json:"org_id"`
	Email      string     `db:"email"       json:"email"`
	InvitedBy  int64      `db:"invited_by"  json:"invited_by"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
	ExpiresAt  time.Time  `db:"expires_at"  json:"expires_at"`
	AcceptedAt *time.Time `db:"accepted_at" json:"accepted_at,omitempty"`
}
