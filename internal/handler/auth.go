package handler

import (
	"fmt"
	"strings"
	"time"

	"pm/internal/auth"
	"pm/views"

	"github.com/gofiber/fiber/v3"
)

// HandleShowLogin renders the login page.
// GET /login
func (h *Handler) HandleShowLogin(c fiber.Ctx) error {
	return render(c, views.LoginPage(""))
}

// HandleLogin processes the login form.
// POST /login
func (h *Handler) HandleLogin(c fiber.Ctx) error {
	email := strings.TrimSpace(c.FormValue("email"))
	password := c.FormValue("password")
	remember := c.FormValue("remember") == "true"

	// Validate email format.
	email, err := auth.ValidateEmail(email)
	if err != nil {
		return render(c, views.LoginPage("Invalid email address."))
	}

	// Look up user.
	user, err := h.store.GetUserByEmail(c.Context(), email)
	if err != nil {
		return render(c, views.LoginPage("Invalid email or password."))
	}

	// Check password.
	if !auth.CheckPassword(user.PasswordHash, password) {
		return render(c, views.LoginPage("Invalid email or password."))
	}

	// Create session.
	sessionID, err := auth.GenerateSessionID()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to generate session")
	}

	duration := auth.SessionDuration
	if remember {
		duration = auth.SessionDurationRemember
	}
	expiresAt := time.Now().Add(duration)

	_, err = h.store.CreateSession(c.Context(), sessionID, user.ID, expiresAt)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create session")
	}

	// Set session cookie.
	c.Cookie(&fiber.Cookie{
		Name:     auth.SessionCookieName,
		Value:    sessionID,
		MaxAge:   int(duration.Seconds()),
		HTTPOnly: true,
		SameSite: "Lax",
		Path:     "/",
	})

	return c.Redirect().To("/")
}

// HandleShowSignup renders the signup page.
// GET /signup
func (h *Handler) HandleShowSignup(c fiber.Ctx) error {
	return render(c, views.SignupPage(""))
}

// HandleSignup processes the signup form.
// POST /signup
func (h *Handler) HandleSignup(c fiber.Ctx) error {
	email := strings.TrimSpace(c.FormValue("email"))
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")
	displayName := strings.TrimSpace(c.FormValue("display_name"))

	// Validate display name.
	if len(displayName) == 0 || len(displayName) > 100 {
		return render(c, views.SignupPage("Display name must be between 1 and 100 characters."))
	}

	// Validate email.
	email, err := auth.ValidateEmail(email)
	if err != nil {
		return render(c, views.SignupPage("Invalid email address."))
	}

	// Check passwords match.
	if password != passwordConfirm {
		return render(c, views.SignupPage("Passwords do not match."))
	}

	// Validate password strength.
	if err := auth.ValidatePassword(password); err != nil {
		return render(c, views.SignupPage(err.Error()))
	}

	// Check if email already taken.
	if _, err := h.store.GetUserByEmail(c.Context(), email); err == nil {
		return render(c, views.SignupPage("An account with this email already exists."))
	}

	// Hash password.
	hash, err := auth.HashPassword(password)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to hash password")
	}

	// Create user.
	user, err := h.store.CreateUser(c.Context(), email, hash, displayName)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create user")
	}

	// Create personal organization.
	orgName := fmt.Sprintf("%s's Organization", email)
	_, err = h.store.CreateOrganization(c.Context(), orgName, user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create organization")
	}

	// Create session and log user in.
	sessionID, err := auth.GenerateSessionID()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to generate session")
	}

	expiresAt := time.Now().Add(auth.SessionDuration)
	_, err = h.store.CreateSession(c.Context(), sessionID, user.ID, expiresAt)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create session")
	}

	c.Cookie(&fiber.Cookie{
		Name:     auth.SessionCookieName,
		Value:    sessionID,
		MaxAge:   int(auth.SessionDuration.Seconds()),
		HTTPOnly: true,
		SameSite: "Lax",
		Path:     "/",
	})

	return c.Redirect().To("/")
}

// HandleLogout destroys the session and redirects to login.
// POST /logout
func (h *Handler) HandleLogout(c fiber.Ctx) error {
	sessionID := c.Cookies(auth.SessionCookieName)
	if sessionID != "" {
		_ = h.store.DeleteSession(c.Context(), sessionID)
	}

	// Clear cookie.
	c.Cookie(&fiber.Cookie{
		Name:     auth.SessionCookieName,
		Value:    "",
		MaxAge:   -1,
		HTTPOnly: true,
		SameSite: "Lax",
		Path:     "/",
	})

	return c.Redirect().To("/login")
}
