package handler

import (
	"pm/internal/store"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	store *store.Store
}

// New creates a new Handler.
func New(s *store.Store) *Handler {
	return &Handler{store: s}
}

// render writes a templ component to the Fiber response.
func render(c fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return component.Render(c.Context(), c.Response().BodyWriter())
}
