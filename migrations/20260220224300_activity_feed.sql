-- +goose Up
CREATE TABLE task_activity (
    id          BIGSERIAL PRIMARY KEY,
    task_id     BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action      TEXT NOT NULL,
    field_name  TEXT,
    old_value   JSONB,
    new_value   JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_activity_task ON task_activity(task_id);
CREATE INDEX idx_task_activity_created_at ON task_activity(created_at);
CREATE INDEX idx_task_activity_user ON task_activity(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_task_activity_user;
DROP INDEX IF EXISTS idx_task_activity_created_at;
DROP INDEX IF EXISTS idx_task_activity_task;
DROP TABLE IF EXISTS task_activity;
