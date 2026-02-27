-- +goose Up
-- Create table for project labels
CREATE TABLE project_labels (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) NOT NULL,
    position INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Unique partial indexes (excluding soft-deleted rows)
CREATE UNIQUE INDEX unique_project_label_name ON project_labels(project_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX unique_project_label_position ON project_labels(project_id, position) WHERE deleted_at IS NULL;
CREATE INDEX idx_project_labels_project ON project_labels(project_id, position) WHERE deleted_at IS NULL;

-- Junction table for task-label many-to-many relationship
CREATE TABLE task_labels (
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    label_id BIGINT NOT NULL REFERENCES project_labels(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (task_id, label_id)
);

CREATE INDEX idx_task_labels_task ON task_labels(task_id);
CREATE INDEX idx_task_labels_label ON task_labels(label_id);

-- +goose Down
DROP TABLE task_labels;
DROP TABLE project_labels;
