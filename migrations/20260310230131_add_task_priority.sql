-- +goose Up
ALTER TABLE tasks ADD COLUMN priority INT NOT NULL DEFAULT 2;
CREATE INDEX idx_tasks_column_priority ON tasks(column_id, priority, position) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_tasks_column_priority;
ALTER TABLE tasks DROP COLUMN IF EXISTS priority;