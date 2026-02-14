-- +goose Up
-- Add new columns to tasks table
ALTER TABLE tasks ADD COLUMN author TEXT DEFAULT '';
ALTER TABLE tasks ADD COLUMN metadata JSONB DEFAULT '{}'::jsonb;
ALTER TABLE tasks ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;

-- Create GIN index for efficient JSONB queries
CREATE INDEX idx_tasks_metadata ON tasks USING GIN (metadata);

-- Partial index for active (non-deleted) tasks
CREATE INDEX idx_tasks_active ON tasks(project_id, status, position) WHERE deleted_at IS NULL;

-- Create task_dependencies table
CREATE TABLE task_dependencies (
    task_id       BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (task_id, depends_on_id),
    -- Prevent self-dependencies
    CONSTRAINT no_self_dependency CHECK (task_id != depends_on_id)
);

-- Index for "what blocks this task?" queries
CREATE INDEX idx_task_dependencies_task ON task_dependencies(task_id);

-- Index for "what does this task block?" queries
CREATE INDEX idx_task_dependencies_depends ON task_dependencies(depends_on_id);

-- +goose Down
DROP INDEX IF EXISTS idx_task_dependencies_depends;
DROP INDEX IF EXISTS idx_task_dependencies_task;
DROP TABLE IF EXISTS task_dependencies;
DROP INDEX IF EXISTS idx_tasks_active;
DROP INDEX IF EXISTS idx_tasks_metadata;
ALTER TABLE tasks DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE tasks DROP COLUMN IF EXISTS metadata;
ALTER TABLE tasks DROP COLUMN IF EXISTS author;
