// Package api implements the HTTP layer of the sortie-test service.
package api

import (
	"encoding/json"
	"net/http"
)

// APIError describes a single failure returned to the client.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// errorResponse is the envelope every failing endpoint responds with.
type errorResponse struct {
	Error APIError `json:"error"`
}

// WriteError responds with the standard error envelope.
//
// Handlers must use this instead of writing a bare status code or a plain
// string, so that clients can rely on a single error shape.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Error: APIError{Code: code, Message: message}})
}

// writeJSON encodes body as JSON and responds with the given status.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
