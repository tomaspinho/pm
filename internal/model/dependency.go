package model

import "time"

// TaskDependency represents a dependency relationship between two tasks.
type TaskDependency struct {
	TaskID      int64     `db:"task_id"       json:"task_id"`
	DependsOnID int64     `db:"depends_on_id" json:"depends_on_id"`
	CreatedAt   time.Time `db:"created_at"    json:"created_at"`
}

// TaskWithDependencies includes the full dependency information.
type TaskWithDependencies struct {
	Task
	BlockedBy      []TaskWithColumn `db:"-" json:"blocked_by"`    // Tasks this depends on
	Blocking       []TaskWithColumn `db:"-" json:"blocking"`      // Tasks that depend on this
	HasCycle       bool             `db:"-" json:"has_cycle"`     // Whether a circular dependency exists
	CyclePath      []int64          `db:"-" json:"cycle_path"`    // The path of the cycle if it exists
	CommentCount   int              `db:"-" json:"comment_count"` // Number of non-deleted comments
	Assignees      []AssigneeInfo   `db:"-" json:"assignees"`     // Users assigned to this task
	CreatedByName  string           `db:"created_by_name" json:"created_by_name"`
	CreatedByEmail string           `db:"created_by_email" json:"created_by_email"`
	ColumnName     string           `db:"column_name" json:"column_name"`   // Name of the current column
	ColumnColor    string           `db:"column_color" json:"column_color"` // Color of the current column
}

// CreatorInitials returns the first two characters of the creator's display name.
func (t *TaskWithDependencies) CreatorInitials() string {
	if t.CreatedByName == "" {
		return "?"
	}
	runes := []rune(t.CreatedByName)
	if len(runes) == 1 {
		return string([]rune{runes[0]})
	}
	return string([]rune{runes[0], runes[1]})
}

// CreatorAvatarColor returns a consistent hex color based on the creator's email.
func (t *TaskWithDependencies) CreatorAvatarColor() string {
	if t.CreatedByEmail == "" {
		return "#6B7280"
	}
	u := &User{Email: t.CreatedByEmail, DisplayName: t.CreatedByName}
	return u.AvatarColor()
}
