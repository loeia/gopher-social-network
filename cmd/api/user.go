package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/loeia/gopherSocialNetwork/internal/store"
)

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	user, err := app.store.Users.GetById(r.Context(), userId)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeJSON(w, http.StatusCreated, user); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}
}
