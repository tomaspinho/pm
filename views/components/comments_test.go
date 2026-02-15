package components

import (
	"testing"
	"time"

	"pm/internal/model"
)

func TestHasVisibleReplies(t *testing.T) {
	now := time.Now()
	deletedAt := &now

	tests := []struct {
		name    string
		comment model.CommentThread
		want    bool
	}{
		{
			name: "no replies",
			comment: model.CommentThread{
				CommentWithAuthor: model.CommentWithAuthor{
					TaskComment: model.TaskComment{
						ID:        1,
						DeletedAt: nil,
					},
				},
				Replies: []model.CommentThread{},
			},
			want: false,
		},
		{
			name: "has one active reply",
			comment: model.CommentThread{
				CommentWithAuthor: model.CommentWithAuthor{
					TaskComment: model.TaskComment{
						ID:        1,
						DeletedAt: nil,
					},
				},
				Replies: []model.CommentThread{
					{
						CommentWithAuthor: model.CommentWithAuthor{
							TaskComment: model.TaskComment{
								ID:        2,
								DeletedAt: nil,
							},
						},
						Replies: []model.CommentThread{},
					},
				},
			},
			want: true,
		},
		{
			name: "has one deleted reply with no children",
			comment: model.CommentThread{
				CommentWithAuthor: model.CommentWithAuthor{
					TaskComment: model.TaskComment{
						ID:        1,
						DeletedAt: nil,
					},
				},
				Replies: []model.CommentThread{
					{
						CommentWithAuthor: model.CommentWithAuthor{
							TaskComment: model.TaskComment{
								ID:        2,
								DeletedAt: deletedAt,
							},
						},
						Replies: []model.CommentThread{},
					},
				},
			},
			want: false, // Deleted reply with no children is not visible
		},
		{
			name: "has deleted reply with active child",
			comment: model.CommentThread{
				CommentWithAuthor: model.CommentWithAuthor{
					TaskComment: model.TaskComment{
						ID:        1,
						DeletedAt: nil,
					},
				},
				Replies: []model.CommentThread{
					{
						CommentWithAuthor: model.CommentWithAuthor{
							TaskComment: model.TaskComment{
								ID:        2,
								DeletedAt: deletedAt,
							},
						},
						Replies: []model.CommentThread{
							{
								CommentWithAuthor: model.CommentWithAuthor{
									TaskComment: model.TaskComment{
										ID:        3,
										DeletedAt: nil,
									},
								},
								Replies: []model.CommentThread{},
							},
						},
					},
				},
			},
			want: true, // Deleted reply with active child is visible
		},
		{
			name: "A > B (deleted) > C (deleted) - chain of deleted",
			comment: model.CommentThread{
				CommentWithAuthor: model.CommentWithAuthor{
					TaskComment: model.TaskComment{
						ID:        1,
						DeletedAt: nil,
					},
				},
				Replies: []model.CommentThread{
					{
						CommentWithAuthor: model.CommentWithAuthor{
							TaskComment: model.TaskComment{
								ID:        2,
								DeletedAt: deletedAt,
							},
						},
						Replies: []model.CommentThread{
							{
								CommentWithAuthor: model.CommentWithAuthor{
									TaskComment: model.TaskComment{
										ID:        3,
										DeletedAt: deletedAt,
									},
								},
								Replies: []model.CommentThread{},
							},
						},
					},
				},
			},
			want: false, // B has C as child, but C is deleted with no children, so B has no visible replies
		},
		{
			name: "A > B (deleted) > C (deleted) > D (active) - nested visible child",
			comment: model.CommentThread{
				CommentWithAuthor: model.CommentWithAuthor{
					TaskComment: model.TaskComment{
						ID:        1,
						DeletedAt: nil,
					},
				},
				Replies: []model.CommentThread{
					{
						CommentWithAuthor: model.CommentWithAuthor{
							TaskComment: model.TaskComment{
								ID:        2,
								DeletedAt: deletedAt,
							},
						},
						Replies: []model.CommentThread{
							{
								CommentWithAuthor: model.CommentWithAuthor{
									TaskComment: model.TaskComment{
										ID:        3,
										DeletedAt: deletedAt,
									},
								},
								Replies: []model.CommentThread{
									{
										CommentWithAuthor: model.CommentWithAuthor{
											TaskComment: model.TaskComment{
												ID:        4,
												DeletedAt: nil,
											},
										},
										Replies: []model.CommentThread{},
									},
								},
							},
						},
					},
				},
			},
			want: true, // B has visible descendant D through deleted C
		},
		{
			name: "multiple replies, one active",
			comment: model.CommentThread{
				CommentWithAuthor: model.CommentWithAuthor{
					TaskComment: model.TaskComment{
						ID:        1,
						DeletedAt: nil,
					},
				},
				Replies: []model.CommentThread{
					{
						CommentWithAuthor: model.CommentWithAuthor{
							TaskComment: model.TaskComment{
								ID:        2,
								DeletedAt: deletedAt,
							},
						},
						Replies: []model.CommentThread{},
					},
					{
						CommentWithAuthor: model.CommentWithAuthor{
							TaskComment: model.TaskComment{
								ID:        3,
								DeletedAt: nil,
							},
						},
						Replies: []model.CommentThread{},
					},
				},
			},
			want: true, // Has one active reply (ID 3)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasVisibleReplies(tt.comment)
			if got != tt.want {
				t.Errorf("hasVisibleReplies() = %v, want %v", got, tt.want)
			}
		})
	}
}
