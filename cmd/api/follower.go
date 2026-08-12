package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/loeia/gopherSocialNetwork/internal/store"
)

// followUserHandler handles following another user.
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	followUserId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if followUserId == user.ID {
		app.badRequestError(w, r, errors.New("cannot follow yourself"))
		return
	}

	if err := app.store.Followers.Follow(r.Context(), followUserId, user.ID); err != nil {
		if errors.Is(err, store.ErrConflict) {
			app.conflictError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// unfollowUserHandler handles unfollowing another user.
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	unfollowUserId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.store.Followers.Unfollow(r.Context(), unfollowUserId, user.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
