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
	BlockedBy     []Task `db:"-" json:"blocked_by"`     // Tasks this depends on
	Blocking      []Task `db:"-" json:"blocking"`       // Tasks that depend on this
	HasCycle      bool   `db:"-" json:"has_cycle"`      // Whether a circular dependency exists
	CyclePath     []int64 `db:"-" json:"cycle_path"`    // The path of the cycle if it exists
}
