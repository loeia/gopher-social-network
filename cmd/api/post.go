package main

import (
	"errors"
	"net/http"

	"github.com/loeia/gopherSocialNetwork/internal/store"
)

type postKey string

const postCtx postKey = "postId"

type PostResponse struct {
	ID           int64    `json:"id"`
	AuthorId     int64    `json:"author_id"`
	Author       string   `json:"author"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	Tags         []string `json:"tags"`
	CommentCount int64    `json:"comments_count"`
	LikeCount    int64    `json:"likes_count"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
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

	resp := PostResponse{
		ID:           post.ID,
		AuthorId:     post.UserID,
		Author:       post.User.Username,
		Title:        post.Title,
		Content:      post.Content,
		Tags:         post.Tags,
		CommentCount: post.CommentCount,
		LikeCount:    post.LikeCount,
		CreatedAt:    post.CreatedAt,
		UpdatedAt:    post.UpdatedAt,
	}

	if err := app.JSONResponse(w, http.StatusCreated, resp); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// getPostHandler returns a single post with its comments.
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	// comments, err := app.store.Comments.GetByPostId(r.Context(), post.ID)
	// if err != nil {
	// 	app.internalServerError(w, r, err)
	// 	return
	// }

	// comments
	// var responseComments []*CommentResponse
	// for _, c := range comments {
	// 	rc := CommentResponse{
	// 		ID:        c.ID,
	// 		Username:  c.User.Username,
	// 		Content:   c.Content,
	// 		CreatedAt: c.CreatedAt,
	// 	}
	// 	responseComments = append(responseComments, &rc)
	// }

	// likes
	// postLikes, err := app.store.PostLikes.GetPostLikes(r.Context(), post.ID)
	// if err != nil {
	// 	app.internalServerError(w, r, err)
	// 	return
	// }

	resp := PostResponse{
		ID:           post.ID,
		AuthorId:     post.UserID,
		Author:       post.User.Username,
		Title:        post.Title,
		Content:      post.Content,
		Tags:         post.Tags,
		CommentCount: post.CommentCount,
		LikeCount:    post.LikeCount,
		CreatedAt:    post.CreatedAt,
		UpdatedAt:    post.UpdatedAt,
	}

	if err := app.JSONResponse(w, http.StatusOK, resp); err != nil {
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

	if err := app.JSONResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// getRandomPosts returns random posts for unauthenticated users.
func (app *application) getRandomPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := app.store.Posts.GetRandomPosts(r.Context(), 20)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusOK, posts); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
