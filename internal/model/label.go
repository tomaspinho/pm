package model

import "time"

type Label struct {
	ID        int64      `db:"id"         json:"id"`
	ProjectID int64      `db:"project_id" json:"project_id"`
	Name      string     `db:"name"       json:"name"`
	Color     string     `db:"color"      json:"color"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

func (l *Label) IsDeleted() bool {
	return l.DeletedAt != nil
}

var LabelColors = []string{
	"#6B7280", // Gray
	"#EF4444", // Red
	"#F59E0B", // Amber
	"#10B981", // Green
	"#3B82F6", // Blue
	"#8B5CF6", // Purple
	"#EC4899", // Pink
	"#14B8A6", // Teal
}
