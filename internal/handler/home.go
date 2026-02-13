package handler

import (
	"time"

	"cracked-pm/views"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
)

// HandleHome renders the home page.
func HandleHome(c fiber.Ctx) error {
	return render(c, views.HomePage())
}

// HandleTime returns the current server time as an htmx partial.
func HandleTime(c fiber.Ctx) error {
	now := time.Now().Format(time.RFC1123)
	return render(c, views.TimeDisplay(now))
}

// render is a helper that renders a templ component into a Fiber response.
func render(c fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return component.Render(c.Context(), c.Response().BodyWriter())
}
