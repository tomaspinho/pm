-- +goose Up
-- Create task_assignees junction table for many-to-many relationship
CREATE TABLE task_assignees (
    task_id     BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (task_id, user_id)
);

-- Index for "who is assigned to this task?" queries
CREATE INDEX idx_task_assignees_task ON task_assignees(task_id);

-- Index for "what tasks is this user assigned to?" queries
CREATE INDEX idx_task_assignees_user ON task_assignees(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_task_assignees_user;
DROP INDEX IF EXISTS idx_task_assignees_task;
DROP TABLE IF EXISTS task_assignees;
