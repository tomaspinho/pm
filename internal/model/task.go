package model

import (
	"encoding/json"
	"time"
)

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

	Labels []Label `db:"-"          json:"labels,omitempty"`
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

// TaskWithColumn extends Task with column information for display purposes
type TaskWithColumn struct {
	Task
	ColumnName  string `db:"column_name" json:"column_name"`
	ColumnColor string `db:"column_color" json:"column_color"`
}

// TaskSearchResult represents a task in search results with full context
type TaskSearchResult struct {
	ID          int64      `db:"id" json:"id"`
	Title       string     `db:"title" json:"title"`
	Description string     `db:"description" json:"description"`
	ProjectID   int64      `db:"project_id" json:"project_id"`
	ProjectName string     `db:"project_name" json:"project_name"`
	ColumnID    int64      `db:"column_id" json:"column_id"`
	ColumnName  string     `db:"column_name" json:"column_name"`
	ColumnColor string     `db:"column_color" json:"column_color"`
	Author      string     `db:"author" json:"author"`
	DueDate     *time.Time `db:"due_date" json:"due_date,omitempty"`
	Labels      []Label    `db:"-" json:"labels,omitempty"`
}

// TaskActivityRecord represents a single activity record with user info.
type TaskActivityRecord struct {
	ID              int64     `db:"id"              json:"id"`
	TaskID          int64     `db:"task_id"         json:"task_id"`
	UserID          int64     `db:"user_id"         json:"user_id"`
	Action          string    `db:"action"          json:"action"`
	FieldName       *string   `db:"field_name"      json:"field_name,omitempty"`
	OldValueRaw     []byte    `db:"old_value"       json:"-"`
	NewValueRaw     []byte    `db:"new_value"       json:"-"`
	CreatedAt       time.Time `db:"created_at"      json:"created_at"`
	UserDisplayName string    `db:"user_display_name" json:"user_display_name"`
	UserEmail       string    `db:"user_email"      json:"user_email"`
}

// OldValue returns the parsed old value as interface{}.
func (r *TaskActivityRecord) OldValue() any {
	if len(r.OldValueRaw) == 0 {
		return nil
	}
	var val any
	if err := json.Unmarshal(r.OldValueRaw, &val); err != nil {
		return string(r.OldValueRaw)
	}
	return val
}

// NewValue returns the parsed new value as interface{}.
func (r *TaskActivityRecord) NewValue() any {
	if len(r.NewValueRaw) == 0 {
		return nil
	}
	var val any
	if err := json.Unmarshal(r.NewValueRaw, &val); err != nil {
		return string(r.NewValueRaw)
	}
	return val
}

// HasValues returns true if the activity has both old and new values.
func (r *TaskActivityRecord) HasValues() bool {
	return len(r.OldValueRaw) > 0 && len(r.NewValueRaw) > 0
}
