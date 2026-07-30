package main

import (
	"net/http"

	"github.com/loeia/gopherSocialNetwork/internal/store"
)

type FeedPostResponse struct {
	ID           int64    `json:"id"`
	AuthorId     int64    `json:"author_id"`
	Author       string   `json:"author"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	Tags         []string `json:"tags"`
	PostLikes    int64    `json:"like_count"`
	CommentCount int64    `json:"comment_count"`
	CreatedAt    string   `json:"created_at"`
}

// getUserFeedHandler returns the paginated feed for the authenticated user.
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	pfq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	p, err := pfq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(p); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	user := getUserFromCtx(r)

	feed, err := app.store.Posts.GetUserFeed(ctx, user.ID, p)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	resp := make([]*FeedPostResponse, len(feed))
	for i, f := range feed {
		resp[i] = &FeedPostResponse{
			ID:           f.Post.ID,
			AuthorId:     f.Post.UserID,
			Author:       f.Post.User.Username,
			Title:        f.Post.Title,
			Content:      f.Post.Content,
			Tags:         f.Post.Tags,
			PostLikes:    f.LikeCount,
			CommentCount: f.CommentCount,
			CreatedAt:    f.Post.CreatedAt,
		}
	}

	if err := app.JSONResponse(w, http.StatusOK, resp); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
