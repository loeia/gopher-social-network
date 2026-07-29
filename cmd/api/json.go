package main

import (
	"encoding/json"
	"net/http"
)

// writeJSON Sends a JSON response with the given status code and data.
func (app *application) writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

// readJSON Decodes a JSON request body into the provided data structure.
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_578 // 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

// writeJSONError Sends a JSON error response with the given status code and message.
func (app *application) writeJSONError(w http.ResponseWriter, status int, err string) error {
	type envelope struct {
		Error string `json:"error"`
	}

	return app.writeJSON(w, status, &envelope{Error: err})
}

// JSONResponse Sends a JSON response wrapped in a "data" envelope.
func (app *application) JSONResponse(w http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}

	return app.writeJSON(w, status, &envelope{Data: data})
}
