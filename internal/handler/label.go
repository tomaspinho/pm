package handler

import (
	"fmt"
	"strconv"
	"strings"

	"pm/internal/middleware"
	"pm/views"

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

// HandleSaveLabelSetup processes the label setup form.
// POST /orgs/:org_id/projects/:project_id/settings/labels
func (h *Handler) HandleSaveLabelSetup(c fiber.Ctx) error {
	orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org_id")
	}

	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project_id")
	}

	type labelUpdate struct {
		ID    *int64
		Name  string
		Color string
	}

	var labels []labelUpdate

	for i := 0; ; i++ {
		nameKey := fmt.Sprintf("labels[%d][name]", i)
		colorKey := fmt.Sprintf("labels[%d][color]", i)
		idKey := fmt.Sprintf("labels[%d][id]", i)

		name := c.FormValue(nameKey)
		if name == "" {
			break
		}

		color := c.FormValue(colorKey)
		idStr := c.FormValue(idKey)

		lbl := labelUpdate{
			Name:  name,
			Color: color,
		}

		if idStr != "" {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err == nil {
				lbl.ID = &id
			}
		}

		labels = append(labels, lbl)
	}

	nameSet := make(map[string]bool)
	for _, lbl := range labels {
		lowerName := strings.ToLower(strings.TrimSpace(lbl.Name))
		if nameSet[lowerName] {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("duplicate label name: %s", lbl.Name))
		}
		nameSet[lowerName] = true
	}

	existingLabels, err := h.store.GetProjectLabels(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load existing labels")
	}

	formLabelIDs := make(map[int64]bool)
	for _, lbl := range labels {
		if lbl.ID != nil {
			formLabelIDs[*lbl.ID] = true
		}
	}

	for _, existing := range existingLabels {
		if !formLabelIDs[existing.ID] {
			err = h.store.DeleteProjectLabel(c.Context(), existing.ID)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to delete label: %v", err))
			}
		}
	}

	for _, lbl := range labels {
		if lbl.ID != nil {
			err = h.store.UpdateProjectLabel(c.Context(), *lbl.ID, lbl.Name, lbl.Color)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to update label: %v", err))
			}
		} else {
			_, err = h.store.CreateProjectLabel(c.Context(), projectID, lbl.Name, lbl.Color)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to create label: %v", err))
			}
		}
	}

	return c.Redirect().To(fmt.Sprintf("/orgs/%d/projects/%d/settings", orgID, projectID))
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

// HandleAddLabelToTask adds a label to a task
// POST /orgs/:org_id/projects/:project_id/tasks/:task_id/labels/:label_id
func (h *Handler) HandleAddLabelToTask(c fiber.Ctx) error {
	orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org_id")
	}

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

	// Get updated task with labels and assignees for kanban card update
	task, err := h.store.GetTask(ctx, taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task")
	}
	taskLabels, err := h.store.GetTaskLabels(ctx, taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task labels")
	}
	taskAssignees, err := h.store.GetTaskAssignees(ctx, taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task assignees")
	}

	// Get activity for OOB update
	activity, err := h.store.GetTaskActivity(ctx, taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get activity")
	}

	return render(c, views.LabelTaskCardUpdateWithActivity(*task, orgID, taskLabels, taskAssignees, activity))
}

// HandleRemoveLabelFromTask removes a label from a task
// DELETE /orgs/:org_id/projects/:project_id/tasks/:task_id/labels/:label_id
func (h *Handler) HandleRemoveLabelFromTask(c fiber.Ctx) error {
	orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org_id")
	}

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

	// Get updated task with labels and assignees for kanban card update
	task, err := h.store.GetTask(ctx, taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task")
	}
	taskLabels, err := h.store.GetTaskLabels(ctx, taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task labels")
	}
	taskAssignees, err := h.store.GetTaskAssignees(ctx, taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get task assignees")
	}

	// Get activity for OOB update
	activity, err := h.store.GetTaskActivity(ctx, taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get activity")
	}

	return render(c, views.LabelTaskCardUpdateWithActivity(*task, orgID, taskLabels, taskAssignees, activity))
}
