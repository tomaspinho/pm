package handler

import (
	"strconv"

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
