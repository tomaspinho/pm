package handler

import (
	"strconv"

	"pm/internal/middleware"
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
		User:         user,
		Orgs:         orgs,
		CurrentOrgID: orgID,
	}

	// Track last viewed project.
	_ = h.store.UpdateLastViewedProject(ctx, user.ID, project.ID)

	return render(c, views.BoardPage(project, columns, grouped, nav))
}
