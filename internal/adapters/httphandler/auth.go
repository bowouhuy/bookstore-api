package httphandler

import (
	"errors"
	"net/http"

	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/entities"
	"github.com/bowouhuy/bookstore-api/internal/middleware"
	"github.com/bowouhuy/bookstore-api/internal/usecases"
)

type registerHandler struct {
	authUsecase usecases.AuthUsecase
}

func NewRegisterHandler(authUsecase usecases.AuthUsecase) Handler {
	return &registerHandler{authUsecase: authUsecase}
}

// Handle handles POST /auth/register.
func (h *registerHandler) Handle(req *http.Request) dto.HTTPResponse {
	var input dto.RegisterRequest
	if err := decodeJSON(req, &input); err != nil {
		return decodeErrorResponse(err)
	}

	user, err := h.authUsecase.Register(req.Context(), input)
	if err != nil {
		return errorResponse(req, err)
	}

	return *dto.NewResponse().WithCode(http.StatusCreated).WithData(user)
}

type loginHandler struct {
	authUsecase usecases.AuthUsecase
}

func NewLoginHandler(authUsecase usecases.AuthUsecase) Handler {
	return &loginHandler{authUsecase: authUsecase}
}

// Handle handles POST /auth/login.
func (h *loginHandler) Handle(req *http.Request) dto.HTTPResponse {
	var input dto.LoginRequest
	if err := decodeJSON(req, &input); err != nil {
		return decodeErrorResponse(err)
	}

	token, err := h.authUsecase.Login(req.Context(), input)
	if err != nil {
		return errorResponse(req, err)
	}

	// Token responses must not be cached (RFC 6749 section 5.1).
	return *dto.NewResponse().
		WithCode(http.StatusOK).
		WithHeader("Cache-Control", "no-store").
		WithData(token)
}

type meHandler struct {
	authUsecase usecases.AuthUsecase
}

func NewMeHandler(authUsecase usecases.AuthUsecase) Handler {
	return &meHandler{authUsecase: authUsecase}
}

// Handle handles GET /auth/me. The route must use middleware.Auth.
func (h *meHandler) Handle(req *http.Request) dto.HTTPResponse {
	userID, ok := middleware.UserIDFromContext(req.Context())
	if !ok {
		return *dto.NewResponse().WithCode(http.StatusUnauthorized)
	}

	user, err := h.authUsecase.GetUser(req.Context(), userID)
	if err != nil {
		// A valid token for a deleted user is treated as unauthenticated.
		if errors.Is(err, entities.ErrNotFound) {
			return *dto.NewResponse().WithCode(http.StatusUnauthorized)
		}
		return errorResponse(req, err)
	}

	return *dto.NewResponse().WithCode(http.StatusOK).WithData(user)
}
