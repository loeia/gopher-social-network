package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/loeia/gopherSocialNetwork/internal/store"
)

type CommentKey string

const CommentCtx CommentKey = "commentId"

type CommentResponse struct {
	ID              int64  `json:"id"`
	Username        string `json:"username"`
	UserID          int64  `json:"user_id"`
	Content         string `json:"content"`
	ParentID        *int64 `json:"parent_id"`
	ReplyToUserID   *int64 `json:"reply_to_user_id"`
	ReplyToUsername string `json:"reply_to_username"`
	LikeCount int64 `json:"like_count"`
	CreatedAt       string `json:"created_at"`
}

func commentResponse(c *store.Comment) CommentResponse {
	return CommentResponse{
		ID:              c.ID,
		Username:        c.User.Username,
		UserID:          c.UserID,
		Content:         c.Content,
		ParentID:        c.ParentID,
		ReplyToUserID:   c.ReplyToUserID,
		ReplyToUsername: c.ReplyToUsername,
		LikeCount: c.LikeCount,
		CreatedAt:       c.CreatedAt,
	}
}

func (app *application) getPostCommentsHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	comments, err := app.store.Comments.GetByPostId(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	responseComments := make([]*CommentResponse, 0, len(comments))
	for _, c := range comments {
		rc := commentResponse(c)
		responseComments = append(responseComments, &rc)
	}

	if err := app.JSONResponse(w, http.StatusOK, responseComments); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	user := getUserFromCtx(r)

	var req CommentReq
	if err := app.readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	c := store.Comment{
		PostID:  post.ID,
		UserID:  user.ID,
		Content: req.Content,
	}

	if raw := chi.URLParam(r, "commentId"); raw != "" {
		parentID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			app.badRequestError(w, r, err)
			return
		}
		req.ParentID = &parentID
	}

	if req.ParentID != nil {
		parentCmt, err := app.store.Comments.GetById(r.Context(), *req.ParentID)
		if err != nil {
			app.badRequestError(w, r, store.ErrNotFound)
			return
		}
		if parentCmt.PostID != post.ID {
			app.badRequestError(w, r, store.ErrNotFound)
			return
		}
		c.ParentID = &parentCmt.ID
		c.ReplyToUserID = &parentCmt.UserID
	}

	comment, err := app.store.Comments.Create(r.Context(), &c)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	comment.User.Username = user.Username

	resp := commentResponse(comment)
	if err := app.JSONResponse(w, http.StatusCreated, resp); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	comment := getCommentFromCtx(r)

	if err := app.store.Comments.Delete(r.Context(), comment.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) getCommentHandler(w http.ResponseWriter, r *http.Request) {
	comment := getCommentFromCtx(r)

	resp := commentResponse(comment)
	if err := app.JSONResponse(w, http.StatusOK, resp); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
