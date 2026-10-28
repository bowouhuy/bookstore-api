package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/bowouhuy/bookstore-api/internal/model"
)

type errorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

// handleServiceError maps service errors to HTTP responses.
// Unknown errors are logged and returned as 500.
func handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var vErr *model.ValidationError

	switch {
	case errors.As(err, &vErr):
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{
			Error:  vErr.Error(),
			Fields: vErr.Fields,
		})
	case errors.Is(err, model.ErrNotFound):
		writeError(w, http.StatusNotFound, "resource not found")
	case errors.Is(err, model.ErrEmailAlreadyExists):
		writeError(w, http.StatusConflict, "email already exists")
	case errors.Is(err, model.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid email or password")
	default:
		slog.Error("internal server error",
			"method", r.Method,
			"path", r.URL.Path,
			"error", err,
		)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
