package store

import (
	"context"
	"fmt"

	"pm/internal/model"
)

// GetProjectLabels returns all active labels for a project, ordered alphabetically by name.
func (s *Store) GetProjectLabels(ctx context.Context, projectID int64) ([]model.Label, error) {
	var labels []model.Label
	err := s.db.SelectContext(ctx, &labels,
		"SELECT * FROM project_labels WHERE project_id = $1 AND deleted_at IS NULL ORDER BY name",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing labels for project %d: %w", projectID, err)
	}
	return labels, nil
}

// GetProjectLabel retrieves a single label by ID (excluding soft-deleted).
func (s *Store) GetProjectLabel(ctx context.Context, labelID int64) (*model.Label, error) {
	var label model.Label
	err := s.db.GetContext(ctx, &label,
		"SELECT * FROM project_labels WHERE id = $1 AND deleted_at IS NULL",
		labelID)
	if err != nil {
		return nil, fmt.Errorf("getting label %d: %w", labelID, err)
	}
	return &label, nil
}

// GetTaskLabels returns all labels for a task.
func (s *Store) GetTaskLabels(ctx context.Context, taskID int64) ([]model.Label, error) {
	var labels []model.Label
	err := s.db.SelectContext(ctx, &labels,
		`SELECT l.* FROM project_labels l
		 INNER JOIN task_labels tl ON l.id = tl.label_id
			WHERE tl.task_id = $1 AND l.deleted_at IS NULL
			ORDER BY l.name`,
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing labels for task %d: %w", taskID, err)
	}
	return labels, nil
}

// CreateProjectLabel creates a single label for a project.
func (s *Store) CreateProjectLabel(ctx context.Context, projectID int64, name, color string) (*model.Label, error) {
	var label model.Label
	err := s.db.GetContext(ctx, &label,
		"INSERT INTO project_labels (project_id, name, color) VALUES ($1, $2, $3) RETURNING *",
		projectID, name, color)
	if err != nil {
		return nil, fmt.Errorf("creating label: %w", err)
	}

	return &label, nil
}

// UpdateProjectLabel updates a label's name and/or color.
func (s *Store) UpdateProjectLabel(ctx context.Context, labelID int64, name, color string) error {
	query := `
		UPDATE project_labels
		SET name = $1, color = $2
		WHERE id = $3 AND deleted_at IS NULL
	`
	_, err := s.db.ExecContext(ctx, query, name, color, labelID)
	if err != nil {
		return fmt.Errorf("updating label %d: %w", labelID, err)
	}
	return nil
}

// DeleteProjectLabel soft-deletes a label (removes it from all tasks via CASCADE).
func (s *Store) DeleteProjectLabel(ctx context.Context, labelID int64) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE project_labels SET deleted_at = NOW() WHERE id = $1",
		labelID)
	if err != nil {
		return fmt.Errorf("deleting label %d: %w", labelID, err)
	}
	return nil
}

// AddLabelToTask adds a label to a task (no-op if already assigned).
func (s *Store) AddLabelToTask(ctx context.Context, taskID int64, labelID int64) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO task_labels (task_id, label_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		taskID, labelID)
	if err != nil {
		return fmt.Errorf("adding label %d to task %d: %w", labelID, taskID, err)
	}
	return nil
}

// RemoveLabelFromTask removes a label from a task.
func (s *Store) RemoveLabelFromTask(ctx context.Context, taskID int64, labelID int64) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM task_labels WHERE task_id = $1 AND label_id = $2",
		taskID, labelID)
	if err != nil {
		return fmt.Errorf("removing label %d from task %d: %w", labelID, taskID, err)
	}
	return nil
}

// ValidateLabelOwnership checks if a label belongs to a specific project.
func (s *Store) ValidateLabelOwnership(ctx context.Context, labelID int64, projectID int64) (bool, error) {
	var count int
	err := s.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM project_labels WHERE id = $1 AND project_id = $2 AND deleted_at IS NULL",
		labelID, projectID)
	if err != nil {
		return false, fmt.Errorf("validating label ownership: %w", err)
	}
	return count > 0, nil
}

// CountLabelTasks returns the number of tasks with a specific label.
func (s *Store) CountLabelTasks(ctx context.Context, labelID int64) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM task_labels WHERE label_id = $1",
		labelID)
	if err != nil {
		return 0, fmt.Errorf("counting tasks with label %d: %w", labelID, err)
	}
	return count, nil
}
