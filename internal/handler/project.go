package handler

import (
	"fmt"
	"strconv"

	"cracked-pm/internal/middleware"
	"cracked-pm/views"

	"github.com/gofiber/fiber/v3"
)

// HandleLandingPage redirects authenticated users to their last viewed project,
// or to the project picker if none.
// GET /
func (h *Handler) HandleLandingPage(c fiber.Ctx) error {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Redirect().To("/login")
	}

	// If user has a last viewed project, redirect to it.
	if user.LastViewedProjectID != nil {
		project, err := h.store.GetProject(c.Context(), *user.LastViewedProjectID)
		if err == nil {
			return c.Redirect().To(fmt.Sprintf("/orgs/%d/projects/%d", project.OrganizationID, project.ID))
		}
	}

	// Otherwise, redirect to project picker.
	return c.Redirect().To("/projects")
}

// HandleProjectPicker renders the project picker page.
// GET /projects?org_id=...
func (h *Handler) HandleProjectPicker(c fiber.Ctx) error {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Redirect().To("/login")
	}

	// Get all user's organizations.
	orgs, err := h.store.GetUserOrganizations(c.Context(), user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load organizations")
	}

	// Check if an org_id is specified.
	orgIDStr := c.Query("org_id")

	if orgIDStr == "" && len(orgs) > 0 {
		// Default to first org.
		return c.Redirect().To(fmt.Sprintf("/projects?org_id=%d", orgs[0].ID))
	}

	if orgIDStr == "" {
		// No orgs at all — show empty state.
		nav := views.NavContext{
			User:         user,
			Orgs:         orgs,
			CurrentOrgID: 0,
		}
		return render(c, views.ProjectPickerPage(user, orgs, nil, nil, nav))
	}

	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org_id")
	}

	// Verify membership.
	isMember, err := h.store.IsMember(c.Context(), orgID, user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to check membership")
	}
	if !isMember {
		return fiber.NewError(fiber.StatusForbidden, "not a member of this organization")
	}

	// Get the org and its projects.
	currentOrg, err := h.store.GetOrganization(c.Context(), orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "organization not found")
	}

	projects, err := h.store.GetOrganizationProjects(c.Context(), orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load projects")
	}

	nav := views.NavContext{
		User:         user,
		Orgs:         orgs,
		CurrentOrgID: orgID,
	}

	return render(c, views.ProjectPickerPage(user, orgs, currentOrg, projects, nav))
}

// HandleShowCreateProject renders the new project form.
// GET /orgs/:org_id/projects/new
func (h *Handler) HandleShowCreateProject(c fiber.Ctx) error {
	orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org_id")
	}

	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Redirect().To("/login")
	}

	orgs, err := h.store.GetUserOrganizations(c.Context(), user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load organizations")
	}

	nav := views.NavContext{
		User:         user,
		Orgs:         orgs,
		CurrentOrgID: orgID,
	}

	return render(c, views.NewProjectPage(orgID, nav))
}

// HandleCreateProject processes the new project form.
// POST /orgs/:org_id/projects
func (h *Handler) HandleCreateProject(c fiber.Ctx) error {
	orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org_id")
	}

	name := c.FormValue("name")
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "project name is required")
	}

	description := c.FormValue("description")

	project, err := h.store.CreateProject(c.Context(), name, description, orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create project")
	}

	return c.Redirect().To(fmt.Sprintf("/orgs/%d/projects/%d", orgID, project.ID))
}
