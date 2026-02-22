# pm - Project Management for Teams

A modern, open-source kanban board for project management. Built with Go, htmx, and PostgreSQL, pm provides organizations with a simple yet powerful way to track work, collaborate on tasks, and manage projects.

## Features

### 📋 Kanban Workflow
- **Drag-and-drop** task management between customizable columns
- **Customizable columns** with color coding (default: To Do, In Progress, Done)
- **Task cards** showing title, description, due date, and assignees
- Task counts per column for quick progress tracking

### 👥 Multi-Tenant Organizations
- **Organization-based workspaces** with member management
- **Multiple projects per organization** for different teams or initiatives
- **Cross-project task dependencies** for complex workflows

### 🔗 Task Dependencies
- **Blocking relationships** between tasks (Task B blocked by Task A)
- **Dual view** of blockers (what's blocking this task) and blocking (what this task blocks)
- **Cycle detection** prevents circular dependencies
- Visual indicators for dependency status

### 👤 Team Collaboration
- **User accounts** with email authentication and session management
- **Task assignments** to multiple team members
- **Threaded comments** with nested replies (up to 3 levels deep)
- **Activity feed** tracks all task changes with audit history

### ⚙️ Flexible Task Properties
- **Arbitrary key-value metadata** stored as JSONB for custom fields
- **Due dates** with visual indicators for overdue and due-soon tasks
- **Rich descriptions** and editing capabilities
- **Author tracking** for task ownership

### 🔍 Search & Discovery
- **Task search** across titles and descriptions
- **User search** for finding team members
- **Project picker** for quick navigation between projects
- **URL-based navigation** for bookmarking and sharing task views

### 🎨 Modern UX
- **htmx-powered partial updates** for instant feedback
- **Keyboard shortcuts** (ESC to close panels)
- **Responsive design** works on desktop and tablet
- **Clean, accessible interface** built with TailwindCSS

## Tech Stack

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.25 |
| **Web Framework** | Fiber v3 |
| **Database** | PostgreSQL with sqlx |
| **Migrations** | goose |
| **Templates** | templ |
| **CSS** | TailwindCSS v4 |
| **Frontend Interactivity** | htmx 2.x |

## Quick Start

```bash
# Install dependencies
mise install
mise run setup

# Start development server
mise run dev

# (Optional) In a separate terminal, run CSS watch mode
mise run dev:css
```

Visit `http://localhost:3000` to get started.

## Organization Use Cases

### Software Development Teams
- **Sprint planning**: Create columns for Backlog, In Progress, In Review, Done
- **Bug tracking**: Use dependencies to link bugs to the features they block
- **Code reviews**: Assign tasks to multiple reviewers for sign-off
- **Release milestones**: Tag tasks with metadata like `milestone: v1.2`

### Marketing Teams
- **Campaign tracking**: Columns for Ideation, Content Creation, Review, Published
- **Content pipeline**: Chain dependencies (brief → draft → design → publish)
- **Creator assignments**: Assign writers, designers, and reviewers to campaigns
- **Due date management**: Track publication deadlines with visual urgency indicators

### Customer Support
- **Ticket management**: Columns for New, In Progress, Awaiting Customer, Resolved
- **Priority routing**: Use metadata for `priority: high`, `SLA: 4h`, `escalated: true`
- **Cross-team dependencies**: Block engineering tasks on customer features
- **Activity tracking**: Full audit log of who did what and when

### Operations & Infrastructure
- **Maintenance schedules**: Tasks for recurring operations with due dates
- **Incident response**: Track incident tasks blocking normal operations
- **Change management**: Dependency chains for multi-step deployments
- **On-call assignments**: Tag tasks with metadata like `team: DB`, `type: alert`

### Consultancies & Agencies
- **Client projects**: Separate projects per client within your organization
- **Deliverable tracking**: Custom metadata for `client: Acme`, `contract: $10k`, `hours: 8`
- **Resource assignments**: Assign multiple consultants to high-touch client work
- **Cross-client dependencies**: Block internal tooling tasks on client projects

### Non-Profit Organizations
- **Volunteer coordination**: Assign tasks to volunteer team members
- **Event planning**: Columns for Planning, Recruitment, Execution, Follow-up
- **Grant management**: Track deliverables with metadata for `grant: 2025-EF`, `deadline: Q2`
- **Program dependencies**: Chain tasks across different program areas

### HR & People Operations
- **Onboarding flows**: Dependency chain from offer accepted → account setup → orientation
- **Employee lifecycle**: Custom columns per stage (Hiring, Onboarding, Active, Offboarding)
- **Policy updates**: Assign tasks to document owners and reviewers
- **Metadata tags**: Track `department`, `urgent`, `quarter`

## Example Workflows

### Starting a New Project
1. Create an organization (or join an existing one)
2. Create a new project with name and description
3. Customize columns to match your workflow (or use defaults)
4. Create tasks directly in columns via the "Add Task" button

### Managing a Task with Dependencies
1. Create or open a task to see the full detail pane
2. In the "Dependencies" section, search for and add blocking tasks
3. Visual indicators show which tasks are blocked and which are blocking
4. The system prevents circular dependencies

### Collaborating with Comments
1. Open any task to see its history
2. Add comments with rich text descriptions
3. Reply to comments for threaded discussions
4. Edit or delete your own comments
5. View the activity feed to see all task changes

### Using Custom Metadata
1. Open a task and scroll to the "Metadata" section
2. Click "Add Metadata" to create key-value pairs
3. Examples: `priority: high`, `story_points: 5`, `estimate: 4h`
4. Search, edit, or delete metadata entries as needed

## Contributing

We welcome contributions! Please see the [AGENTS.md](./AGENTS.md) file for developer guidelines, including our tech stack, build commands, and code style conventions.

## Development

```bash
# Run tests
mise run test

# Run tests with coverage
mise run test:coverage

# Lint code
mise run lint

# Auto-fix linting issues
mise run lint:fix

# Database migrations
mise run db:up       # Apply pending migrations
mise run db:down     # Rollback last migration
mise run db:status   # Show migration status
```

## License

MIT
