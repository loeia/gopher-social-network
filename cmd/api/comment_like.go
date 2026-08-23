package main

import "net/http"

func (app *application) likeCommentHandler(w http.ResponseWriter,r *http.Request) {
	user := getUserFromCtx(r)	
	comment := getCommentFromCtx(r)

	if err := app.store.CommentLike.Like(r.Context(),comment.ID,user.ID) ; err != nil {
		app.internalServerError(w,r,err)
		return 
	}

	w.WriteHeader(http.StatusNoContent)
}

func(app *application) dislikeCommentHandler(w http.ResponseWriter,r *http.Request) {
	user := getUserFromCtx(r)	
	comment := getCommentFromCtx(r)

	if err := app.store.CommentLike.Dislike(r.Context(),comment.ID,user.ID); err != nil {
		app.internalServerError(w,r,err)
		return 
	}

	w.WriteHeader(http.StatusNoContent)
}

