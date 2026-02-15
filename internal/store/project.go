package store

import (
	"context"
	"fmt"

	"cracked-pm/internal/model"
)

// GetProject fetches a single project by ID.
func (s *Store) GetProject(ctx context.Context, id int64) (model.Project, error) {
	var p model.Project
	err := s.db.GetContext(ctx, &p,
		"SELECT * FROM projects WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return p, fmt.Errorf("getting project %d: %w", id, err)
	}
	return p, nil
}

// GetOrganizationProjects returns all active projects for an organization.
func (s *Store) GetOrganizationProjects(ctx context.Context, orgID int64) ([]model.Project, error) {
	var projects []model.Project
	err := s.db.SelectContext(ctx, &projects, `
		SELECT * FROM projects
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY name
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("listing organization projects: %w", err)
	}
	return projects, nil
}

// CreateProject inserts a new project.
func (s *Store) CreateProject(ctx context.Context, name, description string, orgID int64) (*model.Project, error) {
	var p model.Project
	err := s.db.GetContext(ctx, &p, `
		INSERT INTO projects (name, description, organization_id)
		VALUES ($1, $2, $3) RETURNING *
	`, name, description, orgID)
	if err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}
	return &p, nil
}
