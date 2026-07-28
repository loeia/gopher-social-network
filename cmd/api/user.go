package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/loeia/gopherSocialNetwork/internal/mailer"
	"github.com/loeia/gopherSocialNetwork/internal/store"
)

type userKey string

const userCtx userKey = "userID"

type UserWithToken struct {
	User  *store.User `json:"user"`
	Token string      `json:"token"`
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	user, err := app.getUser(r.Context(), userId)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.JSONResponse(w, http.StatusOK, user); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterUserPayload

	if err := app.readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	user := store.User{
		Email:    req.Email,
		Username: req.Username,
		Role: store.Role{
			Name: "user",
		},
	}
	if err := user.Password.Set(req.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	if err := app.store.Users.CreateAndInvite(r.Context(), &user, hashToken, app.config.mail.exp); err != nil {
		switch {
		case errors.Is(err, store.ErrDuplicateUsername), errors.Is(err, store.ErrDuplicateEmail):
			app.conflictError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	uwt := UserWithToken{
		Token: plainToken,
		User:  &user,
	}

	isProdEnv := app.config.env == "production"

	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	// send email
	status, err := app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending welcome email", "error", err)

		if err := app.store.Users.Delete(r.Context(), user.ID); err != nil {
			app.logger.Errorw("error deleting user", "error", err)
		}
		app.internalServerError(w, r, err)
		return
	}
	app.logger.Infow("Email sent", "status code", status)

	if err := app.JSONResponse(w, http.StatusCreated, uwt); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	if err := app.store.Users.Activate(r.Context(), token); err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
