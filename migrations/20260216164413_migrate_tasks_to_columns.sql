-- +goose Up
-- Add new column_id field to tasks table
ALTER TABLE tasks ADD COLUMN column_id BIGINT;

-- Make column_id required and add foreign key constraint
ALTER TABLE tasks ALTER COLUMN column_id SET NOT NULL;
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_column
    FOREIGN KEY (column_id) REFERENCES project_columns(id) ON DELETE RESTRICT;

-- Drop old status-based indexes
DROP INDEX IF EXISTS idx_tasks_project_status;
DROP INDEX IF EXISTS idx_tasks_active;

-- Drop old status column
ALTER TABLE tasks DROP COLUMN status;

-- Create new indexes with column_id
CREATE INDEX idx_tasks_project_column ON tasks(project_id, column_id, position);
CREATE INDEX idx_tasks_active ON tasks(project_id, column_id, position)
    WHERE deleted_at IS NULL;

-- +goose Down
-- Add status column back
ALTER TABLE tasks ADD COLUMN status TEXT;

-- Remove foreign key and column_id
ALTER TABLE tasks DROP CONSTRAINT fk_tasks_column;
DROP INDEX IF EXISTS idx_tasks_project_column;
DROP INDEX IF EXISTS idx_tasks_active;
ALTER TABLE tasks DROP COLUMN column_id;

-- Recreate old indexes
CREATE INDEX idx_tasks_project_status ON tasks(project_id, status, position);
CREATE INDEX idx_tasks_active ON tasks(project_id, status, position)
    WHERE deleted_at IS NULL;
