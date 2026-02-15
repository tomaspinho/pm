package store

import (
	"testing"
	"time"

	"pm/internal/model"
)

func TestBuildCommentTree(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		comments   []model.CommentWithAuthor
		maxDepth   int
		want       int                              // expected number of top-level comments
		wantNested func([]model.CommentThread) bool // additional validation
	}{
		{
			name:     "empty comments",
			comments: []model.CommentWithAuthor{},
			maxDepth: 3,
			want:     0,
		},
		{
			name: "single top-level comment",
			comments: []model.CommentWithAuthor{
				{
					TaskComment: model.TaskComment{
						ID:        1,
						TaskID:    1,
						UserID:    1,
						ParentID:  nil,
						Content:   "Top level",
						CreatedAt: now,
					},
					AuthorEmail: "user@test.com",
				},
			},
			maxDepth: 3,
			want:     1,
			wantNested: func(tree []model.CommentThread) bool {
				return tree[0].Depth == 0 && len(tree[0].Replies) == 0
			},
		},
		{
			name: "comment with one reply",
			comments: []model.CommentWithAuthor{
				{
					TaskComment: model.TaskComment{
						ID:        1,
						TaskID:    1,
						UserID:    1,
						ParentID:  nil,
						Content:   "Top level",
						CreatedAt: now,
					},
					AuthorEmail: "user1@test.com",
				},
				{
					TaskComment: model.TaskComment{
						ID:        2,
						TaskID:    1,
						UserID:    2,
						ParentID:  ptr(int64(1)),
						Content:   "Reply to 1",
						CreatedAt: now.Add(time.Minute),
					},
					AuthorEmail: "user2@test.com",
				},
			},
			maxDepth: 3,
			want:     1,
			wantNested: func(tree []model.CommentThread) bool {
				if len(tree) != 1 {
					return false
				}
				if tree[0].Depth != 0 {
					return false
				}
				if len(tree[0].Replies) != 1 {
					return false
				}
				if tree[0].Replies[0].Depth != 1 {
					return false
				}
				if tree[0].Replies[0].ID != 2 {
					return false
				}
				return true
			},
		},
		{
			name: "deeply nested comments at max depth",
			comments: []model.CommentWithAuthor{
				{
					TaskComment: model.TaskComment{
						ID:        1,
						TaskID:    1,
						UserID:    1,
						ParentID:  nil,
						Content:   "Top level",
						CreatedAt: now,
					},
					AuthorEmail: "user1@test.com",
				},
				{
					TaskComment: model.TaskComment{
						ID:        2,
						TaskID:    1,
						UserID:    2,
						ParentID:  ptr(int64(1)),
						Content:   "Reply depth 1",
						CreatedAt: now.Add(time.Minute),
					},
					AuthorEmail: "user2@test.com",
				},
				{
					TaskComment: model.TaskComment{
						ID:        3,
						TaskID:    1,
						UserID:    3,
						ParentID:  ptr(int64(2)),
						Content:   "Reply depth 2",
						CreatedAt: now.Add(2 * time.Minute),
					},
					AuthorEmail: "user3@test.com",
				},
				{
					TaskComment: model.TaskComment{
						ID:        4,
						TaskID:    1,
						UserID:    4,
						ParentID:  ptr(int64(3)),
						Content:   "Reply depth 3",
						CreatedAt: now.Add(3 * time.Minute),
					},
					AuthorEmail: "user4@test.com",
				},
				{
					TaskComment: model.TaskComment{
						ID:        5,
						TaskID:    1,
						UserID:    5,
						ParentID:  ptr(int64(4)),
						Content:   "Reply depth 4 (should be flattened)",
						CreatedAt: now.Add(4 * time.Minute),
					},
					AuthorEmail: "user5@test.com",
				},
			},
			maxDepth: 3,
			want:     1,
			wantNested: func(tree []model.CommentThread) bool {
				// Should have: 1 (depth 0) -> 2 (depth 1) -> 3 (depth 2) -> 4 (depth 3), and 5 flattened at depth 3
				if len(tree) != 1 || tree[0].Depth != 0 {
					t.Logf("Expected 1 top-level comment at depth 0, got %d comments", len(tree))
					return false
				}
				if len(tree[0].Replies) != 1 || tree[0].Replies[0].Depth != 1 {
					t.Logf("Expected 1 reply at depth 1, got %d replies", len(tree[0].Replies))
					return false
				}
				if len(tree[0].Replies[0].Replies) != 1 || tree[0].Replies[0].Replies[0].Depth != 2 {
					t.Logf("Expected 1 reply at depth 2, got %d replies", len(tree[0].Replies[0].Replies))
					return false
				}
				depth2Comment := tree[0].Replies[0].Replies[0]
				if len(depth2Comment.Replies) != 1 || depth2Comment.Replies[0].Depth != 3 {
					t.Logf("Expected 1 reply at depth 3, got %d replies", len(depth2Comment.Replies))
					return false
				}
				// At depth 3 (maxDepth), comment 4 should have one flattened reply (comment 5 at depth 3)
				depth3Comment := depth2Comment.Replies[0]
				if len(depth3Comment.Replies) != 1 {
					t.Logf("Expected depth 3 comment to have 1 flattened reply, got %d", len(depth3Comment.Replies))
					return false
				}
				// The flattened comment should be at depth 3 (same as parent, at maxDepth)
				if depth3Comment.Replies[0].Depth != 3 {
					t.Logf("Expected flattened comment at depth 3, got depth %d", depth3Comment.Replies[0].Depth)
					return false
				}
				if depth3Comment.Replies[0].ID != 5 {
					t.Logf("Expected flattened comment ID 5, got ID %d", depth3Comment.Replies[0].ID)
					return false
				}
				return true
			},
		},
		{
			name: "multiple top-level comments",
			comments: []model.CommentWithAuthor{
				{
					TaskComment: model.TaskComment{
						ID:        1,
						TaskID:    1,
						UserID:    1,
						ParentID:  nil,
						Content:   "First top level",
						CreatedAt: now,
					},
					AuthorEmail: "user1@test.com",
				},
				{
					TaskComment: model.TaskComment{
						ID:        2,
						TaskID:    1,
						UserID:    2,
						ParentID:  nil,
						Content:   "Second top level",
						CreatedAt: now.Add(time.Minute),
					},
					AuthorEmail: "user2@test.com",
				},
			},
			maxDepth: 3,
			want:     2,
			wantNested: func(tree []model.CommentThread) bool {
				return tree[0].Depth == 0 && tree[1].Depth == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildCommentTree(tt.comments, tt.maxDepth)
			if len(got) != tt.want {
				t.Errorf("BuildCommentTree() returned %d top-level comments, want %d", len(got), tt.want)
			}
			if tt.wantNested != nil && !tt.wantNested(got) {
				t.Errorf("BuildCommentTree() failed nested validation")
			}
		})
	}
}

func ptr(i int64) *int64 {
	return &i
}
