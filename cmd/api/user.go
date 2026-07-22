package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/loeia/gopherSocialNetwork/internal/store"
)

type userKey string

const userCtx userKey = "userID"

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	if err := app.writeJSON(w, http.StatusCreated, user); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) usersContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.store.Users.GetById(ctx, userId)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}

func getUserFromCtx(r *http.Request) *store.User {
	return r.Context().Value(userCtx).(*store.User)
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

	if err := app.JSONResponse(w, http.StatusCreated, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
