package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/bowouhuy/bookstore-api/internal/auth"
	"github.com/bowouhuy/bookstore-api/internal/model"
)

type AuthService interface {
	Register(ctx context.Context, input model.RegisterInput) (*model.User, error)
	Login(ctx context.Context, input model.LoginInput) (*model.AuthToken, error)
	GetUser(ctx context.Context, id int64) (*model.User, error)
}

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input model.RegisterInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeDecodeError(w, err)
		return
	}

	user, err := h.service.Register(r.Context(), input)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input model.LoginInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeDecodeError(w, err)
		return
	}

	token, err := h.service.Login(r.Context(), input)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, token)
}

// Me must be wrapped with middleware.RequireAuth, which puts the user ID in the context.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}
