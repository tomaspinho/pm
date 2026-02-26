package store

import (
	"context"
	"fmt"

	"pm/internal/model"
)

// GetTaskAssignees returns all users assigned to a task.
func (s *Store) GetTaskAssignees(ctx context.Context, taskID int64) ([]model.AssigneeInfo, error) {
	var assignees []model.AssigneeInfo
	err := s.db.SelectContext(ctx, &assignees, `
		SELECT u.id as user_id, u.email, u.display_name
		FROM task_assignees ta
		INNER JOIN users u ON ta.user_id = u.id
		WHERE ta.task_id = $1 AND u.deleted_at IS NULL
		ORDER BY u.display_name, u.email
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("getting task assignees: %w", err)
	}
	return assignees, nil
}

// AssignUserToTask assigns a user to a task.
func (s *Store) AssignUserToTask(ctx context.Context, taskID, userID int64) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO task_assignees (task_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (task_id, user_id) DO NOTHING
	`, taskID, userID)
	if err != nil {
		return fmt.Errorf("assigning user %d to task %d: %w", userID, taskID, err)
	}
	return nil
}

// UnassignUserFromTask removes a user from a task.
func (s *Store) UnassignUserFromTask(ctx context.Context, taskID, userID int64) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM task_assignees
		WHERE task_id = $1 AND user_id = $2
	`, taskID, userID)
	if err != nil {
		return fmt.Errorf("unassigning user %d from task %d: %w", userID, taskID, err)
	}
	return nil
}

// IsUserAssigned checks if a user is assigned to a task.
func (s *Store) IsUserAssigned(ctx context.Context, taskID, userID int64) (bool, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM task_assignees
		WHERE task_id = $1 AND user_id = $2
	`, taskID, userID)
	if err != nil {
		return false, fmt.Errorf("checking if user is assigned: %w", err)
	}
	return count > 0, nil
}

// GetUserTasks returns all tasks assigned to a user in an organization.
func (s *Store) GetUserTasks(ctx context.Context, orgID, userID int64) ([]model.Task, error) {
	var tasks []model.Task
	err := s.db.SelectContext(ctx, &tasks, `
		SELECT t.* FROM tasks t
		INNER JOIN task_assignees ta ON t.id = ta.task_id
		INNER JOIN projects p ON t.project_id = p.id
		WHERE ta.user_id = $1
		  AND p.organization_id = $2
		  AND t.deleted_at IS NULL
		  AND p.deleted_at IS NULL
		ORDER BY t.updated_at DESC
	`, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("getting user tasks: %w", err)
	}
	return tasks, nil
}
