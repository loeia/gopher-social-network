package main

import (
	"net/http"
)

func (app *application) likePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	user := getUserFromCtx(r)

	if err := app.store.PostLikes.Like(r.Context(), post.ID, user.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) unlikePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	user := getUserFromCtx(r)

	if err := app.store.PostLikes.Dislike(r.Context(), post.ID, user.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
