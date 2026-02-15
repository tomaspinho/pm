-- +goose Up
CREATE TABLE task_comments (
    id             BIGSERIAL PRIMARY KEY,
    task_id        BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    parent_id      BIGINT REFERENCES task_comments(id) ON DELETE CASCADE,
    content        TEXT NOT NULL CHECK (char_length(content) <= 10000),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    edited_at      TIMESTAMPTZ,
    deleted_at     TIMESTAMPTZ
);

-- Index for "all comments on a task" queries
CREATE INDEX idx_task_comments_task ON task_comments(task_id) WHERE deleted_at IS NULL;

-- Index for "replies to a comment" queries
CREATE INDEX idx_task_comments_parent ON task_comments(parent_id) WHERE deleted_at IS NULL;

-- Index for "comments by user" queries
CREATE INDEX idx_task_comments_user ON task_comments(user_id) WHERE deleted_at IS NULL;

-- Prevent self-replies (comment cannot be its own parent)
ALTER TABLE task_comments ADD CONSTRAINT no_self_reply CHECK (id != parent_id);

-- +goose Down
DROP INDEX IF EXISTS idx_task_comments_user;
DROP INDEX IF EXISTS idx_task_comments_parent;
DROP INDEX IF EXISTS idx_task_comments_task;
DROP TABLE IF EXISTS task_comments;
