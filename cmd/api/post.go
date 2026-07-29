package main

import (
	"errors"
	"net/http"

	"github.com/loeia/gopherSocialNetwork/internal/store"
)

type postKey string

const postCtx postKey = "post"

type PostResponse struct {
	ID        int64              `json:"id"`
	AuthorId  int64              `json:"author_id"`
	Author    string             `json:"author"`
	Title     string             `json:"title"`
	Content   string             `json:"content"`
	Tags      []string           `json:"tags"`
	Comments  []*CommentResponse `json:"comments,omitempty"`
	CreatedAt string             `json:"created_at"`
	UpdatedAt string             `json:"updated_at"`
}

type CommentResponse struct {
	ID        int64  `json:"id"`
	PostID    int64  `json:"post_id"`
	Username  string `json:"username"`
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

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

	if err := app.JSONResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	comments, err := app.store.Comments.GetById(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	var responseComments []*CommentResponse
	for _, c := range comments {
		rc := CommentResponse{
			ID:        c.ID,
			PostID:    c.PostID,
			UserID:    c.UserID,
			Username:  c.User.Username,
			Content:   c.Content,
			CreatedAt: c.CreatedAt,
		}
		responseComments = append(responseComments, &rc)
	}

	resp := PostResponse{
		ID:        post.ID,
		AuthorId:  post.UserID,
		Author:    post.User.Username,
		Title:     post.Title,
		Content:   post.Content,
		Tags:      post.Tags,
		Comments:  responseComments,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
	}

	if err := app.JSONResponse(w, http.StatusOK, resp); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

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
