package store

import (
	"context"
	"fmt"

	"pm/internal/model"
)

// AddDependency creates a dependency relationship between two tasks.
func (s *Store) AddDependency(ctx context.Context, taskID, dependsOnID int64) error {
	query := `
		INSERT INTO task_dependencies (task_id, depends_on_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err := s.db.ExecContext(ctx, query, taskID, dependsOnID)
	if err != nil {
		return fmt.Errorf("adding dependency: %w", err)
	}
	return nil
}

// RemoveDependency removes a dependency relationship.
func (s *Store) RemoveDependency(ctx context.Context, taskID, dependsOnID int64) error {
	query := `DELETE FROM task_dependencies WHERE task_id = $1 AND depends_on_id = $2`
	_, err := s.db.ExecContext(ctx, query, taskID, dependsOnID)
	if err != nil {
		return fmt.Errorf("removing dependency: %w", err)
	}
	return nil
}

// GetTaskWithDependencies fetches a task with its full dependency information.
func (s *Store) GetTaskWithDependencies(ctx context.Context, taskID int64) (*model.TaskWithDependencies, error) {
	// First get the task itself with creator info and column info
	var result model.TaskWithDependencies
	err := s.db.GetContext(ctx, &result, `
		SELECT t.id, t.project_id, t.title, t.description, t.created_by,
			t.column_id, t.position, t.priority, t.metadata, t.due_date,
			t.created_at, t.updated_at, t.deleted_at,
			u.display_name as created_by_name,
			u.email as created_by_email,
			pc.name as column_name,
			pc.color as column_color
		FROM tasks t
		LEFT JOIN users u ON t.created_by = u.id
		LEFT JOIN project_columns pc ON t.column_id = pc.id
		WHERE t.id = $1 AND t.deleted_at IS NULL
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("getting task %d: %w", taskID, err)
	}

	// Get tasks this depends on (blocked by) - WITH column info via JOIN
	err = s.db.SelectContext(ctx, &result.BlockedBy, `
		SELECT
			t.id, t.project_id, t.title, t.description, t.created_by,
			t.column_id, t.position, t.priority, t.metadata, t.due_date,
			t.created_at, t.updated_at, t.deleted_at,
			pc.name as column_name,
			pc.color as column_color
		FROM tasks t
		INNER JOIN task_dependencies td ON t.id = td.depends_on_id
		INNER JOIN project_columns pc ON t.column_id = pc.id
		WHERE td.task_id = $1
		  AND t.deleted_at IS NULL
		  AND pc.deleted_at IS NULL
		ORDER BY t.title
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("getting blocked_by: %w", err)
	}

	// Get tasks that depend on this (blocking) - WITH column info via JOIN
	err = s.db.SelectContext(ctx, &result.Blocking, `
		SELECT
			t.id, t.project_id, t.title, t.description, t.created_by,
			t.column_id, t.position, t.priority, t.metadata, t.due_date,
			t.created_at, t.updated_at, t.deleted_at,
			pc.name as column_name,
			pc.color as column_color
		FROM tasks t
		INNER JOIN task_dependencies td ON t.id = td.task_id
		INNER JOIN project_columns pc ON t.column_id = pc.id
		WHERE td.depends_on_id = $1
		  AND t.deleted_at IS NULL
		  AND pc.deleted_at IS NULL
		ORDER BY t.title
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("getting blocking: %w", err)
	}

	// Check for circular dependencies
	cycle, cyclePath := s.detectCircularDependency(ctx, taskID)
	result.HasCycle = cycle
	result.CyclePath = cyclePath

	// Count comments
	count, err := s.CountTaskComments(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("counting comments: %w", err)
	}
	result.CommentCount = count

	// Get assignees
	assignees, err := s.GetTaskAssignees(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("getting assignees: %w", err)
	}
	result.Assignees = assignees

	return &result, nil
}

// detectCircularDependency checks if a task is part of a circular dependency chain.
func (s *Store) detectCircularDependency(ctx context.Context, taskID int64) (bool, []int64) {
	query := `
		WITH RECURSIVE dependency_chain AS (
			-- Start from the given task
			SELECT task_id, depends_on_id, ARRAY[task_id, depends_on_id] AS path, 1 AS depth
			FROM task_dependencies
			WHERE task_id = $1

			UNION ALL

			-- Follow the dependency chain
			SELECT td.task_id, td.depends_on_id, dc.path || td.depends_on_id, dc.depth + 1
			FROM task_dependencies td
			INNER JOIN dependency_chain dc ON td.task_id = dc.depends_on_id
			WHERE dc.depth < 100  -- Prevent infinite loops
			  AND NOT (td.depends_on_id = ANY(dc.path))  -- Stop if we've seen this task
		)
		SELECT path FROM dependency_chain WHERE depends_on_id = $1 LIMIT 1
	`

	var path []int64
	err := s.db.GetContext(ctx, &path, query, taskID)
	if err != nil {
		// No cycle found (or error - we'll assume no cycle)
		return false, nil
	}

	return true, path
}

// GetDependencies returns the IDs of tasks that the given task depends on.
func (s *Store) GetDependencies(ctx context.Context, taskID int64) ([]int64, error) {
	var deps []int64
	query := `SELECT depends_on_id FROM task_dependencies WHERE task_id = $1`
	err := s.db.SelectContext(ctx, &deps, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("getting dependencies: %w", err)
	}
	return deps, nil
}
