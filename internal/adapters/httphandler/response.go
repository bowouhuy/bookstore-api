package httphandler

import (
	"errors"
	"net/http"

	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/entities"
	"github.com/bowouhuy/bookstore-api/pkg/logger"
)

func badRequest(message string) dto.HTTPResponse {
	return *dto.NewResponse().WithCode(http.StatusBadRequest).WithMessage(message)
}

func decodeErrorResponse(err error) dto.HTTPResponse {
	if errors.Is(err, errBodyTooLarge) {
		return *dto.NewResponse().WithCode(http.StatusRequestEntityTooLarge)
	}

	return badRequest(err.Error())
}

// errorResponse maps usecase errors to HTTP responses.
// Unknown errors are logged and returned as 500 without details.
func errorResponse(req *http.Request, err error) dto.HTTPResponse {
	var vErr *entities.ValidationError

	switch {
	case errors.As(err, &vErr):
		return *dto.NewResponse().WithCode(http.StatusUnprocessableEntity).WithError(vErr.Fields)
	case errors.Is(err, entities.ErrNotFound):
		return *dto.NewResponse().WithCode(http.StatusNotFound)
	case errors.Is(err, entities.ErrEmailAlreadyExists):
		return *dto.NewResponse().WithCode(http.StatusConflict).WithMessage("email already exists")
	case errors.Is(err, entities.ErrInvalidCredentials):
		return *dto.NewResponse().WithCode(http.StatusUnauthorized).WithMessage("invalid email or password")
	default:
		logger.ErrorWithContext(req.Context(), logger.MessageFormat("internal server error: %v", err))
		return *dto.NewResponse().WithCode(http.StatusInternalServerError)
	}
}
