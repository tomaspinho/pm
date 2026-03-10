-- +goose Up

CREATE TABLE invitations (
    id          BIGSERIAL PRIMARY KEY,
    org_id      BIGINT NOT NULL REFERENCES organizations(id),
    email       TEXT NOT NULL,
    invited_by  BIGINT NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '7 days',
    accepted_at TIMESTAMPTZ
);

CREATE INDEX idx_invitations_org ON invitations(org_id);
CREATE INDEX idx_invitations_email ON invitations(email) WHERE accepted_at IS NULL;
CREATE INDEX idx_invitations_expires ON invitations(expires_at) WHERE accepted_at IS NULL;

-- +goose Down

DROP TABLE IF EXISTS invitations;