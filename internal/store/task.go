package store

import (
	"context"
	"fmt"

	"cracked-pm/internal/model"
)

// ListTasksByProject returns all tasks for a project ordered by status and position.
func (s *Store) ListTasksByProject(ctx context.Context, projectID int64) ([]model.Task, error) {
	var tasks []model.Task
	err := s.db.SelectContext(ctx, &tasks,
		"SELECT * FROM tasks WHERE project_id = $1 ORDER BY position, id",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing tasks for project %d: %w", projectID, err)
	}
	return tasks, nil
}

// GroupTasksByStatus groups a slice of tasks into a map keyed by status.
func GroupTasksByStatus(tasks []model.Task) map[string][]model.Task {
	grouped := make(map[string][]model.Task)
	for _, s := range model.AllStatuses() {
		grouped[s] = []model.Task{}
	}
	for _, t := range tasks {
		grouped[t.Status] = append(grouped[t.Status], t)
	}
	return grouped
}

// MoveTask updates a task's status and position, shifting other tasks as needed.
func (s *Store) MoveTask(ctx context.Context, taskID int64, newStatus string, newPosition int) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Get the task's current state.
	var task model.Task
	if err := tx.GetContext(ctx, &task, "SELECT * FROM tasks WHERE id = $1 FOR UPDATE", taskID); err != nil {
		return fmt.Errorf("getting task %d: %w", taskID, err)
	}

	// If moving to a different column, close the gap in the old column first.
	if task.Status != newStatus {
		_, err := tx.ExecContext(ctx,
			"UPDATE tasks SET position = position - 1, updated_at = now() WHERE project_id = $1 AND status = $2 AND position > $3",
			task.ProjectID, task.Status, task.Position,
		)
		if err != nil {
			return fmt.Errorf("closing gap in old column: %w", err)
		}
	} else {
		// Same column: close the gap at the old position.
		_, err := tx.ExecContext(ctx,
			"UPDATE tasks SET position = position - 1, updated_at = now() WHERE project_id = $1 AND status = $2 AND position > $3 AND id != $4",
			task.ProjectID, task.Status, task.Position, taskID,
		)
		if err != nil {
			return fmt.Errorf("closing gap in same column: %w", err)
		}
	}

	// Open a gap at the new position in the target column.
	_, err = tx.ExecContext(ctx,
		"UPDATE tasks SET position = position + 1, updated_at = now() WHERE project_id = $1 AND status = $2 AND position >= $3 AND id != $4",
		task.ProjectID, newStatus, newPosition, taskID,
	)
	if err != nil {
		return fmt.Errorf("opening gap in target column: %w", err)
	}

	// Move the task.
	_, err = tx.ExecContext(ctx,
		"UPDATE tasks SET status = $1, position = $2, updated_at = now() WHERE id = $3",
		newStatus, newPosition, taskID,
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
func (s *Store) CreateTask(ctx context.Context, projectID int64, title, description, status string) (model.Task, error) {
	var task model.Task

	// Get max position in the column
	var maxPos int
	err := s.db.GetContext(ctx, &maxPos,
		"SELECT COALESCE(MAX(position), -1) FROM tasks WHERE project_id = $1 AND status = $2",
		projectID, status)
	if err != nil {
		return task, fmt.Errorf("getting max position: %w", err)
	}

	// Insert at max + 1
	err = s.db.GetContext(ctx, &task,
		"INSERT INTO tasks (project_id, title, description, status, position) VALUES ($1, $2, $3, $4, $5) RETURNING *",
		projectID, title, description, status, maxPos+1)
	if err != nil {
		return task, fmt.Errorf("creating task: %w", err)
	}

	return task, nil
}

// DeleteTask removes a task and closes the gap in positions.
func (s *Store) DeleteTask(ctx context.Context, taskID int64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Get task info before deleting
	var task model.Task
	err = tx.GetContext(ctx, &task, "SELECT * FROM tasks WHERE id = $1", taskID)
	if err != nil {
		return fmt.Errorf("getting task %d: %w", taskID, err)
	}

	// Delete the task
	_, err = tx.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1", taskID)
	if err != nil {
		return fmt.Errorf("deleting task %d: %w", taskID, err)
	}

	// Close gap in positions
	_, err = tx.ExecContext(ctx,
		"UPDATE tasks SET position = position - 1, updated_at = now() WHERE project_id = $1 AND status = $2 AND position > $3",
		task.ProjectID, task.Status, task.Position)
	if err != nil {
		return fmt.Errorf("closing gap after delete: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// CountTasksByStatus returns the number of tasks in a given status for a project.
func (s *Store) CountTasksByStatus(ctx context.Context, projectID int64, status string) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM tasks WHERE project_id = $1 AND status = $2",
		projectID, status)
	if err != nil {
		return 0, fmt.Errorf("counting tasks: %w", err)
	}
	return count, nil
}
