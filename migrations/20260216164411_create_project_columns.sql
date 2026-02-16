-- +goose Up
-- Create table for custom project columns
CREATE TABLE project_columns (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) NOT NULL,
    position INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Unique partial indexes (excluding soft-deleted rows)
CREATE UNIQUE INDEX unique_project_column_name ON project_columns(project_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX unique_project_column_position ON project_columns(project_id, position) WHERE deleted_at IS NULL;
CREATE INDEX idx_project_columns_project ON project_columns(project_id, position) WHERE deleted_at IS NULL;

-- Create default columns (To Do, In Progress, Done) for all existing projects
INSERT INTO project_columns (project_id, name, color, position)
SELECT p.id, 'To Do', '#6B7280', 0 
FROM projects p 
WHERE p.deleted_at IS NULL
UNION ALL
SELECT p.id, 'In Progress', '#3B82F6', 1 
FROM projects p 
WHERE p.deleted_at IS NULL
UNION ALL
SELECT p.id, 'Done', '#10B981', 2 
FROM projects p 
WHERE p.deleted_at IS NULL;

-- +goose Down
DROP TABLE project_columns;
