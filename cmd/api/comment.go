package main

import (
	"net/http"

	"github.com/loeia/gopherSocialNetwork/internal/store"
)

type CommentKey string

const CommentCtx CommentKey = "commentId"

type CommentResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

func (app *application) getPostCommentsHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	comments, err := app.store.Comments.GetByPostId(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	var responseComments []*CommentResponse
	for _, c := range comments {
		rc := CommentResponse{
			ID:        c.ID,
			Username:  c.User.Username,
			Content:   c.Content,
			CreatedAt: c.CreatedAt,
		}
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

	comment, err := app.store.Comments.Create(r.Context(), &c)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	resp := CommentResponse{
		ID:        comment.ID,
		Username:  user.Username,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}
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

	resp := CommentResponse{
		ID:        comment.ID,
		Username:  comment.User.Username,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}

	if err := app.JSONResponse(w, http.StatusCreated, resp); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
