package store

import (
	"context"
	"fmt"

	"pm/internal/model"
)

// GetProjectColumns returns all active columns for a project, ordered by position.
func (s *Store) GetProjectColumns(ctx context.Context, projectID int64) ([]model.ProjectColumn, error) {
	var columns []model.ProjectColumn
	err := s.db.SelectContext(ctx, &columns,
		"SELECT * FROM project_columns WHERE project_id = $1 AND deleted_at IS NULL ORDER BY position",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing columns for project %d: %w", projectID, err)
	}
	return columns, nil
}

// GetProjectColumn retrieves a single column by ID (excluding soft-deleted).
func (s *Store) GetProjectColumn(ctx context.Context, columnID int64) (*model.ProjectColumn, error) {
	var column model.ProjectColumn
	err := s.db.GetContext(ctx, &column,
		"SELECT * FROM project_columns WHERE id = $1 AND deleted_at IS NULL",
		columnID)
	if err != nil {
		return nil, fmt.Errorf("getting column %d: %w", columnID, err)
	}
	return &column, nil
}

// CreateDefaultColumns creates the default set of columns for a new project.
func (s *Store) CreateDefaultColumns(ctx context.Context, projectID int64) error {
	defaultCols := model.DefaultColumns()
	return s.CreateProjectColumns(ctx, projectID, defaultCols)
}

// CreateProjectColumns creates multiple columns for a project.
func (s *Store) CreateProjectColumns(ctx context.Context, projectID int64, columns []model.ProjectColumn) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	for i, col := range columns {
		_, err := tx.ExecContext(ctx,
			"INSERT INTO project_columns (project_id, name, color, position) VALUES ($1, $2, $3, $4)",
			projectID, col.Name, col.Color, i,
		)
		if err != nil {
			return fmt.Errorf("creating column %q: %w", col.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// CreateProjectColumn creates a single column for a project.
func (s *Store) CreateProjectColumn(ctx context.Context, projectID int64, name, color string) (*model.ProjectColumn, error) {
	var column model.ProjectColumn

	// Get max position
	var maxPos int
	err := s.db.GetContext(ctx, &maxPos,
		"SELECT COALESCE(MAX(position), -1) FROM project_columns WHERE project_id = $1 AND deleted_at IS NULL",
		projectID)
	if err != nil {
		return nil, fmt.Errorf("getting max position: %w", err)
	}

	// Insert at max + 1
	err = s.db.GetContext(ctx, &column,
		"INSERT INTO project_columns (project_id, name, color, position) VALUES ($1, $2, $3, $4) RETURNING *",
		projectID, name, color, maxPos+1)
	if err != nil {
		return nil, fmt.Errorf("creating column: %w", err)
	}

	return &column, nil
}

// UpdateProjectColumn updates a column's name and/or color.
func (s *Store) UpdateProjectColumn(ctx context.Context, columnID int64, name, color string) error {
	query := `
		UPDATE project_columns 
		SET name = $1, color = $2
		WHERE id = $3 AND deleted_at IS NULL
	`
	_, err := s.db.ExecContext(ctx, query, name, color, columnID)
	if err != nil {
		return fmt.Errorf("updating column %d: %w", columnID, err)
	}
	return nil
}

// DeleteProjectColumn soft-deletes a column and moves all its tasks to a target column.
func (s *Store) DeleteProjectColumn(ctx context.Context, columnID, targetColumnID int64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Get the column being deleted
	var column model.ProjectColumn
	err = tx.GetContext(ctx, &column,
		"SELECT * FROM project_columns WHERE id = $1 AND deleted_at IS NULL FOR UPDATE",
		columnID)
	if err != nil {
		return fmt.Errorf("getting column %d: %w", columnID, err)
	}

	// Get max position in target column
	var maxPos int
	err = tx.GetContext(ctx, &maxPos,
		"SELECT COALESCE(MAX(position), -1) FROM tasks WHERE column_id = $1 AND deleted_at IS NULL",
		targetColumnID)
	if err != nil {
		return fmt.Errorf("getting max position in target column: %w", err)
	}

	// Move all tasks from deleted column to target column
	// They'll be appended at the end, maintaining their relative order
	_, err = tx.ExecContext(ctx, `
		UPDATE tasks 
		SET column_id = $1, 
		    position = position + $2 + 1,
		    updated_at = NOW()
		WHERE column_id = $3 AND deleted_at IS NULL
	`, targetColumnID, maxPos, columnID)
	if err != nil {
		return fmt.Errorf("moving tasks to target column: %w", err)
	}

	// Soft delete the column
	_, err = tx.ExecContext(ctx,
		"UPDATE project_columns SET deleted_at = NOW() WHERE id = $1",
		columnID)
	if err != nil {
		return fmt.Errorf("deleting column %d: %w", columnID, err)
	}

	// Close gap in column positions
	_, err = tx.ExecContext(ctx,
		"UPDATE project_columns SET position = position - 1 WHERE project_id = $1 AND position > $2 AND deleted_at IS NULL",
		column.ProjectID, column.Position)
	if err != nil {
		return fmt.Errorf("closing gap in column positions: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// ReorderColumns updates the position of all columns in a project based on the provided order.
func (s *Store) ReorderColumns(ctx context.Context, projectID int64, columnIDs []int64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	for i, columnID := range columnIDs {
		_, err := tx.ExecContext(ctx,
			"UPDATE project_columns SET position = $1 WHERE id = $2 AND project_id = $3 AND deleted_at IS NULL",
			i, columnID, projectID,
		)
		if err != nil {
			return fmt.Errorf("updating position for column %d: %w", columnID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// CountColumnTasks returns the number of active tasks in a column.
func (s *Store) CountColumnTasks(ctx context.Context, columnID int64) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM tasks WHERE column_id = $1 AND deleted_at IS NULL",
		columnID)
	if err != nil {
		return 0, fmt.Errorf("counting tasks in column %d: %w", columnID, err)
	}
	return count, nil
}

// ValidateColumnOwnership checks if a column belongs to a specific project.
func (s *Store) ValidateColumnOwnership(ctx context.Context, columnID, projectID int64) (bool, error) {
	var count int
	err := s.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM project_columns WHERE id = $1 AND project_id = $2 AND deleted_at IS NULL",
		columnID, projectID)
	if err != nil {
		return false, fmt.Errorf("validating column ownership: %w", err)
	}
	return count > 0, nil
}
