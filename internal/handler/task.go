package handler

import (
	"fmt"
	"strconv"
	"time"

	"pm/internal/middleware"
	"pm/internal/model"
	"pm/internal/store"
	"pm/views"
	"pm/views/components"

	"github.com/gofiber/fiber/v3"
)

// parseDueDate parses a date string in YYYY-MM-DD format.
func parseDueDate(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	return &parsed, nil
}

// parseOrgAndProject extracts org_id and project_id from route params.
func parseOrgAndProject(c fiber.Ctx) (orgID, projectID int64, err error) {
	orgID, err = strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return 0, 0, fiber.NewError(fiber.StatusBadRequest, "invalid org_id")
	}

	projectID, err = strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return 0, 0, fiber.NewError(fiber.StatusBadRequest, "invalid project_id")
	}

	return orgID, projectID, nil
}

// HandleMoveTask updates a task's column and position when it is dragged.
// PATCH /orgs/:org_id/projects/:project_id/tasks/:id/move
func (h *Handler) HandleMoveTask(c fiber.Ctx) error {
	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	newColumnID, err := strconv.ParseInt(c.FormValue("columnID"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "columnID is required")
	}

	newPosition, err := strconv.Atoi(c.FormValue("position"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid position")
	}

	// Get current task data
	currentTask, err := h.store.GetTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "task not found")
	}

	// Get column names for activity logging
	oldColName := ""
	newColName := ""

	if currentTask.ColumnID != newColumnID {
		oldCol, err := h.store.GetProjectColumn(c.Context(), currentTask.ColumnID)
		if err == nil {
			oldColName = oldCol.Name
		}
		newCol, err := h.store.GetProjectColumn(c.Context(), newColumnID)
		if err == nil {
			newColName = newCol.Name
		}
	}

	if err := h.store.MoveTask(c.Context(), taskID, newColumnID, newPosition); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to move task")
	}

	// Create activity record for column move
	user, err := middleware.GetCurrentUser(c)
	if err == nil && oldColName != "" && newColName != "" {
		_ = h.store.CreateActivity(c.Context(), taskID, user.ID, "move", "column_id", oldColName, newColName)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// HandleNewTaskForm returns the add task form HTML.
// GET /orgs/:org_id/projects/:project_id/tasks/new-form?columnID=...
func (h *Handler) HandleNewTaskForm(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	columnID, err := strconv.ParseInt(c.Query("columnID"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "columnID is required")
	}

	return render(c, components.AddTaskForm(columnID, orgID, projectID))
}

// HandleCancelForm returns the "Add task" button HTML.
// GET /orgs/:org_id/projects/:project_id/tasks/cancel-form?columnID=...
func (h *Handler) HandleCancelForm(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	columnID, err := strconv.ParseInt(c.Query("columnID"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "columnID is required")
	}

	return render(c, components.AddTaskButton(columnID, orgID, projectID))
}

// HandleCreateTask creates a new task and returns the new card HTML + OOB updates.
// POST /orgs/:org_id/projects/:project_id/tasks
func (h *Handler) HandleCreateTask(c fiber.Ctx) error {
	orgID, _, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	projectID, err := strconv.ParseInt(c.FormValue("project_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project_id")
	}

	title := c.FormValue("title")
	if title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title is required")
	}

	description := c.FormValue("description")
	columnID, err := strconv.ParseInt(c.FormValue("columnID"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "columnID is required")
	}

	dueDate, err := parseDueDate(c.FormValue("due_date"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	task, err := h.store.CreateTask(c.Context(), projectID, title, description, columnID, dueDate)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create task")
	}

	// Create activity record for task creation
	user, err := middleware.GetCurrentUser(c)
	if err == nil {
		_ = h.store.CreateActivity(c.Context(), task.ID, user.ID, "create", "", nil, nil)
	}

	// Get updated count.
	count, err := h.store.CountTasksByColumn(c.Context(), projectID, columnID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task count")
	}

	return render(c, views.NewTaskResponse(task, orgID, columnID, count))
}

// HandleDeleteTask deletes a task and returns OOB count update.
// DELETE /orgs/:org_id/projects/:project_id/tasks/:id
func (h *Handler) HandleDeleteTask(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	// Get task info before deleting (for column ID).
	task, err := h.store.GetTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "task not found")
	}

	err = h.store.DeleteTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete task")
	}

	// Get updated count.
	count, err := h.store.CountTasksByColumn(c.Context(), projectID, task.ColumnID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task count")
	}

	_ = orgID // orgID validated by middleware; not needed for delete response.
	return render(c, views.DeleteTaskResponse(task.ColumnID, count))
}

// HandleTaskDetail returns the detail pane for a task.
// GET /orgs/:org_id/projects/:project_id/tasks/:id/detail
func (h *Handler) HandleTaskDetail(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Redirect().To("/login")
	}

	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "task not found")
	}

	// Load comments
	comments, err := h.store.ListTaskComments(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load comments")
	}

	// Build comment tree with max depth 3
	commentTree := store.BuildCommentTree(comments, 3)
	// Load activity feed
	activity, err := h.store.GetTaskActivity(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load activity")
	}


	return render(c, views.TaskDetailPane(*taskWithDeps, orgID, projectID, commentTree, user, activity))
}

// HandleUpdateTask updates a task's basic fields.
// PATCH /orgs/:org_id/projects/:project_id/tasks/:id
func (h *Handler) HandleUpdateTask(c fiber.Ctx) error {
	orgID, _, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	// Get current task data for comparison
	currentTask, err := h.store.GetTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "task not found")
	}

	title := c.FormValue("title")
	description := c.FormValue("description")
	author := c.FormValue("author")

	if title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title is required")
	}

	dueDate, err := parseDueDate(c.FormValue("due_date"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	err = h.store.UpdateTask(c.Context(), taskID, title, description, author, dueDate)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update task")
	}

	// Create activity records for field changes
	user, err := middleware.GetCurrentUser(c)
	if err == nil {
		if currentTask.Title != title {
			_ = h.store.CreateActivity(c.Context(), taskID, user.ID, "update", "title", currentTask.Title, title)
		}
		if currentTask.Description != description {
			_ = h.store.CreateActivity(c.Context(), taskID, user.ID, "update", "description", currentTask.Description, description)
		}
		if currentTask.Author != author {
			_ = h.store.CreateActivity(c.Context(), taskID, user.ID, "update", "author", currentTask.Author, author)
		}
		if currentTask.DueDateString() != dueDate.Format("2006-01-02") {
			_ = h.store.CreateActivity(c.Context(), taskID, user.ID, "update", "due_date", currentTask.DueDateString(), dueDate.Format("2006-01-02"))
		}
	}

	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	return render(c, views.TaskFieldUpdateResponse(*taskWithDeps, orgID))
}

// HandleAddDependency adds a dependency to a task.
// POST /orgs/:org_id/projects/:project_id/tasks/:id/dependencies
func (h *Handler) HandleAddDependency(c fiber.Ctx) error {
	orgID, _, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	dependsOnID, err := strconv.ParseInt(c.FormValue("depends_on_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid depends_on_id")
	}

	err = h.store.AddDependency(c.Context(), taskID, dependsOnID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to add dependency")
	}

	// Get the dependency task title for the activity
	depTask, err := h.store.GetTask(c.Context(), dependsOnID)
	if err != nil {
		depTask = &model.Task{Title: fmt.Sprintf("Task %d", dependsOnID)}
	}

	// Create activity record
	user, err := middleware.GetCurrentUser(c)
	if err == nil {
		_ = h.store.CreateActivity(c.Context(), taskID, user.ID, "update", "dependency", nil, map[string]string{"title": depTask.Title, "id": fmt.Sprintf("%d", dependsOnID)})
	}

	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	return render(c, views.DependencySection(*taskWithDeps, orgID))
}

// HandleRemoveDependency removes a dependency from a task.
// DELETE /orgs/:org_id/projects/:project_id/tasks/:id/dependencies/:depID
func (h *Handler) HandleRemoveDependency(c fiber.Ctx) error {
	orgID, _, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	dependsOnID, err := strconv.ParseInt(c.Params("depID"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid dependency id")
	}

	// Get the dependency task title for the activity before removing
	depTask, err := h.store.GetTask(c.Context(), dependsOnID)
	if err != nil {
		depTask = &model.Task{Title: fmt.Sprintf("Task %d", dependsOnID)}
	}

	err = h.store.RemoveDependency(c.Context(), taskID, dependsOnID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to remove dependency")
	}

	// Create activity record
	user, err := middleware.GetCurrentUser(c)
	if err == nil {
		_ = h.store.CreateActivity(c.Context(), taskID, user.ID, "update", "dependency", map[string]string{"title": depTask.Title, "id": fmt.Sprintf("%d", dependsOnID)}, nil)
	}

	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	return render(c, views.DependencySection(*taskWithDeps, orgID))
}

// HandleSearchTasks searches for tasks across the organization
// GET /orgs/:org_id/tasks/search?q=query&exclude=123
func (h *Handler) HandleSearchTasks(c fiber.Ctx) error {
	orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org id")
	}

	query := c.Query("q", "")
	excludeTaskID, _ := strconv.ParseInt(c.Query("exclude"), 10, 64)
	limit := 20

	var results []model.TaskSearchResult

	if len(query) < 2 {
		// Return recent tasks if no query or query too short
		results, err = h.store.GetRecentOrganizationTasks(c.Context(), orgID, excludeTaskID, limit)
	} else {
		// Search for tasks matching query
		results, err = h.store.SearchOrganizationTasks(c.Context(), orgID, query, excludeTaskID, limit)
	}

	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to search tasks")
	}

	return c.JSON(results)
}

// HandleCheckDependencyCycle checks if adding a dependency would create a circular dependency
// GET /orgs/:org_id/projects/:project_id/tasks/:id/dependencies/check?depends_on=456
func (h *Handler) HandleCheckDependencyCycle(c fiber.Ctx) error {
	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	dependsOnID, err := strconv.ParseInt(c.Query("depends_on"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid depends_on id")
	}

	wouldCycle, err := h.store.WouldCreateCycle(c.Context(), taskID, dependsOnID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to check for cycles")
	}

	return c.JSON(fiber.Map{
		"would_create_cycle": wouldCycle,
	})
}

// HandleAssignSelf assigns the current user to a task.
// POST /orgs/:org_id/projects/:project_id/tasks/:id/assign-self
func (h *Handler) HandleAssignSelf(c fiber.Ctx) error {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		return err
	}

	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	// Assign user to task
	err = h.store.AssignUserToTask(c.Context(), taskID, currentUser.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to assign user")
	}

	// Create activity record
	_ = h.store.CreateActivity(c.Context(), taskID, currentUser.ID, "assign", "", nil, map[string]string{"user": currentUser.DisplayName, "email": currentUser.Email})

	// Get updated task with dependencies (includes assignees)
	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task")
	}

	// Return the updated assignee section for htmx swap
	return render(c, components.AssigneeSection(taskWithDeps.ID, taskWithDeps.Assignees, orgID, projectID, *currentUser))
}

// HandleUnassign removes a user from a task.
// DELETE /orgs/:org_id/projects/:project_id/tasks/:id/assignees/:user_id
func (h *Handler) HandleUnassign(c fiber.Ctx) error {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		return err
	}

	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	userIDToRemove, err := strconv.ParseInt(c.Params("user_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user id")
	}

	// Permission check: users can unassign themselves, or org owners can unassign anyone
	if userIDToRemove != currentUser.ID {
		// Check if current user is org owner
		org, err := h.store.GetOrganization(c.Context(), orgID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to get organization")
		}
		if org.OwnerUserID != currentUser.ID {
			return fiber.NewError(fiber.StatusForbidden, "only org owners can unassign others")
		}
	}

	// Get user info before unassigning
	userToRemove, err := h.store.GetUser(c.Context(), userIDToRemove)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get user")
	}

	// Unassign user from task
	err = h.store.UnassignUserFromTask(c.Context(), taskID, userIDToRemove)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to unassign user")
	}

	// Create activity record
	_ = h.store.CreateActivity(c.Context(), taskID, currentUser.ID, "unassign", "", map[string]string{"user": userToRemove.DisplayName, "email": userToRemove.Email}, nil)

	// Get updated task with dependencies (includes assignees)
	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task")
	}

	// Return the updated assignee section for htmx swap
	return render(c, components.AssigneeSection(taskWithDeps.ID, taskWithDeps.Assignees, orgID, projectID, *currentUser))
}

// HandleSearchUsers searches for users in an organization
// GET /orgs/:org_id/users/search?q=query
func (h *Handler) HandleSearchUsers(c fiber.Ctx) error {
	orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org id")
	}

	query := c.Query("q", "")
	limit := 20

	// Search for users (include current user - they can see themselves in the list)
	// Pass 0 as excludeUserID to include everyone
	users, err := h.store.SearchOrganizationMembers(c.Context(), orgID, query, 0, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to search users")
	}

	// Convert to response format
	type userSearchResult struct {
		ID          int64  `json:"id"`
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
	}

	results := make([]userSearchResult, len(users))
	for i, u := range users {
		results[i] = userSearchResult{
			ID:          u.ID,
			Email:       u.Email,
			DisplayName: u.DisplayName,
		}
	}

	return c.JSON(results)
}

// HandleAssignUser assigns a user to a task.
// POST /orgs/:org_id/projects/:project_id/tasks/:id/assignees/:user_id
func (h *Handler) HandleAssignUser(c fiber.Ctx) error {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		return err
	}

	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	userIDToAssign, err := strconv.ParseInt(c.Params("user_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user id")
	}

	// Verify the user is a member of the organization
	isMember, err := h.store.IsMember(c.Context(), orgID, userIDToAssign)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to verify membership")
	}
	if !isMember {
		return fiber.NewError(fiber.StatusBadRequest, "user is not a member of this organization")
	}

	// Get user info before assigning
	userToAssign, err := h.store.GetUserByID(c.Context(), userIDToAssign)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get user")
	}

	// Assign user to task
	err = h.store.AssignUserToTask(c.Context(), taskID, userIDToAssign)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to assign user")
	}

	// Create activity record
	_ = h.store.CreateActivity(c.Context(), taskID, currentUser.ID, "assign", "", nil, map[string]string{"user": userToAssign.DisplayName, "email": userToAssign.Email})

	// Get updated task with dependencies (includes assignees)
	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task")
	}

	// Return the updated assignee section for htmx swap
	return render(c, components.AssigneeSection(taskWithDeps.ID, taskWithDeps.Assignees, orgID, projectID, *currentUser))
}

// HandleAddMetadata adds a metadata key-value pair.
// POST /orgs/:org_id/projects/:project_id/tasks/:id/metadata
func (h *Handler) HandleAddMetadata(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	key := c.FormValue("key")
	value := c.FormValue("value")

	if key == "" {
		return fiber.NewError(fiber.StatusBadRequest, "key is required")
	}

	err = h.store.SetMetadataKey(c.Context(), taskID, key, value)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to add metadata")
	}

	// Get current task metadata to get old value
	task, err := h.store.GetTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	// Get old value (should be empty before adding)
	oldVal := ""
	if task.Metadata != nil {
		if v, exists := task.Metadata[key]; exists {
			if vStr, ok := v.(string); ok {
				oldVal = vStr
			} else {
				oldVal = fmt.Sprintf("%v", v)
			}
		}
	}

	// Create activity record
	user, err := middleware.GetCurrentUser(c)
	if err == nil {
		_ = h.store.CreateActivity(c.Context(), taskID, user.ID, "update", fmt.Sprintf("metadata:%s", key), oldVal, value)
	}

	return render(c, views.MetadataSection(taskID, orgID, projectID, task.Metadata))
}

// HandleDeleteMetadata removes a metadata key.
// DELETE /orgs/:org_id/projects/:project_id/tasks/:id/metadata/:key
func (h *Handler) HandleDeleteMetadata(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	key := c.Params("key")
	if key == "" {
		return fiber.NewError(fiber.StatusBadRequest, "key is required")
	}

	err = h.store.DeleteMetadataKey(c.Context(), taskID, key)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete metadata")
	}

	// Create activity record
	user, err := middleware.GetCurrentUser(c)
	if err == nil {
		_ = h.store.CreateActivity(c.Context(), taskID, user.ID, "update", fmt.Sprintf("metadata:%s", key), "deleted", nil)
	}

	task, err := h.store.GetTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	return render(c, views.MetadataSection(taskID, orgID, projectID, task.Metadata))
}

// HandleUpdateMetadata updates an existing metadata key-value pair.
// PATCH /orgs/:org_id/projects/:project_id/tasks/:id/metadata/:oldKey
func (h *Handler) HandleUpdateMetadata(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	oldKey := c.Params("oldKey")
	newKey := c.FormValue("key")
	value := c.FormValue("value")

	if oldKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "old key is required")
	}
	if newKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "new key is required")
	}

	// If key changed, delete old key first.
	if oldKey != newKey {
		task, err := h.store.GetTask(c.Context(), taskID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to get task")
		}
		if _, exists := task.Metadata[newKey]; exists {
			return fiber.NewError(fiber.StatusBadRequest, "key already exists")
		}

		err = h.store.DeleteMetadataKey(c.Context(), taskID, oldKey)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to delete old key")
		}
	}

	err = h.store.SetMetadataKey(c.Context(), taskID, newKey, value)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update metadata")
	}

	// Get current task metadata to get old value
	task, err := h.store.GetTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	// Get old value
	oldVal := ""
	if task.Metadata != nil {
		if v, exists := task.Metadata[newKey]; exists {
			if vStr, ok := v.(string); ok {
				oldVal = vStr
			} else {
				oldVal = fmt.Sprintf("%v", v)
			}
		}
	}

	// Create activity record
	user, err := middleware.GetCurrentUser(c)
	if err == nil {
		_ = h.store.CreateActivity(c.Context(), taskID, user.ID, "update", fmt.Sprintf("metadata:%s", newKey), oldVal, value)
	}

	return render(c, views.MetadataSection(taskID, orgID, projectID, task.Metadata))
}

// HandleUpdateColumn updates a task's column and moves it to the end of the new column.
// PATCH /orgs/:org_id/projects/:project_id/tasks/:id/column
func (h *Handler) HandleUpdateColumn(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	newColumnID, err := strconv.ParseInt(c.FormValue("columnID"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "columnID is required")
	}

	// Validate column belongs to project
	task, err := h.store.GetTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "task not found")
	}

	valid, err := h.store.ValidateColumnOwnership(c.Context(), newColumnID, task.ProjectID)
	if err != nil || !valid {
		return fiber.NewError(fiber.StatusBadRequest, "invalid column")
	}

	// Update column and get the old column ID.
	oldColumnID, err := h.store.UpdateTaskColumn(c.Context(), taskID, newColumnID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update column")
	}

	// Get old and new column names for activity
	oldCol, err := h.store.GetProjectColumn(c.Context(), oldColumnID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get old column")
	}

	newCol, err := h.store.GetProjectColumn(c.Context(), newColumnID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get new column")
	}

	// Create activity record for move
	user, err := middleware.GetCurrentUser(c)
	if err == nil {
		_ = h.store.CreateActivity(c.Context(), taskID, user.ID, "move", "column_id", oldCol.Name, newCol.Name)
	}

	// Get the updated task with dependencies.
	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	// Get all tasks and columns for the project to update the kanban board.
	tasks, err := h.store.ListTasksByProject(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load tasks")
	}

	columns, err := h.store.GetProjectColumns(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load columns")
	}

	tasksByColumn := store.GroupTasksByColumn(tasks, columns)

	return render(c, views.ColumnUpdateResponse(*taskWithDeps, orgID, oldColumnID, tasksByColumn, columns))
}
