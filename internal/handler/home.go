package handler

import (
	"cracked-pm/internal/store"
	"cracked-pm/views"

	"github.com/gofiber/fiber/v3"
)

// HandleHome renders the kanban board for the default project (ID=1).
func (h *Handler) HandleHome(c fiber.Ctx) error {
	ctx := c.Context()

	project, err := h.store.GetProject(ctx, 1)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "project not found")
	}

	tasks, err := h.store.ListTasksByProject(ctx, project.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load tasks")
	}

	grouped := store.GroupTasksByStatus(tasks)
	return render(c, views.BoardPage(project, grouped))
}
