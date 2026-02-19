package store

import (
	"context"
	"fmt"

	"pm/internal/model"
)

// CreateComment creates a new comment (or reply if parentID is provided).
func (s *Store) CreateComment(ctx context.Context, taskID, userID int64, parentID *int64, content string) (*model.TaskComment, error) {
	if len(content) > 10000 {
		return nil, fmt.Errorf("comment content exceeds maximum length of 10000 characters")
	}

	var comment model.TaskComment
	err := s.db.GetContext(ctx, &comment, `
		INSERT INTO task_comments (task_id, user_id, parent_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`, taskID, userID, parentID, content)
	if err != nil {
		return nil, fmt.Errorf("creating comment: %w", err)
	}
	return &comment, nil
}

// GetComment retrieves a comment by ID (excluding soft-deleted).
func (s *Store) GetComment(ctx context.Context, commentID int64) (*model.TaskComment, error) {
	var comment model.TaskComment
	err := s.db.GetContext(ctx, &comment,
		`SELECT * FROM task_comments WHERE id = $1 AND deleted_at IS NULL`, commentID)
	if err != nil {
		return nil, fmt.Errorf("getting comment %d: %w", commentID, err)
	}
	return &comment, nil
}

// GetCommentWithAuthor retrieves a comment with author information.
func (s *Store) GetCommentWithAuthor(ctx context.Context, commentID int64) (*model.CommentWithAuthor, error) {
	var comment model.CommentWithAuthor
	err := s.db.GetContext(ctx, &comment, `
		SELECT c.*, u.email as author_email
		FROM task_comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = $1 AND c.deleted_at IS NULL
	`, commentID)
	if err != nil {
		return nil, fmt.Errorf("getting comment with author %d: %w", commentID, err)
	}
	return &comment, nil
}

// ListTaskComments returns all comments for a task (including soft-deleted for display).
func (s *Store) ListTaskComments(ctx context.Context, taskID int64) ([]model.CommentWithAuthor, error) {
	var comments []model.CommentWithAuthor
	err := s.db.SelectContext(ctx, &comments, `
		SELECT c.*, u.email as author_email
		FROM task_comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.task_id = $1
		ORDER BY c.created_at ASC
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("listing comments for task %d: %w", taskID, err)
	}
	return comments, nil
}

// BuildCommentTree converts a flat list of comments into a nested tree structure.
// Enforces max depth. Comments beyond maxDepth are flattened at maxDepth.
func BuildCommentTree(comments []model.CommentWithAuthor, maxDepth int) []model.CommentThread {
	if len(comments) == 0 {
		return []model.CommentThread{}
	}

	// Build maps for quick lookups
	commentMap := make(map[int64]*model.CommentThread)
	childrenMap := make(map[int64][]int64) // parent_id -> []child_ids
	var topLevelIDs []int64                // IDs of top-level comments (parent_id is NULL)

	// First pass: create all comment threads
	for _, c := range comments {
		thread := model.CommentThread{
			CommentWithAuthor: c,
			Replies:           []model.CommentThread{},
			Depth:             0, // Will be set during tree building
		}
		commentMap[c.ID] = &thread

		// Build children map
		if c.ParentID != nil {
			childrenMap[*c.ParentID] = append(childrenMap[*c.ParentID], c.ID)
		} else {
			// Track top-level comments (parent_id is NULL)
			topLevelIDs = append(topLevelIDs, c.ID)
		}
	}

	// Second pass: build tree recursively
	var buildTree func(parentID int64, depth int) []model.CommentThread
	buildTree = func(parentID int64, depth int) []model.CommentThread {
		childIDs := childrenMap[parentID]
		if len(childIDs) == 0 {
			return []model.CommentThread{}
		}

		result := make([]model.CommentThread, 0, len(childIDs))
		for _, childID := range childIDs {
			thread := commentMap[childID]
			thread.Depth = depth

			// If we're at max depth, flatten any deeper children at this level
			if depth < maxDepth {
				thread.Replies = buildTree(childID, depth+1)
			} else {
				// At max depth, collect all descendants and flatten them here
				thread.Replies = flattenDescendants(childID, childrenMap, commentMap, depth)
			}

			result = append(result, *thread)
		}
		return result
	}

	// Build tree starting from top-level comments
	result := make([]model.CommentThread, 0, len(topLevelIDs))
	for _, topLevelID := range topLevelIDs {
		thread := commentMap[topLevelID]
		thread.Depth = 0
		thread.Replies = buildTree(topLevelID, 1)
		result = append(result, *thread)
	}

	return result
}

// flattenDescendants collects all descendants of a comment and returns them at the same depth.
func flattenDescendants(parentID int64, childrenMap map[int64][]int64, commentMap map[int64]*model.CommentThread, depth int) []model.CommentThread {
	childIDs := childrenMap[parentID]
	if len(childIDs) == 0 {
		return []model.CommentThread{}
	}

	result := make([]model.CommentThread, 0)
	for _, childID := range childIDs {
		thread := commentMap[childID]
		thread.Depth = depth
		thread.Replies = []model.CommentThread{} // Flatten - no more nesting
		result = append(result, *thread)

		// Recursively collect descendants
		descendants := flattenDescendants(childID, childrenMap, commentMap, depth)
		result = append(result, descendants...)
	}
	return result
}

// UpdateComment updates a comment's content and sets edited_at timestamp.
func (s *Store) UpdateComment(ctx context.Context, commentID int64, content string) error {
	if len(content) > 10000 {
		return fmt.Errorf("comment content exceeds maximum length of 10000 characters")
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE task_comments
		SET content = $1, edited_at = NOW(), updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`, content, commentID)
	if err != nil {
		return fmt.Errorf("updating comment %d: %w", commentID, err)
	}
	return nil
}

// DeleteComment soft-deletes a comment.
func (s *Store) DeleteComment(ctx context.Context, commentID int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE task_comments
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, commentID)
	if err != nil {
		return fmt.Errorf("deleting comment %d: %w", commentID, err)
	}
	return nil
}

// CountTaskComments returns the number of non-deleted comments on a task.
func (s *Store) CountTaskComments(ctx context.Context, taskID int64) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM task_comments WHERE task_id = $1 AND deleted_at IS NULL`, taskID)
	if err != nil {
		return 0, fmt.Errorf("counting comments for task %d: %w", taskID, err)
	}
	return count, nil
}

// HasReplies returns true if a comment has any child comments (deleted or not).
func (s *Store) HasReplies(ctx context.Context, commentID int64) (bool, error) {
	var count int
	err := s.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM task_comments WHERE parent_id = $1`, commentID)
	if err != nil {
		return false, fmt.Errorf("checking replies for comment %d: %w", commentID, err)
	}
	return count > 0, nil
}

// GetCommentDepth calculates the depth of a comment in the tree.
func (s *Store) GetCommentDepth(ctx context.Context, commentID int64) (int, error) {
	var depth int
	err := s.db.GetContext(ctx, &depth, `
		WITH RECURSIVE comment_path AS (
			SELECT id, parent_id, 0 AS depth
			FROM task_comments
			WHERE id = $1

			UNION ALL

			SELECT c.id, c.parent_id, cp.depth + 1
			FROM task_comments c
			JOIN comment_path cp ON c.id = cp.parent_id
		)
		SELECT COALESCE(MAX(depth), 0) FROM comment_path
	`, commentID)
	if err != nil {
		return 0, fmt.Errorf("getting comment depth for %d: %w", commentID, err)
	}
	return depth, nil
}
