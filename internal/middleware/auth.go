package middleware

import (
	"log"
	"strconv"

	"pm/internal/auth"
	"pm/internal/model"
	"pm/internal/store"

	"github.com/gofiber/fiber/v3"
)

const (
	// Context keys for storing user data.
	contextKeyUser   = "user"
	contextKeyUserID = "user_id"
)

// RequireAuth checks the session cookie and loads the user into the Fiber context.
// If no valid session exists, it redirects to the login page.
func RequireAuth(s *store.Store) fiber.Handler {
	return func(c fiber.Ctx) error {
		sessionID := c.Cookies(auth.SessionCookieName)
		if sessionID == "" {
			return c.Redirect().To("/login")
		}

		session, err := s.GetSession(c.Context(), sessionID)
		if err != nil {
			// Invalid or expired session — clear cookie and redirect.
			clearSessionCookie(c)
			return c.Redirect().To("/login")
		}

		user, err := s.GetUserByID(c.Context(), session.UserID)
		if err != nil {
			log.Printf("session references unknown user %d: %v", session.UserID, err)
			clearSessionCookie(c)
			return c.Redirect().To("/login")
		}

		c.Locals(contextKeyUser, user)
		c.Locals(contextKeyUserID, user.ID)
		return c.Next()
	}
}

// RequireOrgAccess verifies the current user belongs to the organization specified
// in the :org_id route parameter. Returns 403 if unauthorized.
func RequireOrgAccess(s *store.Store) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := GetCurrentUserID(c)
		if err != nil {
			return c.Redirect().To("/login")
		}

		orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid organization id")
		}

		isMember, err := s.IsMember(c.Context(), orgID, userID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to check membership")
		}

		if !isMember {
			return fiber.NewError(fiber.StatusForbidden, "you do not have access to this organization")
		}

		return c.Next()
	}
}

// GetCurrentUser extracts the authenticated user from the Fiber context.
func GetCurrentUser(c fiber.Ctx) (*model.User, error) {
	user, ok := c.Locals(contextKeyUser).(*model.User)
	if !ok || user == nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "not authenticated")
	}
	return user, nil
}

// GetCurrentUserID extracts the authenticated user's ID from the Fiber context.
func GetCurrentUserID(c fiber.Ctx) (int64, error) {
	id, ok := c.Locals(contextKeyUserID).(int64)
	if !ok {
		return 0, fiber.NewError(fiber.StatusUnauthorized, "not authenticated")
	}
	return id, nil
}

// clearSessionCookie removes the session cookie.
func clearSessionCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     auth.SessionCookieName,
		Value:    "",
		MaxAge:   -1,
		HTTPOnly: true,
		SameSite: "Lax",
	})
}
