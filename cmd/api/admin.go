package main

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/loeia/gopherSocialNetwork/internal/mailer"
	"github.com/loeia/gopherSocialNetwork/internal/store"
)

func (app *application) adminDeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := app.store.Users.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	isProdEnv := app.config.env == "production"

	vars := struct {
		Username string
	}{
		Username: user.Username,
	}

	status, err := app.mailer.Send(mailer.AccountBanTemplate, user.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending ban email", "error", err)
		app.internalServerError(w, r, err)
		return
	}
	app.logger.Infow("Ban email sent", "status code", status)

	if err := app.store.Users.BanByUsername(r.Context(), username); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.invalidateUserCache(r, user.ID)

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) adminUnbanUserHandler(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := app.store.Users.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	isProdEnv := app.config.env == "production"

	vars := struct {
		Username string
	}{
		Username: user.Username,
	}

	status, err := app.mailer.Send(mailer.AccountUnbanTemplate, user.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending unban email", "error", err)
		app.internalServerError(w, r, err)
		return
	}
	app.logger.Infow("Unban email sent", "status code", status)

	if err := app.store.Users.UnbanByUsername(r.Context(), username); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.invalidateUserCache(r, user.ID)

	w.WriteHeader(http.StatusNoContent)
}
