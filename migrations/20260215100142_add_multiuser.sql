-- +goose Up

-- Users table
CREATE TABLE users (
    id          BIGSERIAL   PRIMARY KEY,
    email       TEXT        UNIQUE NOT NULL,
    password_hash TEXT      NOT NULL,
    last_viewed_project_id BIGINT,
    last_viewed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;

-- Organizations table
CREATE TABLE organizations (
    id          BIGSERIAL   PRIMARY KEY,
    name        TEXT        NOT NULL,
    owner_user_id BIGINT   NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_organizations_owner ON organizations(owner_user_id);
CREATE INDEX idx_organizations_deleted_at ON organizations(deleted_at) WHERE deleted_at IS NOT NULL;

-- Organization members junction table
CREATE TABLE organization_members (
    organization_id BIGINT  NOT NULL REFERENCES organizations(id),
    user_id         BIGINT  NOT NULL REFERENCES users(id),
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    PRIMARY KEY (organization_id, user_id)
);

CREATE INDEX idx_org_members_user ON organization_members(user_id);
CREATE INDEX idx_org_members_deleted_at ON organization_members(deleted_at) WHERE deleted_at IS NOT NULL;

-- Sessions table
CREATE TABLE sessions (
    id          TEXT        PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users(id),
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

-- Add organization_id to projects (allow NULL temporarily for migration)
ALTER TABLE projects ADD COLUMN organization_id BIGINT REFERENCES organizations(id);
ALTER TABLE projects ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX idx_projects_org ON projects(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_deleted_at ON projects(deleted_at) WHERE deleted_at IS NOT NULL;

-- Add foreign key from users.last_viewed_project_id to projects
-- (deferred because projects table must exist first)
ALTER TABLE users ADD CONSTRAINT fk_users_last_viewed_project
    FOREIGN KEY (last_viewed_project_id) REFERENCES projects(id) ON DELETE SET NULL;

-- Seed: create default user, org, and migrate existing projects
-- Password is bcrypt hash of "changeme123" (cost 10)
INSERT INTO users (id, email, password_hash)
VALUES (1, 'admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy');

INSERT INTO organizations (id, name, owner_user_id)
VALUES (1, 'admin@example.com''s Organization', 1);

INSERT INTO organization_members (organization_id, user_id)
VALUES (1, 1);

-- Migrate all existing projects to the default organization
UPDATE projects SET organization_id = 1 WHERE organization_id IS NULL;

-- Now make organization_id NOT NULL
ALTER TABLE projects ALTER COLUMN organization_id SET NOT NULL;

-- Reset sequences
SELECT setval('users_id_seq', (SELECT COALESCE(MAX(id), 0) FROM users));
SELECT setval('organizations_id_seq', (SELECT COALESCE(MAX(id), 0) FROM organizations));

-- +goose Down

ALTER TABLE projects DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE projects DROP COLUMN IF EXISTS organization_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_last_viewed_project;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS users;
