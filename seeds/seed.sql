-- Seed data for pm.
-- Idempotent: uses ON CONFLICT to skip existing rows.
-- Run with: mise run db:seed

-- Default user (password: changeme123)
-- Note: display_name is required now, updating existing user
UPDATE users SET display_name = 'Admin User' WHERE id = 1 AND (display_name = '' OR display_name = 'admin');

INSERT INTO users (id, email, password_hash, display_name)
VALUES (1, 'admin@example.com', '$2a$12$r61zoAW6/P0GJSm2cnx7X.2Cbv2H1J0jxkR4qZ/PMBa08p9vY/aJS', 'Admin User')
ON CONFLICT (id) DO UPDATE SET display_name = EXCLUDED.display_name;

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

-- Default labels for project 1 (My Project)
INSERT INTO project_labels (id, project_id, name, color)
VALUES
    (1, 1, 'Bug', '#EF4444'),
    (2, 1, 'Feature', '#3B82F6'),
    (3, 1, 'Documentation', '#6B7280'),
    (4, 1, 'Urgent', '#F59E0B'),
    (5, 1, 'DevOps', '#8B5CF6'),
    (6, 1, 'Backend', '#10B981'),
    (7, 1, 'Frontend', '#EC4899'),
    (8, 1, 'Security', '#14B8A6')
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

-- Task labels for project 1
INSERT INTO task_labels (task_id, label_id)
VALUES
    (1, 5), (1, 6), (1, 2),  -- Set up CI pipeline: DevOps, Backend, Feature
    (2, 7), (2, 3), (2, 2),  -- Design landing page: Frontend, Documentation, Feature
    (3, 6), (3, 8), (3, 4),  -- Implement user auth: Backend, Security, Urgent
    (4, 6), (4, 3)           -- Write database migrations: Backend, Documentation
ON CONFLICT (task_id, label_id) DO NOTHING;

-- Second project with 100 tasks for testing search and dependencies
INSERT INTO projects (id, name, description, organization_id)
VALUES (2, 'E-Commerce Platform', 'Large scale e-commerce application with 100+ tasks', 1)
ON CONFLICT (id) DO NOTHING;

-- Columns for the second project
INSERT INTO project_columns (id, project_id, name, color, position)
VALUES
    (4, 2, 'Backlog', '#6B7280', 0),
    (5, 2, 'Active', '#F59E0B', 1),
    (6, 2, 'Completed', '#10B981', 2)
ON CONFLICT (id) DO NOTHING;

-- Labels for project 2 (E-Commerce Platform)
INSERT INTO project_labels (id, project_id, name, color)
VALUES
    (9, 2, 'Critical', '#EF4444'),
    (10, 2, 'High Priority', '#F59E0B'),
    (11, 2, 'Development', '#3B82F6'),
    (12, 2, 'Design', '#EC4899'),
    (13, 2, 'Infrastructure', '#8B5CF6'),
    (14, 2, 'Testing', '#10B981'),
    (15, 2, 'Frontend', '#14B8A6'),
    (16, 2, 'Backend', '#6B7280')
ON CONFLICT (id) DO NOTHING;

-- 100 tasks spread across the 3 columns (roughly 40/35/25 distribution)
INSERT INTO tasks (id, project_id, title, description, column_id, position, author, metadata)
VALUES
    -- Backlog (40 tasks, column_id = 4)
    (100, 2, 'Implement product catalog', 'Create database schema and API for product listings', 4, 0, 'dev@example.com', '{"priority": "high"}'),
    (101, 2, 'Design shopping cart UI', 'Wireframes and prototypes for cart interface', 4, 1, 'designer@example.com', '{"priority": "high"}'),
    (102, 2, 'Set up payment gateway', 'Integrate Stripe for payment processing', 4, 2, 'dev@example.com', '{"priority": "critical"}'),
    (103, 2, 'Build user registration', 'Email/password signup with verification', 4, 3, 'dev@example.com', '{"priority": "high"}'),
    (104, 2, 'Create admin dashboard', 'Backend admin panel for managing products', 4, 4, 'dev@example.com', '{"priority": "medium"}'),
    (105, 2, 'Implement search functionality', 'Full-text search for products with filters', 4, 5, 'dev@example.com', '{"priority": "high"}'),
    (106, 2, 'Add product reviews', 'Rating and review system for products', 4, 6, 'dev@example.com', '{"priority": "medium"}'),
    (107, 2, 'Design checkout flow', 'Multi-step checkout process wireframes', 4, 7, 'designer@example.com', '{"priority": "high"}'),
    (108, 2, 'Implement inventory management', 'Track stock levels and low stock alerts', 4, 8, 'dev@example.com', '{"priority": "high"}'),
    (109, 2, 'Create email templates', 'Order confirmation and shipping notification emails', 4, 9, 'designer@example.com', '{"priority": "medium"}'),
    (110, 2, 'Set up analytics tracking', 'Google Analytics and custom event tracking', 4, 10, 'dev@example.com', '{"priority": "low"}'),
    (111, 2, 'Build wishlist feature', 'Allow users to save products for later', 4, 11, 'dev@example.com', '{"priority": "low"}'),
    (112, 2, 'Add coupon system', 'Discount codes and promotional campaigns', 4, 12, 'dev@example.com', '{"priority": "medium"}'),
    (113, 2, 'Implement order tracking', 'Real-time shipping status updates', 4, 13, 'dev@example.com', '{"priority": "medium"}'),
    (114, 2, 'Create mobile app design', 'iOS and Android app mockups', 4, 14, 'designer@example.com', '{"priority": "low"}'),
    (115, 2, 'Set up CDN for images', 'CloudFront or similar for product images', 4, 15, 'devops@example.com', '{"priority": "medium"}'),
    (116, 2, 'Implement social login', 'Google and Facebook OAuth integration', 4, 16, 'dev@example.com', '{"priority": "low"}'),
    (117, 2, 'Add product recommendations', 'ML-based product suggestion engine', 4, 17, 'dev@example.com', '{"priority": "low"}'),
    (118, 2, 'Create return/refund system', 'Process for handling returns and refunds', 4, 18, 'dev@example.com', '{"priority": "medium"}'),
    (119, 2, 'Build affiliate program', 'Referral tracking and commission system', 4, 19, 'dev@example.com', '{"priority": "low"}'),
    (120, 2, 'Implement live chat support', 'Customer service chat widget', 4, 20, 'dev@example.com', '{"priority": "medium"}'),
    (121, 2, 'Add multi-currency support', 'Display prices in different currencies', 4, 21, 'dev@example.com', '{"priority": "medium"}'),
    (122, 2, 'Create loyalty program', 'Points system for repeat customers', 4, 22, 'dev@example.com', '{"priority": "low"}'),
    (123, 2, 'Set up A/B testing', 'Framework for testing conversion improvements', 4, 23, 'dev@example.com', '{"priority": "low"}'),
    (124, 2, 'Implement gift cards', 'Purchase and redeem digital gift cards', 4, 24, 'dev@example.com', '{"priority": "low"}'),
    (125, 2, 'Add subscription products', 'Recurring billing for subscription items', 4, 25, 'dev@example.com', '{"priority": "medium"}'),
    (126, 2, 'Create blog system', 'Content marketing blog with CMS', 4, 26, 'dev@example.com', '{"priority": "low"}'),
    (127, 2, 'Implement tax calculations', 'Automatic tax calculation by location', 4, 27, 'dev@example.com', '{"priority": "high"}'),
    (128, 2, 'Add shipping calculator', 'Real-time shipping cost estimation', 4, 28, 'dev@example.com', '{"priority": "high"}'),
    (129, 2, 'Create product comparison', 'Side-by-side product comparison tool', 4, 29, 'dev@example.com', '{"priority": "low"}'),
    (130, 2, 'Set up monitoring alerts', 'Error tracking and performance monitoring', 4, 30, 'devops@example.com', '{"priority": "high"}'),
    (131, 2, 'Implement SEO optimization', 'Meta tags, sitemaps, and structured data', 4, 31, 'dev@example.com', '{"priority": "medium"}'),
    (132, 2, 'Add bulk order functionality', 'Wholesale ordering for business customers', 4, 32, 'dev@example.com', '{"priority": "low"}'),
    (133, 2, 'Create product bundles', 'Package deals with discounted pricing', 4, 33, 'dev@example.com', '{"priority": "low"}'),
    (134, 2, 'Implement back-in-stock alerts', 'Email notifications when items restock', 4, 34, 'dev@example.com', '{"priority": "low"}'),
    (135, 2, 'Add size/color variants', 'Product variations with separate SKUs', 4, 35, 'dev@example.com', '{"priority": "high"}'),
    (136, 2, 'Create vendor marketplace', 'Multi-vendor platform with commission splits', 4, 36, 'dev@example.com', '{"priority": "low"}'),
    (137, 2, 'Set up backup strategy', 'Automated database backups and recovery', 4, 37, 'devops@example.com', '{"priority": "critical"}'),
    (138, 2, 'Implement GDPR compliance', 'Cookie consent and data export features', 4, 38, 'dev@example.com', '{"priority": "high"}'),
    (139, 2, 'Add accessibility features', 'WCAG 2.1 AA compliance improvements', 4, 39, 'dev@example.com', '{"priority": "medium"}'),

    -- Active (35 tasks, column_id = 5)
    (140, 2, 'Build REST API endpoints', 'Core API for products, orders, and users', 5, 0, 'dev@example.com', '{"priority": "critical"}'),
    (141, 2, 'Design homepage layout', 'Hero section and featured products', 5, 1, 'designer@example.com', '{"priority": "high"}'),
    (142, 2, 'Configure database indexes', 'Optimize query performance for catalog', 5, 2, 'dev@example.com', '{"priority": "high"}'),
    (143, 2, 'Implement cart persistence', 'Save cart across sessions with localStorage', 5, 3, 'dev@example.com', '{"priority": "high"}'),
    (144, 2, 'Create product detail page', 'Images, description, and add-to-cart button', 5, 4, 'dev@example.com', '{"priority": "high"}'),
    (145, 2, 'Set up authentication flow', 'Login, logout, and session management', 5, 5, 'dev@example.com', '{"priority": "critical"}'),
    (146, 2, 'Design mobile responsive layout', 'Ensure all pages work on mobile devices', 5, 6, 'designer@example.com', '{"priority": "high"}'),
    (147, 2, 'Implement form validation', 'Client and server-side validation for all forms', 5, 7, 'dev@example.com', '{"priority": "high"}'),
    (148, 2, 'Add image upload system', 'Product image upload with compression', 5, 8, 'dev@example.com', '{"priority": "medium"}'),
    (149, 2, 'Create navigation menu', 'Category navigation and mega menu', 5, 9, 'dev@example.com', '{"priority": "high"}'),
    (150, 2, 'Implement pagination', 'Product listing pagination and infinite scroll', 5, 10, 'dev@example.com', '{"priority": "medium"}'),
    (151, 2, 'Add filtering system', 'Price, category, and attribute filters', 5, 11, 'dev@example.com', '{"priority": "high"}'),
    (152, 2, 'Create order confirmation page', 'Thank you page with order summary', 5, 12, 'dev@example.com', '{"priority": "high"}'),
    (153, 2, 'Set up email service', 'SendGrid or similar for transactional emails', 5, 13, 'dev@example.com', '{"priority": "high"}'),
    (154, 2, 'Implement rate limiting', 'Protect API from abuse and DDoS', 5, 14, 'dev@example.com', '{"priority": "high"}'),
    (155, 2, 'Add error handling', 'Global error pages and logging', 5, 15, 'dev@example.com', '{"priority": "high"}'),
    (156, 2, 'Create user profile page', 'View and edit account information', 5, 16, 'dev@example.com', '{"priority": "medium"}'),
    (157, 2, 'Implement order history', 'List of past orders with details', 5, 17, 'dev@example.com', '{"priority": "medium"}'),
    (158, 2, 'Add loading states', 'Skeleton screens and spinners', 5, 18, 'designer@example.com', '{"priority": "medium"}'),
    (159, 2, 'Create footer component', 'Links, social media, and newsletter signup', 5, 19, 'dev@example.com', '{"priority": "low"}'),
    (160, 2, 'Set up staging environment', 'Separate staging server for testing', 5, 20, 'devops@example.com', '{"priority": "high"}'),
    (161, 2, 'Implement breadcrumbs', 'Navigation breadcrumbs for all pages', 5, 21, 'dev@example.com', '{"priority": "low"}'),
    (162, 2, 'Add sort functionality', 'Sort products by price, name, popularity', 5, 22, 'dev@example.com', '{"priority": "medium"}'),
    (163, 2, 'Create FAQ page', 'Common questions and answers', 5, 23, 'content@example.com', '{"priority": "low"}'),
    (164, 2, 'Implement guest checkout', 'Allow purchases without registration', 5, 24, 'dev@example.com', '{"priority": "high"}'),
    (165, 2, 'Add quick view modal', 'Product preview without leaving catalog', 5, 25, 'dev@example.com', '{"priority": "low"}'),
    (166, 2, 'Create terms of service', 'Legal terms and conditions', 5, 26, 'legal@example.com', '{"priority": "medium"}'),
    (167, 2, 'Implement privacy policy', 'Data collection and usage policy', 5, 27, 'legal@example.com', '{"priority": "high"}'),
    (168, 2, 'Add recently viewed items', 'Show products user has browsed', 5, 28, 'dev@example.com', '{"priority": "low"}'),
    (169, 2, 'Create contact form', 'Customer inquiry form with email notification', 5, 29, 'dev@example.com', '{"priority": "medium"}'),
    (170, 2, 'Set up SSL certificate', 'HTTPS for secure transactions', 5, 30, 'devops@example.com', '{"priority": "critical"}'),
    (171, 2, 'Implement address book', 'Save multiple shipping addresses', 5, 31, 'dev@example.com', '{"priority": "medium"}'),
    (172, 2, 'Add password reset', 'Forgot password email flow', 5, 32, 'dev@example.com', '{"priority": "high"}'),
    (173, 2, 'Create about us page', 'Company story and team information', 5, 33, 'content@example.com', '{"priority": "low"}'),
    (174, 2, 'Implement notification system', 'Toast messages for user actions', 5, 34, 'dev@example.com', '{"priority": "medium"}'),

    -- Completed (25 tasks, column_id = 6)
    (175, 2, 'Set up project repository', 'Initialize Git repo with README', 6, 0, 'dev@example.com', '{"priority": "critical"}'),
    (176, 2, 'Choose tech stack', 'Finalize framework and database choices', 6, 1, 'architect@example.com', '{"priority": "critical"}'),
    (177, 2, 'Create project wireframes', 'Low-fidelity sketches of all pages', 6, 2, 'designer@example.com', '{"priority": "high"}'),
    (178, 2, 'Define database schema', 'ERD for all tables and relationships', 6, 3, 'architect@example.com', '{"priority": "critical"}'),
    (179, 2, 'Set up development environment', 'Docker compose for local development', 6, 4, 'devops@example.com', '{"priority": "high"}'),
    (180, 2, 'Install dependencies', 'All npm/pip packages and tools', 6, 5, 'dev@example.com', '{"priority": "high"}'),
    (181, 2, 'Configure linting tools', 'ESLint and Prettier setup', 6, 6, 'dev@example.com', '{"priority": "medium"}'),
    (182, 2, 'Create design system', 'Color palette, typography, components', 6, 7, 'designer@example.com', '{"priority": "high"}'),
    (183, 2, 'Set up CI/CD pipeline', 'GitHub Actions for automated deployment', 6, 8, 'devops@example.com', '{"priority": "high"}'),
    (184, 2, 'Write project documentation', 'Architecture decisions and setup guide', 6, 9, 'architect@example.com', '{"priority": "medium"}'),
    (185, 2, 'Configure environment variables', 'Separate configs for dev/staging/prod', 6, 10, 'devops@example.com', '{"priority": "high"}'),
    (186, 2, 'Set up error logging', 'Sentry or similar for error tracking', 6, 11, 'dev@example.com', '{"priority": "high"}'),
    (187, 2, 'Create brand logo', 'Company logo and favicon', 6, 12, 'designer@example.com', '{"priority": "high"}'),
    (188, 2, 'Define API conventions', 'RESTful standards and naming conventions', 6, 13, 'architect@example.com', '{"priority": "medium"}'),
    (189, 2, 'Set up test framework', 'Jest and testing library configuration', 6, 14, 'dev@example.com', '{"priority": "high"}'),
    (190, 2, 'Create mockup designs', 'High-fidelity designs for all pages', 6, 15, 'designer@example.com', '{"priority": "high"}'),
    (191, 2, 'Configure database migrations', 'Migration tool and initial schema', 6, 16, 'dev@example.com', '{"priority": "high"}'),
    (192, 2, 'Set up code review process', 'PR templates and review guidelines', 6, 17, 'manager@example.com', '{"priority": "medium"}'),
    (193, 2, 'Define git workflow', 'Branching strategy and commit conventions', 6, 18, 'manager@example.com', '{"priority": "medium"}'),
    (194, 2, 'Create component library', 'Reusable UI components with Storybook', 6, 19, 'dev@example.com', '{"priority": "medium"}'),
    (195, 2, 'Set up performance monitoring', 'Lighthouse CI and performance budgets', 6, 20, 'devops@example.com', '{"priority": "medium"}'),
    (196, 2, 'Write security guidelines', 'Security best practices for the team', 6, 21, 'security@example.com', '{"priority": "high"}'),
    (197, 2, 'Configure CORS settings', 'Cross-origin resource sharing rules', 6, 22, 'dev@example.com', '{"priority": "high"}'),
    (198, 2, 'Set up domain and DNS', 'Purchase domain and configure nameservers', 6, 23, 'devops@example.com', '{"priority": "high"}'),
    (199, 2, 'Create sprint planning template', 'Agile workflow and sprint structure', 6, 24, 'manager@example.com', '{"priority": "low"}')
ON CONFLICT (id) DO NOTHING;

-- Sample task assignments (assign admin user to some tasks)
INSERT INTO task_assignees (task_id, user_id)
VALUES
    -- Assign to tasks in project 1
    (1, 1),  -- CI pipeline
    (3, 1),  -- User auth
    -- Assign to several tasks in project 2
    (100, 1), -- Product catalog
    (102, 1), -- Payment gateway
    (105, 1), -- Search functionality
    (140, 1), -- REST API
    (145, 1), -- Authentication flow
    (175, 1), -- Set up repo
    (176, 1)  -- Choose tech stack
ON CONFLICT (task_id, user_id) DO NOTHING;

-- Task labels for project 2 (E-Commerce Platform)
-- Backlog tasks labels (IDs 100-139)
INSERT INTO task_labels (task_id, label_id) VALUES
    (100, 11), (100, 16), (100, 10),
    (101, 4), (101, 2), (101, 12),
    (102, 9), (102, 10), (102, 16),
    (103, 11), (103, 16),
    (104, 11), (104, 16),
    (105, 11), (105, 16),
    (106, 11), (105, 15),
    (107, 4), (107, 12), (107, 2),
    (108, 11), (108, 16), (108, 10),
    (109, 4), (109, 2),
    (110, 5), (110, 13),
    (111, 11), (101, 15),
    (112, 11),
    (113, 11), (112, 16),
    (114, 4), (105, 12),
    (115, 5), (106, 13),
    (116, 11), (115, 16),
    (117, 11),
    (118, 11), (117, 16),
    (119, 11),
    (120, 11), (119, 15),
    (121, 11), (120, 15),
    (122, 11),
    (123, 5), (123, 13),
    (124, 11),
    (125, 11),
    (126, 4), (125, 2),
    (127, 9), (127, 10), (127, 16),
    (128, 11), (128, 16),
    (129, 11), (128, 15),
    (130, 5), (129, 13), (130, 9),
    (131, 11), (130, 15),
    (132, 11),
    (133, 11),
    (134, 11),
    (135, 9), (135, 10), (135, 16),
    (136, 11),
    (137, 5), (136, 13), (137, 9),
    (138, 9), (138, 8), (138, 16),
    (139, 11), (138, 15)
ON CONFLICT (task_id, label_id) DO NOTHING;

-- Active tasks labels (IDs 140-174)
INSERT INTO task_labels (task_id, label_id) VALUES
    (140, 9), (140, 10), (140, 11), (140, 16),
    (141, 4), (141, 2), (141, 12), (141, 15),
    (142, 11), (142, 16),
    (143, 11), (143, 15),
    (144, 11), (144, 15),
    (145, 9), (145, 10), (145, 11), (145, 16),
    (146, 4), (146, 2), (146, 12), (146, 15),
    (147, 11),
    (148, 11), (148, 16),
    (149, 11), (149, 15),
    (150, 11), (150, 15),
    (151, 9), (151, 10), (151, 11), (151, 15),
    (152, 11), (152, 15),
    (153, 5), (154, 13),
    (154, 9), (154, 8), (154, 16),
    (155, 11),
    (156, 11), (156, 15),
    (157, 11), (157, 15),
    (158, 4), (158, 2), (158, 12),
    (159, 11), (158, 15),
    (160, 5), (159, 13), (160, 10),
    (161, 11), (161, 15),
    (162, 11), (162, 15),
    (163, 4), (162, 2),
    (164, 9), (164, 10), (164, 11), (164, 16),
    (165, 11), (165, 15),
    (166, 4), (165, 2),
    (167, 9), (167, 10), (167, 8),
    (168, 11), (168, 15),
    (169, 11),
    (170, 9), (169, 10), (169, 5), (170, 13),
    (171, 11), (171, 16),
    (172, 9), (172, 10), (172, 8), (172, 16),
    (173, 4), (172, 2),
    (174, 11), (174, 15)
ON CONFLICT (task_id, label_id) DO NOTHING;

-- Completed tasks labels (IDs 175-199)
INSERT INTO task_labels (task_id, label_id) VALUES
    (175, 9), (175, 10), (175, 5), (175, 13), (175, 11), (175, 16),
    (176, 9), (176, 10), (176, 11), (176, 16),
    (177, 10), (177, 4), (177, 2),
    (178, 9), (178, 10), (178, 11), (178, 16),
    (179, 9), (179, 10), (179, 5), (179, 13),
    (180, 10), (180, 11),
    (181, 11),
    (182, 9), (182, 10), (182, 4), (182, 2), (182, 12), (183, 15),
    (183, 9), (183, 10), (183, 5), (183, 11), (183, 13),
    (184, 11), (184, 2),
    (185, 9), (185, 10), (185, 5), (185, 11), (185, 13),
    (186, 9), (186, 10), (186, 5), (186, 11), (186, 16),
    (187, 9), (187, 10), (187, 4), (187, 12),
    (188, 10), (188, 11), (188, 16),
    (189, 9), (189, 10), (189, 5), (189, 11),
    (190, 9), (190, 10), (190, 4), (190, 2), (190, 12), (191, 15),
    (191, 9), (191, 10), (191, 11), (191, 16),
    (192, 11), (192, 2),
    (193, 11), (193, 2),
    (194, 11), (195, 15),
    (195, 10), (195, 5), (195, 11), (195, 13),
    (196, 9), (196, 10), (196, 8), (196, 2),
    (197, 9), (197, 10), (197, 8), (197, 11), (197, 16),
    (198, 9), (198, 10), (198, 5), (198, 11), (198, 13),
    (199, 11), (199, 2)
ON CONFLICT (task_id, label_id) DO NOTHING;

-- Reset sequences to avoid conflicts with future inserts.
SELECT setval('users_id_seq', (SELECT COALESCE(MAX(id), 0) FROM users));
SELECT setval('organizations_id_seq', (SELECT COALESCE(MAX(id), 0) FROM organizations));
SELECT setval('projects_id_seq', (SELECT COALESCE(MAX(id), 0) FROM projects));
SELECT setval('project_columns_id_seq', (SELECT COALESCE(MAX(id), 0) FROM project_columns));
SELECT setval('project_labels_id_seq', (SELECT COALESCE(MAX(id), 0) FROM project_labels));
SELECT setval('tasks_id_seq', (SELECT COALESCE(MAX(id), 1) FROM tasks));
