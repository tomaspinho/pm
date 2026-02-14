-- Seed data for cracked-pm.
-- Idempotent: uses ON CONFLICT to skip existing rows.
-- Run with: mise run db:seed

-- Default project
INSERT INTO projects (id, name, description)
VALUES (1, 'My Project', 'A sample project to get started with cracked-pm')
ON CONFLICT (id) DO NOTHING;

-- Example tasks in different statuses
INSERT INTO tasks (id, project_id, title, description, status, position)
VALUES
    (1, 1, 'Set up CI pipeline',         'Configure GitHub Actions for build, lint, and test.',    'todo',        0),
    (2, 1, 'Design landing page',        'Create wireframes and mockups for the landing page.',    'todo',        1),
    (3, 1, 'Implement user auth',        'Add login and registration with session cookies.',       'in_progress', 0),
    (4, 1, 'Write database migrations',  'Create tables for projects, tasks, and users.',          'done',        0)
ON CONFLICT (id) DO NOTHING;

-- Reset sequences to avoid conflicts with future inserts.
SELECT setval('projects_id_seq', (SELECT COALESCE(MAX(id), 0) FROM projects));
SELECT setval('tasks_id_seq',    (SELECT COALESCE(MAX(id), 0) FROM tasks));
