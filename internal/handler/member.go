package handler

import (
	"strconv"
	"strings"

	"pm/internal/middleware"
	"pm/views"

	"github.com/gofiber/fiber/v3"
)

func (h *Handler) HandleShowMembers(c fiber.Ctx) error {
	orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org_id")
	}

	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Redirect().To("/login")
	}

	org, err := h.store.GetOrganization(c.Context(), orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "organization not found")
	}

	members, err := h.store.GetOrganizationMembers(c.Context(), orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get members")
	}

	membersWithRole := make([]views.MemberWithRole, len(members))
	for i, m := range members {
		membersWithRole[i] = views.MemberWithRole{
			User:    m,
			IsOwner: m.ID == org.OwnerUserID,
		}
	}

	invitations, err := h.store.GetPendingInvitationsByOrg(c.Context(), orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get invitations")
	}

	invitationsWithInviter := make([]views.InvitationWithOrg, len(invitations))
	for i, inv := range invitations {
		inviter, err := h.store.GetUser(c.Context(), inv.InvitedBy)
		inviterName := "Unknown"
		if err == nil {
			inviterName = inviter.DisplayName
		}
		invitationsWithInviter[i] = views.InvitationWithOrg{
			ID:        inv.ID,
			OrgID:     inv.OrgID,
			OrgName:   org.Name,
			Email:     inv.Email,
			InvitedBy: inviterName,
			CreatedAt: inv.CreatedAt,
		}
	}

	orgs, err := h.store.GetUserOrganizations(c.Context(), user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load organizations")
	}

	nav := views.NavContext{
		User:             user,
		Orgs:             orgs,
		CurrentOrgID:     orgID,
		CurrentProjectID: 0,
	}

	return render(c, views.MembersPage(org, membersWithRole, invitationsWithInviter, user.ID, nav))
}

func (h *Handler) HandleInviteMember(c fiber.Ctx) error {
	orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org_id")
	}

	email := strings.TrimSpace(c.FormValue("email"))
	if email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email is required")
	}

	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	existingUser, err := h.store.GetUserByEmail(c.Context(), email)
	if err == nil {
		isMember, err := h.store.IsMember(c.Context(), orgID, existingUser.ID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to check membership")
		}
		if isMember {
			return fiber.NewError(fiber.StatusBadRequest, "user is already a member")
		}

		if err := h.store.AddMember(c.Context(), orgID, existingUser.ID); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to add member")
		}

		return h.HandleShowMembers(c)
	}

	if _, err := h.store.CreateInvitation(c.Context(), orgID, email, currentUser.ID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create invitation")
	}

	return h.HandleShowMembers(c)
}

func (h *Handler) HandleRemoveMember(c fiber.Ctx) error {
	orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org_id")
	}

	userID, err := strconv.ParseInt(c.Params("user_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user_id")
	}

	org, err := h.store.GetOrganization(c.Context(), orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "organization not found")
	}

	if userID == org.OwnerUserID {
		return fiber.NewError(fiber.StatusBadRequest, "cannot remove the owner")
	}

	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	if currentUser.ID != org.OwnerUserID && currentUser.ID != userID {
		return fiber.NewError(fiber.StatusForbidden, "only the owner can remove members")
	}

	if err := h.store.RemoveMember(c.Context(), orgID, userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to remove member")
	}

	if currentUser.ID == userID {
		return c.Redirect().To("/projects")
	}

	return h.HandleShowMembers(c)
}

func (h *Handler) HandleCancelInvitation(c fiber.Ctx) error {
	invitationID, err := strconv.ParseInt(c.Params("invitation_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid invitation_id")
	}

	orgID, err := strconv.ParseInt(c.Params("org_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid org_id")
	}

	org, err := h.store.GetOrganization(c.Context(), orgID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "organization not found")
	}

	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	if currentUser.ID != org.OwnerUserID {
		return fiber.NewError(fiber.StatusForbidden, "only the owner can cancel invitations")
	}

	if err := h.store.DeleteInvitation(c.Context(), invitationID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to cancel invitation")
	}

	return h.HandleShowMembers(c)
}
