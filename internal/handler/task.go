package handler

import (
	"strconv"

	"cracked-pm/views"
	"cracked-pm/views/components"

	"github.com/gofiber/fiber/v3"
)

// HandleMoveTask updates a task's status and position when it is dragged.
// PATCH /tasks/:id/move
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
// GET /tasks/new-form?status=...
func (h *Handler) HandleNewTaskForm(c fiber.Ctx) error {
	status := c.Query("status")
	if status == "" {
		return fiber.NewError(fiber.StatusBadRequest, "status is required")
	}

	// For simplicity, we'll use project ID 1 (the default project)
	projectID := int64(1)

	return render(c, components.AddTaskForm(status, projectID))
}

// HandleCancelForm returns the "Add task" button HTML.
// GET /tasks/cancel-form?status=...
func (h *Handler) HandleCancelForm(c fiber.Ctx) error {
	status := c.Query("status")
	if status == "" {
		return fiber.NewError(fiber.StatusBadRequest, "status is required")
	}

	return render(c, components.AddTaskButton(status))
}

// HandleCreateTask creates a new task and returns the new card HTML + OOB updates.
// POST /tasks
func (h *Handler) HandleCreateTask(c fiber.Ctx) error {
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

	task, err := h.store.CreateTask(c.Context(), projectID, title, description, status)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create task")
	}

	// Get updated count
	count, err := h.store.CountTasksByStatus(c.Context(), projectID, status)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task count")
	}

	// Return new task card + OOB updates for count and form reset
	return render(c, views.NewTaskResponse(task, status, count))
}

// HandleDeleteTask deletes a task and returns OOB count update.
// DELETE /tasks/:id
func (h *Handler) HandleDeleteTask(c fiber.Ctx) error {
	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	// Get task info before deleting (for project ID and status)
	tasks, err := h.store.ListTasksByProject(c.Context(), 1) // Assuming project 1
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load tasks")
	}

	var taskStatus string
	var projectID int64 = 1
	for _, t := range tasks {
		if t.ID == taskID {
			taskStatus = t.Status
			projectID = t.ProjectID
			break
		}
	}

	if taskStatus == "" {
		return fiber.NewError(fiber.StatusNotFound, "task not found")
	}

	err = h.store.DeleteTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete task")
	}

	// Get updated count
	count, err := h.store.CountTasksByStatus(c.Context(), projectID, taskStatus)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task count")
	}

	// Return OOB count update
	return render(c, views.DeleteTaskResponse(taskStatus, count))
}

// HandleTaskDetail returns the detail pane for a task.
// GET /tasks/:id/detail
func (h *Handler) HandleTaskDetail(c fiber.Ctx) error {
	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "task not found")
	}

	return render(c, views.TaskDetailPane(*taskWithDeps))
}

// HandleUpdateTask updates a task's basic fields.
// PATCH /tasks/:id
func (h *Handler) HandleUpdateTask(c fiber.Ctx) error {
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

	err = h.store.UpdateTask(c.Context(), taskID, title, description, author)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update task")
	}

	// Return the updated task detail pane
	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	return render(c, views.TaskDetailPane(*taskWithDeps))
}

// HandleAddDependency adds a dependency to a task.
// POST /tasks/:id/dependencies
func (h *Handler) HandleAddDependency(c fiber.Ctx) error {
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

	// Return updated dependency section
	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	return render(c, views.DependencySection(*taskWithDeps))
}

// HandleRemoveDependency removes a dependency from a task.
// DELETE /tasks/:id/dependencies/:depID
func (h *Handler) HandleRemoveDependency(c fiber.Ctx) error {
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

	// Return updated dependency section
	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	return render(c, views.DependencySection(*taskWithDeps))
}

// HandleAddMetadata adds a metadata key-value pair.
// POST /tasks/:id/metadata
func (h *Handler) HandleAddMetadata(c fiber.Ctx) error {
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

	// Get updated task
	task, err := h.store.GetTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	return render(c, views.MetadataSection(taskID, task.Metadata))
}

// HandleDeleteMetadata removes a metadata key.
// DELETE /tasks/:id/metadata/:key
func (h *Handler) HandleDeleteMetadata(c fiber.Ctx) error {
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

	// Get updated task
	task, err := h.store.GetTask(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	return render(c, views.MetadataSection(taskID, task.Metadata))
}

// HandleUpdateStatus updates a task's status and moves it to the end of the new column.
// PATCH /tasks/:id/status
func (h *Handler) HandleUpdateStatus(c fiber.Ctx) error {
	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	newStatus := c.FormValue("status")
	if newStatus == "" {
		return fiber.NewError(fiber.StatusBadRequest, "status is required")
	}

	// Validate status
	validStatuses := map[string]bool{
		"todo":        true,
		"in_progress": true,
		"done":        true,
	}
	if !validStatuses[newStatus] {
		return fiber.NewError(fiber.StatusBadRequest, "invalid status")
	}

	err = h.store.UpdateTaskStatus(c.Context(), taskID, newStatus)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update status")
	}

	// Return the updated task detail pane
	taskWithDeps, err := h.store.GetTaskWithDependencies(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload task")
	}

	return render(c, views.TaskDetailPane(*taskWithDeps))
}
