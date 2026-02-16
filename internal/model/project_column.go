package model

import "time"

// ProjectColumn represents a custom workflow column/stage in a project's kanban board.
type ProjectColumn struct {
	ID        int64      `db:"id"         json:"id"`
	ProjectID int64      `db:"project_id" json:"project_id"`
	Name      string     `db:"name"       json:"name"`
	Color     string     `db:"color"      json:"color"`
	Position  int        `db:"position"   json:"position"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// IsDeleted checks if the column is soft-deleted.
func (pc *ProjectColumn) IsDeleted() bool {
	return pc.DeletedAt != nil
}

// ColumnColors is the preset color palette for columns.
var ColumnColors = []string{
	"#6B7280", // Gray
	"#EF4444", // Red
	"#F59E0B", // Amber
	"#10B981", // Green
	"#3B82F6", // Blue
	"#8B5CF6", // Purple
	"#EC4899", // Pink
	"#14B8A6", // Teal
}

// DefaultColumns returns the default set of columns for new projects.
func DefaultColumns() []ProjectColumn {
	return []ProjectColumn{
		{Name: "To Do", Color: "#6B7280", Position: 0},
		{Name: "In Progress", Color: "#3B82F6", Position: 1},
		{Name: "Done", Color: "#10B981", Position: 2},
	}
}
