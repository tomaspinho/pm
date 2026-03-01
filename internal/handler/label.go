package handler

import (
	"encoding/json"
	"fmt"
	"strconv"

	"pm/internal/middleware"
	"pm/views"
	"pm/views/components"

	"github.com/gofiber/fiber/v3"
)

// HandleShowSettings renders the combined project settings page (columns + labels tabs)
// GET /orgs/:org_id/projects/:project_id/settings
func (h *Handler) HandleShowSettings(c fiber.Ctx) error {
	orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org_id")
	}

	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project_id")
	}

	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Redirect().To("/login")
	}

	project, err := h.store.GetProject(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "project not found")
	}

	// Load columns
	columns, err := h.store.GetProjectColumns(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load columns")
	}

	// Get task counts for each column
	columnTaskCounts := make(map[int64]int)
	for _, col := range columns {
		count, err := h.store.CountColumnTasks(c.Context(), col.ID)
		if err == nil {
			columnTaskCounts[col.ID] = count
		}
	}

	// Load labels
	labels, err := h.store.GetProjectLabels(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load labels")
	}

	// Get task counts for each label
	labelTaskCounts := make(map[int64]int)
	for _, label := range labels {
		count, err := h.store.CountLabelTasks(c.Context(), label.ID)
		if err == nil {
			labelTaskCounts[label.ID] = count
		}
	}

	// Get user's organizations for nav
	orgs, err := h.store.GetUserOrganizations(c.Context(), user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load organizations")
	}

	nav := views.NavContext{
		User:             user,
		Orgs:             orgs,
		CurrentOrgID:     orgID,
		CurrentProjectID: projectID,
	}

	return render(c, views.SettingsPage(project, orgID, columns, labels, columnTaskCounts, labelTaskCounts, nav))
}

// HandleGetLabels returns all labels for a project as JSON
// GET /orgs/:org_id/projects/:project_id/labels
func (h *Handler) HandleGetLabels(c fiber.Ctx) error {
	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project_id")
	}

	labels, err := h.store.GetProjectLabels(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load labels")
	}

	return c.JSON(labels)
}

// HandleCreateLabel creates a new label for a project
// POST /orgs/:org_id/projects/:project_id/labels
func (h *Handler) HandleCreateLabel(c fiber.Ctx) error {
	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project_id")
	}

	name := c.FormValue("name")
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "label name is required")
	}

	if len(name) > 50 {
		return fiber.NewError(fiber.StatusBadRequest, "label name must be 50 characters or less")
	}

	color := c.FormValue("color")
	if color == "" {
		return fiber.NewError(fiber.StatusBadRequest, "label color is required")
	}

	// Check for duplicate name (case-insensitive)
	existingLabels, err := h.store.GetProjectLabels(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to validate label name")
	}

	for _, label := range existingLabels {
		if label.Name == name {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("label '%s' already exists", name))
		}
	}

	label, err := h.store.CreateProjectLabel(c.Context(), projectID, name, color)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create label")
	}

	return c.JSON(label)
}

// HandleUpdateLabel updates a label's name and/or color
// PATCH /orgs/:org_id/projects/:project_id/labels/:label_id
func (h *Handler) HandleUpdateLabel(c fiber.Ctx) error {
	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project_id")
	}

	labelID, err := strconv.ParseInt(c.Params("label_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid label_id")
	}

	name := c.FormValue("name")
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "label name is required")
	}

	if len(name) > 50 {
		return fiber.NewError(fiber.StatusBadRequest, "label name must be 50 characters or less")
	}

	color := c.FormValue("color")
	if color == "" {
		return fiber.NewError(fiber.StatusBadRequest, "label color is required")
	}

	// Validate ownership
	owns, err := h.store.ValidateLabelOwnership(c.Context(), labelID, projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to validate label ownership")
	}
	if !owns {
		return fiber.NewError(fiber.StatusNotFound, "label not found")
	}

	err = h.store.UpdateProjectLabel(c.Context(), labelID, name, color)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update label")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// HandleDeleteLabel deletes a label (soft delete)
// DELETE /orgs/:org_id/projects/:project_id/labels/:label_id
func (h *Handler) HandleDeleteLabel(c fiber.Ctx) error {
	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project_id")
	}

	labelID, err := strconv.ParseInt(c.Params("label_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid label_id")
	}

	// Validate ownership
	owns, err := h.store.ValidateLabelOwnership(c.Context(), labelID, projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to validate label ownership")
	}
	if !owns {
		return fiber.NewError(fiber.StatusNotFound, "label not found")
	}

	err = h.store.DeleteProjectLabel(c.Context(), labelID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete label")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// HandleReorderLabels updates the position of all labels in a project
// POST /orgs/:org_id/projects/:project_id/labels/reorder
func (h *Handler) HandleReorderLabels(c fiber.Ctx) error {
	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project_id")
	}

	var labelIDs []int64
	err = json.Unmarshal(c.Body(), &labelIDs)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	err = h.store.ReorderLabels(c.Context(), projectID, labelIDs)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reorder labels")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// HandleAddLabelToTask adds a label to a task
// POST /orgs/:org_id/projects/:project_id/tasks/:task_id/labels/:label_id
func (h *Handler) HandleAddLabelToTask(c fiber.Ctx) error {
	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project_id")
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task_id")
	}

	labelID, err := strconv.ParseInt(c.Params("label_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid label_id")
	}

	ctx := c.Context()

	// Get label
	label, err := h.store.GetProjectLabel(ctx, labelID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "label not found")
	}

	// Validate label belongs to project
	if label.ProjectID != projectID {
		return fiber.NewError(fiber.StatusBadRequest, "label does not belong to this project")
	}

	// Add label to task
	err = h.store.AddLabelToTask(ctx, taskID, labelID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to add label to task")
	}

	// Log activity
	userID, err := middleware.GetCurrentUserID(c)
	if err == nil {
		newValue := map[string]interface{}{
			"label_id":   labelID,
			"label_name": label.Name,
		}
		_ = h.store.CreateActivity(c.Context(), taskID, userID, "add_label", "", nil, newValue)
	}

	// Fetch updated labels
	labels, err := h.store.GetTaskLabels(ctx, taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load updated labels")
	}

	// Load all project labels for dropdown
	allLabels, err := h.store.GetProjectLabels(ctx, projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load project labels")
	}

	// Return the full LabelSection for the outerHTML swap
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Redirect().To("/login")
	}

	orgID, _ := strconv.ParseInt(c.Params("org_id"), 10, 64)

	return render(c, components.LabelSection(taskID, labels, allLabels, orgID, projectID, *user))
}

// HandleRemoveLabelFromTask removes a label from a task
// DELETE /orgs/:org_id/projects/:project_id/tasks/:task_id/labels/:label_id
func (h *Handler) HandleRemoveLabelFromTask(c fiber.Ctx) error {
	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project_id")
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task_id")
	}

	labelID, err := strconv.ParseInt(c.Params("label_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid label_id")
	}

	ctx := c.Context()

	// Get label
	label, err := h.store.GetProjectLabel(ctx, labelID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "label not found")
	}

	// Validate label belongs to project
	if label.ProjectID != projectID {
		return fiber.NewError(fiber.StatusBadRequest, "label does not belong to this project")
	}

	// Remove label from task
	err = h.store.RemoveLabelFromTask(ctx, taskID, labelID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to remove label from task")
	}

	// Log activity
	userID, err := middleware.GetCurrentUserID(c)
	if err == nil {
		oldValue := map[string]interface{}{
			"label_id":   labelID,
			"label_name": label.Name,
		}
		_ = h.store.CreateActivity(c.Context(), taskID, userID, "remove_label", "", oldValue, nil)
	}

	// Fetch updated labels
	labels, err := h.store.GetTaskLabels(ctx, taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load updated labels")
	}

	// Load all project labels for dropdown
	allLabels, err := h.store.GetProjectLabels(ctx, projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load project labels")
	}

	// Return the full LabelSection for the outerHTML swap
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Redirect().To("/login")
	}

	orgID, _ := strconv.ParseInt(c.Params("org_id"), 10, 64)

	return render(c, components.LabelSection(taskID, labels, allLabels, orgID, projectID, *user))
}
