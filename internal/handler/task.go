package handler

import (
	"fmt"
	"strconv"
	"time"

	"pm/internal/middleware"
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

// HandleMoveTask updates a task's status and position when it is dragged.
// PATCH /orgs/:org_id/projects/:project_id/tasks/:id/move
func (h *Handler) HandleMoveTask(c fiber.Ctx) error {
	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	newStatus := c.FormValue("status")
	if newStatus == "" {
		return fiber.NewError(fiber.StatusBadRequest, "status is required")
	}

	newPosition, err := strconv.Atoi(c.FormValue("position"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid position")
	}

	if err := h.store.MoveTask(c.Context(), taskID, newStatus, newPosition); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to move task")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// HandleNewTaskForm returns the add task form HTML.
// GET /orgs/:org_id/projects/:project_id/tasks/new-form?status=...
func (h *Handler) HandleNewTaskForm(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	status := c.Query("status")
	if status == "" {
		return fiber.NewError(fiber.StatusBadRequest, "status is required")
	}

	return render(c, components.AddTaskForm(status, orgID, projectID))
}

// HandleCancelForm returns the "Add task" button HTML.
// GET /orgs/:org_id/projects/:project_id/tasks/cancel-form?status=...
func (h *Handler) HandleCancelForm(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	status := c.Query("status")
	if status == "" {
		return fiber.NewError(fiber.StatusBadRequest, "status is required")
	}

	return render(c, components.AddTaskButton(status, orgID, projectID))
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
	status := c.FormValue("status")
	if status == "" {
		return fiber.NewError(fiber.StatusBadRequest, "status is required")
	}

	dueDate, err := parseDueDate(c.FormValue("due_date"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	task, err := h.store.CreateTask(c.Context(), projectID, title, description, status, dueDate)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create task")
	}

	// Get updated count.
	count, err := h.store.CountTasksByStatus(c.Context(), projectID, status)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task count")
	}

	return render(c, views.NewTaskResponse(task, orgID, status, count))
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

	// Get task info before deleting (for status).
	task, err := h.store.GetTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "task not found")
	}

	err = h.store.DeleteTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete task")
	}

	// Get updated count.
	count, err := h.store.CountTasksByStatus(c.Context(), projectID, task.Status)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task count")
	}

	_ = orgID // orgID validated by middleware; not needed for delete response.
	return render(c, views.DeleteTaskResponse(task.Status, count))
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

	return render(c, views.TaskDetailPane(*taskWithDeps, orgID, projectID, commentTree, user))
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

	err = h.store.RemoveDependency(c.Context(), taskID, dependsOnID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to remove dependency")
	}

	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	return render(c, views.DependencySection(*taskWithDeps, orgID))
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

	task, err := h.store.GetTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
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

	task, err := h.store.GetTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	return render(c, views.MetadataSection(taskID, orgID, projectID, task.Metadata))
}

// HandleUpdateStatus updates a task's status and moves it to the end of the new column.
// PATCH /orgs/:org_id/projects/:project_id/tasks/:id/status
func (h *Handler) HandleUpdateStatus(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	newStatus := c.FormValue("status")
	if newStatus == "" {
		return fiber.NewError(fiber.StatusBadRequest, "status is required")
	}

	// Validate status.
	validStatuses := map[string]bool{
		"todo":        true,
		"in_progress": true,
		"done":        true,
	}
	if !validStatuses[newStatus] {
		return fiber.NewError(fiber.StatusBadRequest, "invalid status")
	}

	// Update status and get the old status.
	oldStatus, err := h.store.UpdateTaskStatus(c.Context(), taskID, newStatus)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update status")
	}

	// Get the updated task with dependencies.
	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	// Get all tasks for the project to update the kanban board.
	tasks, err := h.store.ListTasksByProject(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load tasks")
	}

	tasksByStatus := store.GroupTasksByStatus(tasks)

	return render(c, views.StatusUpdateResponse(*taskWithDeps, orgID, oldStatus, tasksByStatus))
}
