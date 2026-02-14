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
