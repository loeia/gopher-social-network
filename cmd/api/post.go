package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/loeia/gopherSocialNetwork/internal/store"
)

type postKey string

const postCtx postKey = "post"

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

	post := store.Post{
		// TODO: Change after auth
		UserID:  1,
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
	post.Comments = comments

	if err := app.JSONResponse(w, http.StatusOK, post); err != nil {
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

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postIdStr := chi.URLParam(r, "postId")
		postId, err := strconv.ParseInt(postIdStr, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()

		post, err := app.store.Posts.GetById(ctx, postId)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	return r.Context().Value(postCtx).(*store.Post)
}
