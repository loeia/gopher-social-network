package main

import (
	"log/slog"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("internal server error", "error", err.Error(), "method", r.Method, "url_path", r.URL.Path)

	app.writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("bad request error", "error", err.Error(), "method", r.Method, "url_path", r.URL.Path)

	app.writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) conflictError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("conflict error", "error", err.Error(), "method", r.Method, "url_path", r.URL.Path)

	app.writeJSONError(w, http.StatusConflict, err.Error())
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("not found error", "error", err.Error(), "method", r.Method, "url_path", r.URL.Path)

	app.writeJSONError(w, http.StatusNotFound, "resource not found")
}
