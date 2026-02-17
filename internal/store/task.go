package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"pm/internal/model"
)

// ListTasksByProject returns all active (non-deleted) tasks for a project.
func (s *Store) ListTasksByProject(ctx context.Context, projectID int64) ([]model.Task, error) {
	var tasks []model.Task
	err := s.db.SelectContext(ctx, &tasks,
		"SELECT * FROM tasks WHERE project_id = $1 AND deleted_at IS NULL ORDER BY position, id",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing tasks for project %d: %w", projectID, err)
	}
	return tasks, nil
}

// GetTask retrieves a single task by ID (excluding soft-deleted).
func (s *Store) GetTask(ctx context.Context, taskID int64) (*model.Task, error) {
	var task model.Task
	err := s.db.GetContext(ctx, &task,
		"SELECT * FROM tasks WHERE id = $1 AND deleted_at IS NULL",
		taskID)
	if err != nil {
		return nil, fmt.Errorf("getting task %d: %w", taskID, err)
	}
	return &task, nil
}

// GroupTasksByColumn groups a slice of tasks into a map keyed by column ID.
func GroupTasksByColumn(tasks []model.Task, columns []model.ProjectColumn) map[int64][]model.Task {
	grouped := make(map[int64][]model.Task)
	for _, col := range columns {
		grouped[col.ID] = []model.Task{}
	}
	for _, t := range tasks {
		grouped[t.ColumnID] = append(grouped[t.ColumnID], t)
	}
	return grouped
}

// MoveTask updates a task's column and position, shifting other tasks as needed.
func (s *Store) MoveTask(ctx context.Context, taskID int64, newColumnID int64, newPosition int) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Get the task's current state
	var task model.Task
	if err := tx.GetContext(ctx, &task, "SELECT * FROM tasks WHERE id = $1 AND deleted_at IS NULL FOR UPDATE", taskID); err != nil {
		return fmt.Errorf("getting task %d: %w", taskID, err)
	}

	// If moving to a different column, close the gap in the old column first
	if task.ColumnID != newColumnID {
		_, err := tx.ExecContext(ctx,
			"UPDATE tasks SET position = position - 1, updated_at = now() WHERE project_id = $1 AND column_id = $2 AND position > $3 AND deleted_at IS NULL",
			task.ProjectID, task.ColumnID, task.Position,
		)
		if err != nil {
			return fmt.Errorf("closing gap in old column: %w", err)
		}
	} else {
		// Same column: close the gap at the old position
		_, err := tx.ExecContext(ctx,
			"UPDATE tasks SET position = position - 1, updated_at = now() WHERE project_id = $1 AND column_id = $2 AND position > $3 AND id != $4 AND deleted_at IS NULL",
			task.ProjectID, task.ColumnID, task.Position, taskID,
		)
		if err != nil {
			return fmt.Errorf("closing gap in same column: %w", err)
		}
	}

	// Open a gap at the new position in the target column
	_, err = tx.ExecContext(ctx,
		"UPDATE tasks SET position = position + 1, updated_at = now() WHERE project_id = $1 AND column_id = $2 AND position >= $3 AND id != $4 AND deleted_at IS NULL",
		task.ProjectID, newColumnID, newPosition, taskID,
	)
	if err != nil {
		return fmt.Errorf("opening gap in target column: %w", err)
	}

	// Move the task
	_, err = tx.ExecContext(ctx,
		"UPDATE tasks SET column_id = $1, position = $2, updated_at = now() WHERE id = $3",
		newColumnID, newPosition, taskID,
	)
	if err != nil {
		return fmt.Errorf("updating task %d: %w", taskID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// CreateTask inserts a new task at the end of a column.
func (s *Store) CreateTask(ctx context.Context, projectID int64, title, description string, columnID int64, dueDate *time.Time) (model.Task, error) {
	var task model.Task

	// Get max position in the column
	var maxPos int
	err := s.db.GetContext(ctx, &maxPos,
		"SELECT COALESCE(MAX(position), -1) FROM tasks WHERE project_id = $1 AND column_id = $2 AND deleted_at IS NULL",
		projectID, columnID)
	if err != nil {
		return task, fmt.Errorf("getting max position: %w", err)
	}

	// Insert at max + 1
	err = s.db.GetContext(ctx, &task,
		"INSERT INTO tasks (project_id, title, description, column_id, position, due_date) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *",
		projectID, title, description, columnID, maxPos+1, dueDate)
	if err != nil {
		return task, fmt.Errorf("creating task: %w", err)
	}

	return task, nil
}

// DeleteTask soft-deletes a task and closes the gap in positions.
func (s *Store) DeleteTask(ctx context.Context, taskID int64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Get task info before deleting
	var task model.Task
	err = tx.GetContext(ctx, &task, "SELECT * FROM tasks WHERE id = $1 AND deleted_at IS NULL", taskID)
	if err != nil {
		return fmt.Errorf("getting task %d: %w", taskID, err)
	}

	// Soft delete the task
	_, err = tx.ExecContext(ctx, "UPDATE tasks SET deleted_at = NOW() WHERE id = $1", taskID)
	if err != nil {
		return fmt.Errorf("deleting task %d: %w", taskID, err)
	}

	// Close gap in positions
	_, err = tx.ExecContext(ctx,
		"UPDATE tasks SET position = position - 1, updated_at = now() WHERE project_id = $1 AND column_id = $2 AND position > $3 AND deleted_at IS NULL",
		task.ProjectID, task.ColumnID, task.Position)
	if err != nil {
		return fmt.Errorf("closing gap after delete: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// CountTasksByColumn returns the number of active tasks in a given column for a project.
func (s *Store) CountTasksByColumn(ctx context.Context, projectID int64, columnID int64) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM tasks WHERE project_id = $1 AND column_id = $2 AND deleted_at IS NULL",
		projectID, columnID)
	if err != nil {
		return 0, fmt.Errorf("counting tasks: %w", err)
	}
	return count, nil
}

// UpdateTask updates a task's basic fields.
func (s *Store) UpdateTask(ctx context.Context, taskID int64, title, description, author string, dueDate *time.Time) error {
	query := `
		UPDATE tasks 
		SET title = $1, description = $2, author = $3, due_date = $4, updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL
	`
	_, err := s.db.ExecContext(ctx, query, title, description, author, dueDate, taskID)
	if err != nil {
		return fmt.Errorf("updating task %d: %w", taskID, err)
	}
	return nil
}

// UpdateTaskMetadata updates the entire metadata JSON for a task.
func (s *Store) UpdateTaskMetadata(ctx context.Context, taskID int64, metadata model.TaskMetadata) error {
	query := `
		UPDATE tasks 
		SET metadata = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err := s.db.ExecContext(ctx, query, metadata, taskID)
	if err != nil {
		return fmt.Errorf("updating task metadata: %w", err)
	}
	return nil
}

// parseMetadataValue converts user input to valid JSON for PostgreSQL jsonb.
// Objects {} and arrays [] are preserved as-is.
// Everything else (strings, numbers, booleans, null) is converted to a JSON string.
func parseMetadataValue(value string) (string, error) {
	// Trim whitespace
	trimmed := strings.TrimSpace(value)

	// Empty string becomes empty JSON string
	if trimmed == "" {
		return `""`, nil
	}

	// Try to parse as JSON to detect valid JSON
	var jsonTest interface{}
	if err := json.Unmarshal([]byte(trimmed), &jsonTest); err == nil {
		// Valid JSON - check if it's an object or array
		if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
			(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
			// Object or array - return as-is
			return trimmed, nil
		}
	}

	// Not an object/array (or invalid JSON) - treat as plain string
	// Use json.Marshal to properly escape the string
	marshaled, err := json.Marshal(trimmed)
	if err != nil {
		return "", fmt.Errorf("marshaling string value: %w", err)
	}
	return string(marshaled), nil
}

// SetMetadataKey sets a single key in the task's metadata.
func (s *Store) SetMetadataKey(ctx context.Context, taskID int64, key string, value interface{}) error {
	// Convert value to string if it isn't already
	valueStr, ok := value.(string)
	if !ok {
		valueStr = fmt.Sprintf("%v", value)
	}

	// Parse the value to ensure it's valid JSON
	jsonValue, err := parseMetadataValue(valueStr)
	if err != nil {
		return fmt.Errorf("parsing metadata value: %w", err)
	}

	query := `
		UPDATE tasks 
		SET metadata = jsonb_set(metadata, $1, $2::jsonb, true),
		    updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`
	// PostgreSQL jsonb_set requires path as text array like '{key}'
	path := fmt.Sprintf("{%s}", key)
	_, err = s.db.ExecContext(ctx, query, path, jsonValue, taskID)
	if err != nil {
		return fmt.Errorf("setting metadata key: %w", err)
	}
	return nil
}

// DeleteMetadataKey removes a key from the task's metadata.
func (s *Store) DeleteMetadataKey(ctx context.Context, taskID int64, key string) error {
	query := `
		UPDATE tasks 
		SET metadata = metadata - $1,
		    updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err := s.db.ExecContext(ctx, query, key, taskID)
	if err != nil {
		return fmt.Errorf("deleting metadata key: %w", err)
	}
	return nil
}

// UpdateTaskColumn updates a task's column and moves it to the end of the new column.
// Returns the old column ID for updating the UI.
func (s *Store) UpdateTaskColumn(ctx context.Context, taskID int64, newColumnID int64) (int64, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Get the task's current state
	var task model.Task
	if err := tx.GetContext(ctx, &task, "SELECT * FROM tasks WHERE id = $1 AND deleted_at IS NULL FOR UPDATE", taskID); err != nil {
		return 0, fmt.Errorf("getting task %d: %w", taskID, err)
	}

	oldColumnID := task.ColumnID

	// If column hasn't changed, nothing to do
	if task.ColumnID == newColumnID {
		return oldColumnID, nil
	}

	// Close gap in old column
	_, err = tx.ExecContext(ctx,
		"UPDATE tasks SET position = position - 1, updated_at = now() WHERE project_id = $1 AND column_id = $2 AND position > $3 AND deleted_at IS NULL",
		task.ProjectID, task.ColumnID, task.Position,
	)
	if err != nil {
		return 0, fmt.Errorf("closing gap in old column: %w", err)
	}

	// Get max position in new column
	var maxPos int
	err = tx.GetContext(ctx, &maxPos,
		"SELECT COALESCE(MAX(position), -1) FROM tasks WHERE project_id = $1 AND column_id = $2 AND deleted_at IS NULL",
		task.ProjectID, newColumnID)
	if err != nil {
		return 0, fmt.Errorf("getting max position in new column: %w", err)
	}

	// Move task to end of new column
	_, err = tx.ExecContext(ctx,
		"UPDATE tasks SET column_id = $1, position = $2, updated_at = NOW() WHERE id = $3",
		newColumnID, maxPos+1, taskID,
	)
	if err != nil {
		return 0, fmt.Errorf("updating task column: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("committing transaction: %w", err)
	}
	return oldColumnID, nil
}

// SearchOrganizationTasks searches for tasks across all projects in an organization
// using ILIKE for substring matches and PostgreSQL trigram similarity for fuzzy matching.
// Prioritizes exact substring matches, then fuzzy matches, then recent tasks.
// Returns up to 'limit' results ordered by relevance.
func (s *Store) SearchOrganizationTasks(ctx context.Context, orgID int64, query string, excludeTaskID int64, limit int) ([]model.TaskSearchResult, error) {
	sql := `
		SELECT 
			t.id,
			t.title,
			LEFT(t.description, 100) as description,
			t.project_id,
			p.name as project_name,
			t.column_id,
			pc.name as column_name,
			pc.color as column_color,
			t.author,
			t.due_date
		FROM tasks t
		INNER JOIN projects p ON t.project_id = p.id
		INNER JOIN project_columns pc ON t.column_id = pc.id
		WHERE p.organization_id = $1
		  AND t.deleted_at IS NULL
		  AND p.deleted_at IS NULL
		  AND pc.deleted_at IS NULL
		  AND t.id != $2
		  AND (
			  t.title ILIKE '%' || $3 || '%'
			  OR t.description ILIKE '%' || $3 || '%'
			  OR similarity(t.title, $3) > 0.2
			  OR similarity(t.description, $3) > 0.15
		  )
		ORDER BY 
			CASE WHEN t.title ILIKE '%' || $3 || '%' THEN 1 ELSE 2 END,
			similarity(t.title, $3) DESC,
			similarity(t.description, $3) DESC,
			t.updated_at DESC
		LIMIT $4
	`

	var results []model.TaskSearchResult
	err := s.db.SelectContext(ctx, &results, sql, orgID, excludeTaskID, query, limit)
	if err != nil {
		return nil, fmt.Errorf("searching tasks: %w", err)
	}

	return results, nil
}

// GetRecentOrganizationTasks returns the most recently updated tasks across all
// projects in an organization. Used for the default dropdown state.
func (s *Store) GetRecentOrganizationTasks(ctx context.Context, orgID int64, excludeTaskID int64, limit int) ([]model.TaskSearchResult, error) {
	sql := `
		SELECT 
			t.id,
			t.title,
			LEFT(t.description, 100) as description,
			t.project_id,
			p.name as project_name,
			t.column_id,
			pc.name as column_name,
			pc.color as column_color,
			t.author,
			t.due_date
		FROM tasks t
		INNER JOIN projects p ON t.project_id = p.id
		INNER JOIN project_columns pc ON t.column_id = pc.id
		WHERE p.organization_id = $1
		  AND t.deleted_at IS NULL
		  AND p.deleted_at IS NULL
		  AND pc.deleted_at IS NULL
		  AND t.id != $2
		ORDER BY t.updated_at DESC
		LIMIT $3
	`

	var results []model.TaskSearchResult
	err := s.db.SelectContext(ctx, &results, sql, orgID, excludeTaskID, limit)
	if err != nil {
		return nil, fmt.Errorf("getting recent tasks: %w", err)
	}

	return results, nil
}

// WouldCreateCycle checks if adding a dependency would create a circular dependency chain
func (s *Store) WouldCreateCycle(ctx context.Context, taskID, dependsOnID int64) (bool, error) {
	// Check if dependsOnID eventually depends on taskID (which would create a cycle)
	query := `
		WITH RECURSIVE dependency_chain AS (
			SELECT task_id, depends_on_id, 1 AS depth
			FROM task_dependencies
			WHERE task_id = $1
			
			UNION ALL
			
			SELECT td.task_id, td.depends_on_id, dc.depth + 1
			FROM task_dependencies td
			INNER JOIN dependency_chain dc ON td.task_id = dc.depends_on_id
			WHERE dc.depth < 100
		)
		SELECT EXISTS (
			SELECT 1 FROM dependency_chain WHERE depends_on_id = $2
		)
	`

	var wouldCycle bool
	err := s.db.GetContext(ctx, &wouldCycle, query, dependsOnID, taskID)
	if err != nil {
		return false, fmt.Errorf("checking for cycles: %w", err)
	}

	return wouldCycle, nil
}
