package handler

import (
	"strings"

	"pm/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

// HandleUpdateProfile updates the current user's display name.
// PATCH /profile
func (h *Handler) HandleUpdateProfile(c fiber.Ctx) error {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		return err
	}

	displayName := strings.TrimSpace(c.FormValue("display_name"))

	// Validate display name
	if len(displayName) == 0 || len(displayName) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Display name must be between 1 and 100 characters",
		})
	}

	// Update in database
	if err = h.store.UpdateUserDisplayName(c.Context(), currentUser.ID, displayName); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update display name",
		})
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"display_name": displayName,
	})
}

// HandleGetOrgMembers returns all members of an organization as JSON.
// GET /orgs/:org_id/members
func (h *Handler) HandleGetOrgMembers(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org or project id")
	}
	_ = projectID // Not used but required by parseOrgAndProject

	members, err := h.store.GetOrganizationMembers(c.Context(), orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get organization members")
	}

	return c.JSON(members)
}
