package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/loeia/gopherSocialNetwork/internal/mailer"
	"github.com/loeia/gopherSocialNetwork/internal/store"
)

type userKey string

const userCtx userKey = "userId"

const maxAvatarSize = 2 << 20 // 2 MB
const maxAvatarDim = 2000

type UserWithToken struct {
	User  *store.User `json:"user"`
	Token string      `json:"token"`
}

type PublicUser struct {
	ID             int64     `json:"id"`
	Username       string    `json:"username"`
	AvatarURL      string    `json:"avatar_url"`
	Bio            string    `json:"bio"`
	Links          []string  `json:"links"`
	CreatedAt      time.Time `json:"created_at"`
	FollowersCount int64     `json:"followers_count"`
	FollowingCount int64     `json:"following_count"`
	PostsCount     int64     `json:"posts_count"`
	LikesCount     int64     `json:"likes_count"`
	RepliesCount   int64     `json:"replies_count"`
}

func userResponse(u *store.User) PublicUser {
	return PublicUser{
		ID:             u.ID,
		Username:       u.Username,
		AvatarURL:      u.AvatarURL,
		Bio:            u.Bio,
		Links:          u.Links,
		CreatedAt:      u.CreatedAt,
		FollowersCount: u.FollowersCount,
		FollowingCount: u.FollowingCount,
		PostsCount:     u.PostsCount,
		LikesCount:     u.LikesCount,
		RepliesCount:   u.RepliesCount,
	}
}

// invalidateUserCache deletes a user's cached profile so the next read
// refetches fresh data (e.g. after post/comment/like counts change).
func (app *application) invalidateUserCache(r *http.Request, userId int64) {
	if !app.config.redisCfg.enabled {
		return
	}
	if err := app.cache.Delete(r.Context(), userId); err != nil {
		app.logger.Errorw("error deleting user from cache", "error", err)
	}
}

// getUserHandler returns a user by ID.
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

	if err := app.JSONResponse(w, http.StatusOK, userResponse(user)); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			app.notFoundError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}
}

// registerUserHandler handles user registration and sends an activation email.
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

// activateUserHandler activates a user account using a token from the URL.
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

// resetUserPasswordHandler reset user password, permission authentication is required
func (app *application) resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	var req ResetPassword
	if err := app.readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(&req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	dbUser, err := app.store.Users.GetById(r.Context(), user.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := dbUser.Password.Compare(req.OldPassword); err != nil {
		app.unauthorizedErrorResponse(w, r, err)
		return
	}

	if err := app.store.Users.UpdatePassword(r.Context(), req.NewPassword, user.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.invalidateUserCache(r, user.ID)

	w.WriteHeader(http.StatusOK)
}

func (app *application) getUserFollowersHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	pq := store.PaginationQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}
	p, err := pq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(p); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	followers, err := app.store.Users.GetUserFollowers(r.Context(), userId, p)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusOK, followers); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getUserFollowingHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	pq := store.PaginationQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}
	p, err := pq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(p); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	following, err := app.store.Users.GetUserFollowing(r.Context(), userId, p)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusOK, following); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

// uploadAvatarHandler processes an avatar image upload for the authenticated user.
func (app *application) uploadAvatarHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize+64<<10)
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		app.unsupportedMediaTypeResponse(w, r)
		return
	}
	if err := r.ParseMultipartForm(maxAvatarSize >> 2); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			app.requestEntityTooLargeResponse(w, r)
			return
		}
		app.badRequestError(w, r, err)
		return
	}

	defer r.MultipartForm.RemoveAll()

	file, _, err := r.FormFile("avatar")
	if err != nil {
		app.badRequestError(w, r, fmt.Errorf("avatar file is required"))
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxAvatarSize+1))
	if err != nil {
		app.requestEntityTooLargeResponse(w, r)
		return
	}
	if len(data) > maxAvatarSize {
		app.requestEntityTooLargeResponse(w, r)
		return
	}

	switch http.DetectContentType(data) {
	case "image/jpeg", "image/png":
	default:
		app.unsupportedMediaTypeResponse(w, r)
		return
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		app.badRequestError(w, r, fmt.Errorf("invalid image"))
		return
	}
	if cfg.Width > maxAvatarDim || cfg.Height > maxAvatarDim {
		app.requestEntityTooLargeResponse(w, r)
		return
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		app.badRequestError(w, r, fmt.Errorf("invalid image"))
		return
	}

	var buf bytes.Buffer
	if format == "png" {
		if err := png.Encode(&buf, img); err != nil {
			app.internalServerError(w, r, err)
			return
		}
	} else {
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
			app.internalServerError(w, r, err)
			return
		}
		format = "jpeg"
	}

	if err := app.store.Users.UpdateAvatar(r.Context(), user.ID, buf.Bytes(), "image/"+format); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.invalidateUserCache(r, user.ID)

	w.WriteHeader(http.StatusNoContent)
}

// getAvatarHandler returns the avatar image bytes for a user.
func (app *application) getAvatarHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	data, mime, err := app.store.Users.GetAvatar(r.Context(), userId)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			w.WriteHeader(http.StatusNoContent)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.Header().Set("Content-Type", mime)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(data)
}

func (app *application) updateProfileHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	var req UpdateProfilePayload
	if err := app.readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(&req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := app.store.Users.UpdateProfile(r.Context(), user.ID, req.Bio, req.Links); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.invalidateUserCache(r, user.ID)

	w.WriteHeader(http.StatusNoContent)
}

// forgetPassHandler handles the forgot password request and sends a reset email.
func (app *application) forgetPassHandler(w http.ResponseWriter, r *http.Request) {
	var req ForgetPasswordPayload
	if err := app.readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	user, err := app.store.Users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.JSONResponse(w, http.StatusOK, map[string]string{"message": "if that email exists, a reset link has been sent"})
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	if err := app.store.Users.CreatePasswordReset(r.Context(), hashToken, user.ID, app.config.mail.resetExp); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	isProdEnv := app.config.env == "production"

	resetURL := fmt.Sprintf("%s/reset-password/%s", app.config.frontendURL, plainToken)
	vars := struct {
		Username string
		ResetURL string
	}{
		Username: user.Username,
		ResetURL: resetURL,
	}

	status, err := app.mailer.Send(mailer.PasswordResetTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending password reset email", "error", err)
		app.internalServerError(w, r, err)
		return
	}
	app.logger.Infow("Password reset email sent", "status code", status)

	app.JSONResponse(w, http.StatusOK, map[string]string{"message": "if that email exists, a reset link has been sent"})
}

// resetPasswordFromTokenHandler handles the password reset using a token (no auth required).
func (app *application) resetPasswordFromTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordPayload
	if err := app.readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := app.store.Users.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		switch err {
		case store.ErrNotFound:
			app.badRequestError(w, r, fmt.Errorf("invalid or expired token"))
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	app.JSONResponse(w, http.StatusOK, map[string]string{"message": "password has been reset successfully"})
}

func (app *application) getUserCommentLikesHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	pq := store.PaginationQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}
	p, err := pq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(p); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	comments, err := app.store.CommentLikes.GetUserFavoriteComments(r.Context(), user.ID, p)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusOK, comments); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// userRenameHandler user rename interface requires authentication
func (app *application) userRenameHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	var req RenamePayload
	if err := app.readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(&req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if user.Username == req.NewName {
		app.badRequestError(w, r, fmt.Errorf("new username is the same as the current username"))
		return
	}

	if err := app.store.Users.Rename(r.Context(), user.ID, req.NewName); err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundError(w, r, err)
			return
		case store.ErrDuplicateUsername:
			app.conflictError(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	app.invalidateUserCache(r, user.ID)

	w.WriteHeader(http.StatusNoContent)
}

// getUserPostsHandler returns posts by a specific user
func (app *application) getUserPostsHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	pq := store.PaginationQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}
	p, err := pq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(p); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	posts, err := app.store.Users.GetUserOwnPosts(r.Context(), userId, p)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	var resp []*PostResponse
	for i := range posts {
		p := postResponse(posts[i])
		resp = append(resp, p)
	}

	if err := app.JSONResponse(w, http.StatusOK, resp); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// getUserLikedPostsHandler returns posts liked by a specific user
func (app *application) getUserLikedPostsHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	pq := store.PaginationQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}
	p, err := pq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(p); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	posts, err := app.store.PostLikes.GetUserFavoritePosts(r.Context(), userId, p)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusOK, posts); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// getUserCommentsHandler returns comments by a specific user
func (app *application) getUserCommentsHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	pq := store.PaginationQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}
	p, err := pq.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(p); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	comments, err := app.store.Comments.GetUserComments(r.Context(), userId, p)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.JSONResponse(w, http.StatusOK, comments); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
