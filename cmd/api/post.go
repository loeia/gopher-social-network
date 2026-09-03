package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/loeia/gopherSocialNetwork/internal/store"
)

type postKey string

const postCtx postKey = "postId"

type PostResponse struct {
	ID           int64     `json:"id"`
	AuthorId     int64     `json:"author_id"`
	Author       string    `json:"author"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Tags         []string  `json:"tags"`
	CommentCount int64     `json:"comment_count"`
	LikeCount    int64     `json:"like_count"`
	ViewCount    int64     `json:"view_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func postResponse(p *store.Post) *PostResponse {
	return &PostResponse{
		ID:           p.ID,
		AuthorId:     p.UserID,
		Author:       p.User.Username,
		Title:        p.Title,
		Content:      p.Content,
		Tags:         p.Tags,
		CommentCount: p.CommentCount,
		LikeCount:    p.LikeCount,
		ViewCount:    p.ViewCount,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

// createPostHandler handles creating a new post.
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var p CreatePostPayload
	if err := app.readJSON(w, r, &p); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(p); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	user := getUserFromCtx(r)
	post := store.Post{
		UserID:  user.ID,
		Title:   p.Title,
		Content: p.Content,
		Tags:    p.Tags,
	}

	if err := app.store.Posts.Create(r.Context(), &post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.invalidateUserCache(r, post.UserID)

	if err := app.JSONResponse(w, http.StatusCreated, postResponse(&post)); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// getPostHandler returns a single post with its comments.
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	if err := app.store.Posts.IncrementViewCount(r.Context(), post.ID); err != nil {
		app.logger.Errorw("failed to increment view count", "error", err)
	}

	post.ViewCount++

	if err := app.JSONResponse(w, http.StatusOK, postResponse(post)); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// deletePostHandler handles deleting a post.
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	if err := app.store.Posts.Delete(r.Context(), post); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	app.invalidatePostCache(r, post.ID)
	app.invalidateUserCache(r, post.UserID)

	w.WriteHeader(http.StatusNoContent)
}

// updatePostHandler handles updating a post's title, content, or tags.
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var payload UpdatePostPayload
	if err := app.readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Tags != nil {
		post.Tags = *payload.Tags
	}

	if err := app.store.Posts.Update(r.Context(), post); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	app.invalidatePostCache(r, post.ID)

	if err := app.JSONResponse(w, http.StatusOK, postResponse(post)); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// getFreePostsHandler returns random posts for unauthenticated users.
func (app *application) getFreePostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := app.store.Posts.GetFree(r.Context(), 20)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	resp := make([]*PostResponse, len(posts))
	for i, post := range posts {
		resp[i] = postResponse(post)
	}

	if err := app.JSONResponse(w, http.StatusOK, resp); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getSearchPostHandler(w http.ResponseWriter, r *http.Request) {
	pfq := store.FilterQuery{
		Limit:  20,
		Offset: 0,
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

	posts, err := app.store.Posts.Search(r.Context(), p)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	resp := make([]*PostResponse, len(posts))
	for i, post := range posts {
		resp[i] = postResponse(post)
	}

	if err := app.JSONResponse(w, http.StatusOK, resp); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// invalidatePostCache deletes a post's cached data so the next read fetches fresh data.
func (app *application) invalidatePostCache(r *http.Request, postId int64) {
	if !app.config.redisCfg.enabled {
		return
	}
	if err := app.cache.Post.Delete(r.Context(), postId); err != nil {
		app.logger.Errorw("error deleting post from cache", "error", err)
	}
}

// getUserFeedHandler returns the paginated feed for the authenticated user.
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	pfq := store.PaginationQuery{
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

	user := getUserFromCtx(r)

	feed, err := app.store.Posts.GetUserFeed(r.Context(), user.ID, p)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	resp := make([]*PostResponse, len(feed))
	for i, p := range feed {
		resp[i] = postResponse(p)
	}

	if err := app.JSONResponse(w, http.StatusOK, resp); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
