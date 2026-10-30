package controller

import (
	"encoding/json"
	"net/http"
)

// WriteJSON writes a JSON payload with standard content-type and status code.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteError formats an error as a JSON object `{"error": message}`.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}
