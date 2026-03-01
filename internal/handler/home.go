package handler

import (
	"strconv"

	"pm/internal/middleware"
	"pm/internal/model"
	"pm/internal/store"
	"pm/views"

	"github.com/gofiber/fiber/v3"
)

// HandleBoard renders the kanban board for a specific project within an organization.
// GET /orgs/:org_id/projects/:project_id
func (h *Handler) HandleBoard(c fiber.Ctx) error {
	ctx := c.Context()

	orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org_id")
	}

	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project_id")
	}

	project, err := h.store.GetProject(ctx, projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "project not found")
	}

	// Verify the project belongs to this organization.
	if project.OrganizationID != orgID {
		return fiber.NewError(fiber.StatusNotFound, "project not found in this organization")
	}

	tasks, err := h.store.ListTasksByProject(ctx, project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load tasks")
	}

	columns, err := h.store.GetProjectColumns(ctx, project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load columns")
	}

	grouped := store.GroupTasksByColumn(tasks, columns)

	// Load project labels
	labels, err := h.store.GetProjectLabels(ctx, project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load labels")
	}

	// Load labels for each task
	tasksWithLabels := make(map[int64][]model.Label)
	for _, task := range tasks {
		taskLabels, err := h.store.GetTaskLabels(ctx, task.ID)
		if err == nil {
			tasksWithLabels[task.ID] = taskLabels
		}
	}

	// Build nav context.
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Redirect().To("/login")
	}

	orgs, err := h.store.GetUserOrganizations(ctx, user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load organizations")
	}

	nav := views.NavContext{
		User:             user,
		Orgs:             orgs,
		CurrentOrgID:     orgID,
		CurrentProjectID: projectID,
	}

	// Track last viewed project.
	_ = h.store.UpdateLastViewedProject(ctx, user.ID, project.ID)

	return render(c, views.BoardPage(project, columns, grouped, labels, tasksWithLabels, nav))
}
