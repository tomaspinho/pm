package handler

import (
	"strconv"
	"strings"

	"cracked-pm/internal/middleware"
	"cracked-pm/internal/model"
	"cracked-pm/internal/store"
	"cracked-pm/views/components"

	"github.com/gofiber/fiber/v3"
)

// HandleCreateComment creates a new top-level comment on a task.
// POST /orgs/:org_id/projects/:project_id/tasks/:id/comments
func (h *Handler) HandleCreateComment(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Redirect().To("/login")
	}

	content := strings.TrimSpace(c.FormValue("content"))
	if content == "" {
		return fiber.NewError(fiber.StatusBadRequest, "comment content is required")
	}

	// Create comment with parent_id = NULL (top-level)
	_, err = h.store.CreateComment(c.Context(), taskID, user.ID, nil, content)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create comment")
	}

	// Reload all comments and build tree
	comments, err := h.store.ListTaskComments(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load comments")
	}

	commentTree := store.BuildCommentTree(comments, 3)

	return render(c, components.CommentSection(taskID, orgID, projectID, commentTree, user))
}

// HandleReplyToComment creates a reply to an existing comment.
// POST /orgs/:org_id/projects/:project_id/tasks/:id/comments/:parent_id/reply
func (h *Handler) HandleReplyToComment(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	parentID, err := strconv.ParseInt(c.Params("parent_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid parent comment id")
	}

	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Redirect().To("/login")
	}

	content := strings.TrimSpace(c.FormValue("content"))
	if content == "" {
		return fiber.NewError(fiber.StatusBadRequest, "comment content is required")
	}

	// Verify parent comment exists and belongs to this task
	parentComment, err := h.store.GetComment(c.Context(), parentID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "parent comment not found")
	}
	if parentComment.TaskID != taskID {
		return fiber.NewError(fiber.StatusBadRequest, "parent comment does not belong to this task")
	}

	// Check depth of parent comment
	depth, err := h.store.GetCommentDepth(c.Context(), parentID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to check comment depth")
	}

	// If parent is at depth 3 or more, cannot add reply
	if depth >= 3 {
		return fiber.NewError(fiber.StatusBadRequest, "maximum nesting depth reached")
	}

	// Create reply
	_, err = h.store.CreateComment(c.Context(), taskID, user.ID, &parentID, content)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create reply")
	}

	// Reload all comments and build tree
	comments, err := h.store.ListTaskComments(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load comments")
	}

	commentTree := store.BuildCommentTree(comments, 3)

	return render(c, components.CommentSection(taskID, orgID, projectID, commentTree, user))
}

// HandleUpdateComment updates a comment's content.
// PATCH /orgs/:org_id/projects/:project_id/tasks/:id/comments/:comment_id
func (h *Handler) HandleUpdateComment(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	commentID, err := strconv.ParseInt(c.Params("comment_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid comment id")
	}

	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Redirect().To("/login")
	}

	// Load comment and verify ownership
	comment, err := h.store.GetComment(c.Context(), commentID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "comment not found")
	}

	if comment.UserID != user.ID {
		return fiber.NewError(fiber.StatusForbidden, "you can only edit your own comments")
	}

	if comment.TaskID != taskID {
		return fiber.NewError(fiber.StatusBadRequest, "comment does not belong to this task")
	}

	content := strings.TrimSpace(c.FormValue("content"))
	if content == "" {
		return fiber.NewError(fiber.StatusBadRequest, "comment content is required")
	}

	// Update comment
	err = h.store.UpdateComment(c.Context(), commentID, content)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update comment")
	}

	// Reload comment with author for display
	updatedComment, err := h.store.GetCommentWithAuthor(c.Context(), commentID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reload comment")
	}

	// Build a single-depth comment thread for display
	commentThread := store.BuildCommentTree([]model.CommentWithAuthor{*updatedComment}, 3)
	if len(commentThread) == 0 {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to build comment thread")
	}

	return render(c, components.CommentDisplayOOB(commentThread[0], taskID, orgID, projectID, true))
}

// HandleDeleteComment soft-deletes a comment.
// DELETE /orgs/:org_id/projects/:project_id/tasks/:id/comments/:comment_id
func (h *Handler) HandleDeleteComment(c fiber.Ctx) error {
	orgID, projectID, err := parseOrgAndProject(c)
	if err != nil {
		return err
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid task id")
	}

	commentID, err := strconv.ParseInt(c.Params("comment_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid comment id")
	}

	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Redirect().To("/login")
	}

	// Load comment and verify ownership
	comment, err := h.store.GetComment(c.Context(), commentID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "comment not found")
	}

	if comment.UserID != user.ID {
		return fiber.NewError(fiber.StatusForbidden, "you can only delete your own comments")
	}

	if comment.TaskID != taskID {
		return fiber.NewError(fiber.StatusBadRequest, "comment does not belong to this task")
	}

	// Soft delete comment
	err = h.store.DeleteComment(c.Context(), commentID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete comment")
	}

	// Reload all comments and build tree
	comments, err := h.store.ListTaskComments(c.Context(), taskID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load comments")
	}

	commentTree := store.BuildCommentTree(comments, 3)

	return render(c, components.CommentSection(taskID, orgID, projectID, commentTree, user))
}
