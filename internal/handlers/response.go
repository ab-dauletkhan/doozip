package handlers

import (
	"encoding/json"
	"net/http"
)

// Response represents a standardized API response.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// WriteJSON writes a successful JSON response.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, "failed to marshal JSON response", http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

// WriteError writes an error JSON response.
func WriteError(w http.ResponseWriter, status int, err string) {
	WriteJSON(w, status, Response{Success: false, Error: err})
}
