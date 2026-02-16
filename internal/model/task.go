package model

import "time"

// Task represents a kanban card belonging to a project.
type Task struct {
	ID          int64        `db:"id"          json:"id"`
	ProjectID   int64        `db:"project_id"  json:"project_id"`
	Title       string       `db:"title"       json:"title"`
	Description string       `db:"description" json:"description"`
	Author      string       `db:"author"      json:"author"`
	ColumnID    int64        `db:"column_id"   json:"column_id"`
	Position    int          `db:"position"    json:"position"`
	Metadata    TaskMetadata `db:"metadata"    json:"metadata"`
	DueDate     *time.Time   `db:"due_date"    json:"due_date,omitempty"`
	CreatedAt   time.Time    `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"  json:"updated_at"`
	DeletedAt   *time.Time   `db:"deleted_at"  json:"deleted_at,omitempty"`
}

// IsDeleted checks if the task is soft-deleted.
func (t *Task) IsDeleted() bool {
	return t.DeletedAt != nil
}

// IsOverdue returns true if the task is past its due date
func (t *Task) IsOverdue() bool {
	if t.DueDate == nil {
		return false
	}
	now := time.Now().Truncate(24 * time.Hour)
	due := t.DueDate.Truncate(24 * time.Hour)
	return due.Before(now)
}

// IsDueSoon returns true if the task is due within 3 days
func (t *Task) IsDueSoon() bool {
	if t.DueDate == nil || t.IsOverdue() {
		return false
	}
	now := time.Now().Truncate(24 * time.Hour)
	due := t.DueDate.Truncate(24 * time.Hour)
	threeDays := now.Add(72 * time.Hour)
	return !due.After(threeDays)
}

// DueDateString returns the due date in YYYY-MM-DD format, or empty string
func (t *Task) DueDateString() string {
	if t.DueDate == nil {
		return ""
	}
	return t.DueDate.Format("2006-01-02")
}

