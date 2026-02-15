-- +goose Up
-- Add due_date column to tasks table
ALTER TABLE tasks ADD COLUMN due_date DATE DEFAULT NULL;

-- Index for sorting/filtering by due date (partial index excludes NULL and deleted tasks)
CREATE INDEX idx_tasks_due_date ON tasks(due_date) WHERE due_date IS NOT NULL AND deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_tasks_due_date;
ALTER TABLE tasks DROP COLUMN IF EXISTS due_date;
