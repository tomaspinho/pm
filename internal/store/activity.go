package store

import (
	"context"
	"encoding/json"
	"fmt"
	"pm/internal/model"
)

// CreateActivity inserts a new activity record.
func (s *Store) CreateActivity(ctx context.Context, taskID, userID int64, action, fieldName string, oldValue, newValue any) error {
	var oldJSON, newJSON []byte
	var err error

	if oldValue != nil {
		oldJSON, err = json.Marshal(oldValue)
		if err != nil {
			return fmt.Errorf("marshaling old value: %w", err)
		}
	}

	if newValue != nil {
		newJSON, err = json.Marshal(newValue)
		if err != nil {
			return fmt.Errorf("marshaling new value: %w", err)
		}
	}

	query := `INSERT INTO task_activity (task_id, user_id, action, field_name, old_value, new_value) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = s.db.ExecContext(ctx, query, taskID, userID, action, fieldName, oldJSON, newJSON)
	if err != nil {
		return fmt.Errorf("creating activity: %w", err)
	}
	return nil
}

// GetTaskActivity retrieves all activity for a task, ordered by creation time (oldest first).
func (s *Store) GetTaskActivity(ctx context.Context, taskID int64) ([]model.TaskActivityRecord, error) {
	var records []model.TaskActivityRecord
	err := s.db.SelectContext(ctx, &records, `
		SELECT 
			a.id,
			a.action,
			a.field_name,
			a.old_value,
			a.new_value,
			a.created_at,
			u.display_name as user_display_name,
			u.email as user_email
		FROM task_activity a
		JOIN users u ON a.user_id = u.id
		WHERE a.task_id = $1
		ORDER BY a.created_at ASC
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("getting activity for task %d: %w", taskID, err)
	}
	return records, nil
}
