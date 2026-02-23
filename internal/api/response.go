package api

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Message string `json:"message"`
	Status  int
}

func newError(message string, status int) *Error {
	return &Error{message, status}
}

func writeJSON(w http.ResponseWriter, data any, statusCode int) {
	b, err := json.Marshal(data)

	if err != nil {
		writeError(w, newError("error while writing json response body", http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(b)
}

func writeError(w http.ResponseWriter, apiError *Error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(apiError.Status)

	json.NewEncoder(w).Encode(map[string]string{"error": apiError.Message})
}
