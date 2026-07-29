package main

import (
	"net/http"
)

// internalServerError Logs an internal server error and returns a 500 status code.
func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error", "error", err.Error(), "method", r.Method, "url_path", r.URL.Path)

	app.writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

// badRequestError Logs a bad request error and returns a 400 status code.
func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("bad request error", "error", err.Error(), "method", r.Method, "url_path", r.URL.Path)

	app.writeJSONError(w, http.StatusBadRequest, err.Error())
}

// conflictError Logs a conflict error and returns a 409 status code.
func (app *application) conflictError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("conflict error", "error", err.Error(), "method", r.Method, "url_path", r.URL.Path)

	app.writeJSONError(w, http.StatusConflict, err.Error())
}

// notFoundError Logs a not found error and returns a 404 status code.
func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("not found error", "error", err.Error(), "method", r.Method, "url_path", r.URL.Path)

	app.writeJSONError(w, http.StatusNotFound, "resource not found")
}

// unauthorizedErrorResponse Logs an unauthorized error and returns a 401 status code.
func (app *application) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	app.writeJSONError(w, http.StatusUnauthorized, "unauthorized")
}

// forbiddenResponse Logs a forbidden access and returns a 403 status code.
func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	app.logger.Warnw("forbidden", "method", r.Method, "path", r.URL.Path)

	app.writeJSONError(w, http.StatusForbidden, "forbidden")
}

// rateLimitExceededResponse Logs a rate limit exceeded and returns a 429 status code.
func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path)

	w.Header().Set("Retry-After", retryAfter)

	app.writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter)
}
