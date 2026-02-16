-- +goose Up
-- Add new column_id field to tasks table
ALTER TABLE tasks ADD COLUMN column_id BIGINT;

-- Map old status values to new column_id based on project's default columns
UPDATE tasks t
SET column_id = pc.id
FROM project_columns pc
WHERE t.project_id = pc.project_id
  AND pc.deleted_at IS NULL
  AND (
    (t.status = 'todo' AND pc.name = 'To Do' AND pc.position = 0)
    OR (t.status = 'in_progress' AND pc.name = 'In Progress' AND pc.position = 1)
    OR (t.status = 'done' AND pc.name = 'Done' AND pc.position = 2)
  );

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

-- Map column_id back to status based on position
UPDATE tasks t
SET status = CASE 
    WHEN pc.position = 0 THEN 'todo'
    WHEN pc.position = 1 THEN 'in_progress'
    WHEN pc.position = 2 THEN 'done'
    ELSE 'todo'
END
FROM project_columns pc
WHERE t.column_id = pc.id;

-- Remove foreign key and column_id
ALTER TABLE tasks DROP CONSTRAINT fk_tasks_column;
DROP INDEX IF EXISTS idx_tasks_project_column;
DROP INDEX IF EXISTS idx_tasks_active;
ALTER TABLE tasks DROP COLUMN column_id;

-- Recreate old indexes
CREATE INDEX idx_tasks_project_status ON tasks(project_id, status, position);
CREATE INDEX idx_tasks_active ON tasks(project_id, status, position) 
    WHERE deleted_at IS NULL;
