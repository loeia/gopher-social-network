package main

import (
	"errors"
	"net/http"

	"github.com/loeia/gopherSocialNetwork/internal/store"
)

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	var followUser FollowUser
	if err := app.readJSON(w, r, &followUser); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := app.store.Followers.Follow(r.Context(), followUser.UserID, user.ID); err != nil {
		if errors.Is(err, store.ErrConflict) {
			app.conflictError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	// TODO: Revert back to auth userID from ctx
	var unfollowUser FollowUser
	if err := app.readJSON(w, r, &unfollowUser); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := app.store.Followers.Unfollow(r.Context(), unfollowUser.UserID, user.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
