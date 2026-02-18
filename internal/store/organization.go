package store

import (
	"context"
	"fmt"

	"pm/internal/model"
)

// CreateOrganization creates a new organization and adds the owner as a member.
func (s *Store) CreateOrganization(ctx context.Context, name string, ownerUserID int64) (*model.Organization, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	var org model.Organization
	err = tx.GetContext(ctx, &org,
		`INSERT INTO organizations (name, owner_user_id) VALUES ($1, $2) RETURNING *`,
		name, ownerUserID)
	if err != nil {
		return nil, fmt.Errorf("creating organization: %w", err)
	}

	// Add owner as a member
	_, err = tx.ExecContext(ctx,
		`INSERT INTO organization_members (organization_id, user_id) VALUES ($1, $2)`,
		org.ID, ownerUserID)
	if err != nil {
		return nil, fmt.Errorf("adding owner as member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return &org, nil
}

// GetOrganization fetches an organization by ID.
func (s *Store) GetOrganization(ctx context.Context, id int64) (*model.Organization, error) {
	var org model.Organization
	err := s.db.GetContext(ctx, &org,
		`SELECT * FROM organizations WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return nil, fmt.Errorf("getting organization %d: %w", id, err)
	}
	return &org, nil
}

// GetUserOrganizations returns all organizations a user belongs to.
func (s *Store) GetUserOrganizations(ctx context.Context, userID int64) ([]model.Organization, error) {
	var orgs []model.Organization
	err := s.db.SelectContext(ctx, &orgs, `
		SELECT o.* FROM organizations o
		INNER JOIN organization_members om ON o.id = om.organization_id
		WHERE om.user_id = $1 AND om.deleted_at IS NULL AND o.deleted_at IS NULL
		ORDER BY o.name
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("listing user organizations: %w", err)
	}
	return orgs, nil
}

// IsMember checks if a user is a member of an organization.
func (s *Store) IsMember(ctx context.Context, orgID, userID int64) (bool, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM organization_members
		WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, orgID, userID)
	if err != nil {
		return false, fmt.Errorf("checking membership: %w", err)
	}
	return count > 0, nil
}

// AddMember adds a user to an organization.
func (s *Store) AddMember(ctx context.Context, orgID, userID int64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO organization_members (organization_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		orgID, userID)
	if err != nil {
		return fmt.Errorf("adding member: %w", err)
	}
	return nil
}

// RemoveMember removes a user from an organization (soft delete).
func (s *Store) RemoveMember(ctx context.Context, orgID, userID int64) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE organization_members SET deleted_at = NOW(), updated_at = NOW() WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		orgID, userID)
	if err != nil {
		return fmt.Errorf("removing member: %w", err)
	}
	return nil
}

// GetOrganizationMembers returns all active members of an organization.
func (s *Store) GetOrganizationMembers(ctx context.Context, orgID int64) ([]model.User, error) {
	var users []model.User
	err := s.db.SelectContext(ctx, &users, `
		SELECT u.* FROM users u
		INNER JOIN organization_members om ON u.id = om.user_id
		WHERE om.organization_id = $1 
		  AND om.deleted_at IS NULL 
		  AND u.deleted_at IS NULL
		ORDER BY u.display_name, u.email
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("getting organization members: %w", err)
	}
	return users, nil
}

// SearchOrganizationMembers searches for members in an organization by display name or email.
// Returns up to 'limit' results ordered by relevance (trigram similarity).
func (s *Store) SearchOrganizationMembers(ctx context.Context, orgID int64, query string, excludeUserID int64, limit int) ([]model.User, error) {
	var users []model.User

	// If query is empty, return recent members (excluding the specified user)
	if query == "" {
		err := s.db.SelectContext(ctx, &users, `
			SELECT u.* FROM users u
			INNER JOIN organization_members om ON u.id = om.user_id
			WHERE om.organization_id = $1 
			  AND u.id != $2
			  AND om.deleted_at IS NULL 
			  AND u.deleted_at IS NULL
			ORDER BY u.updated_at DESC
			LIMIT $3
		`, orgID, excludeUserID, limit)
		if err != nil {
			return nil, fmt.Errorf("getting recent organization members: %w", err)
		}
		return users, nil
	}

	// Search by display name or email using trigram similarity
	err := s.db.SelectContext(ctx, &users, `
		SELECT u.* FROM users u
		INNER JOIN organization_members om ON u.id = om.user_id
		WHERE om.organization_id = $1 
		  AND u.id != $2
		  AND om.deleted_at IS NULL 
		  AND u.deleted_at IS NULL
		  AND (
		    u.display_name ILIKE '%' || $3 || '%'
		    OR u.email ILIKE '%' || $3 || '%'
		  )
		ORDER BY 
		  GREATEST(
		    similarity(u.display_name, $3),
		    similarity(u.email, $3)
		  ) DESC,
		  u.display_name
		LIMIT $4
	`, orgID, excludeUserID, query, limit)
	if err != nil {
		return nil, fmt.Errorf("searching organization members: %w", err)
	}
	return users, nil
}
