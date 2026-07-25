package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error", "error", err.Error(), "method", r.Method, "url_path", r.URL.Path)

	app.writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("bad request error", "error", err.Error(), "method", r.Method, "url_path", r.URL.Path)

	app.writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) conflictError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("conflict error", "error", err.Error(), "method", r.Method, "url_path", r.URL.Path)

	app.writeJSONError(w, http.StatusConflict, err.Error())
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("not found error", "error", err.Error(), "method", r.Method, "url_path", r.URL.Path)

	app.writeJSONError(w, http.StatusNotFound, "resource not found")
}

func (app *application) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	app.writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}
