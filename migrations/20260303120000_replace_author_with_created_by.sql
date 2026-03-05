-- +goose Up
-- Replace author text column with created_by foreign key to users
ALTER TABLE tasks ADD COLUMN created_by BIGINT REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE tasks DROP COLUMN IF EXISTS author;
CREATE INDEX idx_tasks_created_by ON tasks(created_by);

-- +goose Down
DROP INDEX IF EXISTS idx_tasks_created_by;
ALTER TABLE tasks ADD COLUMN author TEXT DEFAULT '';
ALTER TABLE tasks DROP COLUMN created_by;