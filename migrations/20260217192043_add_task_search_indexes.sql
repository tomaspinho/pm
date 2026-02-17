-- +goose Up
-- Enable pg_trgm extension for fuzzy text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create trigram indexes for fast text search on title and description
CREATE INDEX idx_tasks_title_trgm ON tasks USING gin (title gin_trgm_ops);
CREATE INDEX idx_tasks_description_trgm ON tasks USING gin (description gin_trgm_ops);

-- Composite index for filtering by project + active tasks
CREATE INDEX IF NOT EXISTS idx_tasks_project_active 
    ON tasks(project_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_tasks_project_active;
DROP INDEX IF EXISTS idx_tasks_description_trgm;
DROP INDEX IF EXISTS idx_tasks_title_trgm;
DROP EXTENSION IF EXISTS pg_trgm;
