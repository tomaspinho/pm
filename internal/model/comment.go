package model

import "time"

// TaskComment represents a comment on a task (or a reply to another comment).
type TaskComment struct {
	ID        int64      `db:"id"         json:"id"`
	TaskID    int64      `db:"task_id"    json:"task_id"`
	UserID    int64      `db:"user_id"    json:"user_id"`
	ParentID  *int64     `db:"parent_id"  json:"parent_id,omitempty"`
	Content   string     `db:"content"    json:"content"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	EditedAt  *time.Time `db:"edited_at"  json:"edited_at,omitempty"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// IsEdited returns true if the comment has been edited.
func (c *TaskComment) IsEdited() bool {
	return c.EditedAt != nil
}

// IsDeleted returns true if the comment is soft-deleted.
func (c *TaskComment) IsDeleted() bool {
	return c.DeletedAt != nil
}

// CommentWithAuthor includes author information for display.
type CommentWithAuthor struct {
	TaskComment
	AuthorEmail string `db:"author_email" json:"author_email"`
}

// CommentThread represents a hierarchical comment tree.
type CommentThread struct {
	CommentWithAuthor
	Replies []CommentThread `json:"replies"`
	Depth   int             `json:"depth"` // Nesting level (0 = top-level)
}
