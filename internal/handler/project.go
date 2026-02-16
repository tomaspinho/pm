package handler

import (
	"fmt"
	"strconv"

	"pm/internal/middleware"
	"pm/internal/model"
	"pm/views"

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

	// Redirect to column setup page instead of creating default columns automatically.
	return c.Redirect().To(fmt.Sprintf("/orgs/%d/projects/%d/columns/setup", orgID, project.ID))
}

// HandleShowColumnSetup renders the column setup page for a new project.
// GET /orgs/:org_id/projects/:project_id/columns/setup
func (h *Handler) HandleShowColumnSetup(c fiber.Ctx) error {
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

	// Load project to display name.
	project, err := h.store.GetProject(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "project not found")
	}

	// Check if columns already exist.
	existingColumns, err := h.store.GetProjectColumns(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load columns")
	}

	// Get user's organizations for nav.
	orgs, err := h.store.GetUserOrganizations(c.Context(), user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load organizations")
	}

	nav := views.NavContext{
		User:         user,
		Orgs:         orgs,
		CurrentOrgID: orgID,
	}

	return render(c, views.ColumnSetupPage(project, orgID, existingColumns, nav))
}

// HandleSaveColumnSetup processes the column setup form.
// POST /orgs/:org_id/projects/:project_id/columns/setup
func (h *Handler) HandleSaveColumnSetup(c fiber.Ctx) error {
	orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org_id")
	}

	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project_id")
	}

	// Parse form data - columns come as indexed form values.
	// Expected format: columns[0][name], columns[0][color], columns[1][name], etc.

	// Build columns from form data.
	var columns []struct {
		Name  string
		Color string
		Pos   int
	}

	// Parse columns from form values.
	for i := 0; ; i++ {
		nameKey := fmt.Sprintf("columns[%d][name]", i)
		colorKey := fmt.Sprintf("columns[%d][color]", i)

		name := c.FormValue(nameKey)
		color := c.FormValue(colorKey)

		if name == "" {
			break // No more columns.
		}

		columns = append(columns, struct {
			Name  string
			Color string
			Pos   int
		}{
			Name:  name,
			Color: color,
			Pos:   i,
		})
	}

	// Validate: at least 1 column required.
	if len(columns) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "at least one column is required")
	}

	// Validate: no duplicate names.
	nameSet := make(map[string]bool)
	for _, col := range columns {
		if nameSet[col.Name] {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("duplicate column name: %s", col.Name))
		}
		nameSet[col.Name] = true
	}

	// Convert to model.ProjectColumn slice.
	var modelColumns []model.ProjectColumn
	for _, col := range columns {
		modelColumns = append(modelColumns, model.ProjectColumn{
			Name:     col.Name,
			Color:    col.Color,
			Position: col.Pos,
		})
	}

	// Create columns in database.
	err = h.store.CreateProjectColumns(c.Context(), projectID, modelColumns)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create columns")
	}

	// Redirect to the project board.
	return c.Redirect().To(fmt.Sprintf("/orgs/%d/projects/%d", orgID, projectID))
}
