package store

import (
	"context"
	"fmt"
	"time"

	"pm/internal/model"
)

func (s *Store) CreateInvitation(ctx context.Context, orgID int64, email string, invitedBy int64) (*model.Invitation, error) {
	var inv model.Invitation
	err := s.db.GetContext(ctx, &inv, `
		INSERT INTO invitations (org_id, email, invited_by)
		VALUES ($1, $2, $3)
		RETURNING *
	`, orgID, email, invitedBy)
	if err != nil {
		return nil, fmt.Errorf("creating invitation: %w", err)
	}
	return &inv, nil
}

func (s *Store) GetPendingInvitationsByOrg(ctx context.Context, orgID int64) ([]model.Invitation, error) {
	var invs []model.Invitation
	err := s.db.SelectContext(ctx, &invs, `
		SELECT * FROM invitations
		WHERE org_id = $1 AND accepted_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("getting pending invitations: %w", err)
	}
	return invs, nil
}

func (s *Store) GetPendingInvitationsByEmail(ctx context.Context, email string) ([]model.Invitation, error) {
	var invs []model.Invitation
	err := s.db.SelectContext(ctx, &invs, `
		SELECT * FROM invitations
		WHERE email = $1 AND accepted_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC
	`, email)
	if err != nil {
		return nil, fmt.Errorf("getting pending invitations by email: %w", err)
	}
	return invs, nil
}

func (s *Store) AcceptInvitation(ctx context.Context, invitationID int64) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE invitations SET accepted_at = $1 WHERE id = $2 AND accepted_at IS NULL
	`, time.Now(), invitationID)
	if err != nil {
		return fmt.Errorf("accepting invitation: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("invitation not found or already accepted")
	}
	return nil
}

func (s *Store) DeleteInvitation(ctx context.Context, invitationID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM invitations WHERE id = $1`, invitationID)
	if err != nil {
		return fmt.Errorf("deleting invitation: %w", err)
	}
	return nil
}

func (s *Store) DeleteExpiredInvitations(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM invitations WHERE expires_at < NOW() AND accepted_at IS NULL
	`)
	if err != nil {
		return 0, fmt.Errorf("deleting expired invitations: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("checking rows affected: %w", err)
	}
	return rows, nil
}
