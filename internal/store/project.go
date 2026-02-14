package store

import (
	"context"
	"fmt"

	"cracked-pm/internal/model"
)

// GetProject fetches a single project by ID.
func (s *Store) GetProject(ctx context.Context, id int64) (model.Project, error) {
	var p model.Project
	err := s.db.GetContext(ctx, &p, "SELECT * FROM projects WHERE id = $1", id)
	if err != nil {
		return p, fmt.Errorf("getting project %d: %w", id, err)
	}
	return p, nil
}
