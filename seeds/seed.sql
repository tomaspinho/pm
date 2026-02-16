-- Seed data for pm.
-- Idempotent: uses ON CONFLICT to skip existing rows.
-- Run with: mise run db:seed

-- Default user (password: changeme123)
INSERT INTO users (id, email, password_hash)
VALUES (1, 'admin@example.com', '$2a$12$r61zoAW6/P0GJSm2cnx7X.2Cbv2H1J0jxkR4qZ/PMBa08p9vY/aJS')
ON CONFLICT (id) DO NOTHING;

-- Default organization for the user
INSERT INTO organizations (id, name, owner_user_id)
VALUES (1, 'admin@example.com''s Organization', 1)
ON CONFLICT (id) DO NOTHING;

-- Add user as member of their organization
INSERT INTO organization_members (organization_id, user_id)
VALUES (1, 1)
ON CONFLICT (organization_id, user_id) DO NOTHING;

-- Default project (now owned by the organization)
INSERT INTO projects (id, name, description, organization_id)
VALUES (1, 'My Project', 'A sample project to get started with pm', 1)
ON CONFLICT (id) DO NOTHING;

-- Default columns for the sample project
-- Note: Using specific IDs (1, 2, 3) so tasks can reference them
INSERT INTO project_columns (id, project_id, name, color, position)
VALUES
    (1, 1, 'To Do', '#6B7280', 0),
    (2, 1, 'In Progress', '#3B82F6', 1),
    (3, 1, 'Done', '#10B981', 2)
ON CONFLICT (id) DO NOTHING;

-- Example tasks in different columns
-- Note: column_id 1 = To Do, 2 = In Progress, 3 = Done (from columns above)
INSERT INTO tasks (id, project_id, title, description, column_id, position, author, metadata)
VALUES
    (1, 1, 'Set up CI pipeline',         'Configure GitHub Actions for build, lint, and test.',    1, 0, 'alice@example.com', '{"priority": "high", "labels": ["devops", "infrastructure"]}'),
    (2, 1, 'Design landing page',        'Create wireframes and mockups for the landing page.',    1, 1, 'bob@example.com',   '{"priority": "medium", "labels": ["design", "frontend"]}'),
    (3, 1, 'Implement user auth',        'Add login and registration with session cookies.',       2, 0, 'alice@example.com', '{"priority": "high", "labels": ["backend", "security"]}'),
    (4, 1, 'Write database migrations',  'Create tables for projects, tasks, and users.',          3, 0, 'charlie@example.com', '{"priority": "high", "labels": ["database"]}')
ON CONFLICT (id) DO NOTHING;

-- Example task dependencies
INSERT INTO task_dependencies (task_id, depends_on_id)
VALUES
    (3, 4)  -- "Implement user auth" depends on "Write database migrations"
ON CONFLICT (task_id, depends_on_id) DO NOTHING;

-- Reset sequences to avoid conflicts with future inserts.
SELECT setval('users_id_seq', (SELECT COALESCE(MAX(id), 0) FROM users));
SELECT setval('organizations_id_seq', (SELECT COALESCE(MAX(id), 0) FROM organizations));
SELECT setval('projects_id_seq', (SELECT COALESCE(MAX(id), 0) FROM projects));
SELECT setval('project_columns_id_seq', (SELECT COALESCE(MAX(id), 0) FROM project_columns));
SELECT setval('tasks_id_seq', (SELECT COALESCE(MAX(id), 1) FROM tasks));
