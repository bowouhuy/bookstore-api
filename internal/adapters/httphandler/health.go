package httphandler

import (
	"net/http"

	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/usecases"
	"github.com/bowouhuy/bookstore-api/pkg/logger"
)

// healthHandler is a struct for health handler
type healthHandler struct {
	healthUsecase usecases.HealthUsecase
}

// NewHealthHandler is a constructor function for health handler
func NewHealthHandler(
	healthUsecase usecases.HealthUsecase,
) Handler {
	return &healthHandler{
		healthUsecase: healthUsecase,
	}
}

// Handle handles GET /liveness.
func (h *healthHandler) Handle(req *http.Request) dto.HTTPResponse {
	message, err := h.healthUsecase.Liveness(req.Context())
	if err != nil {
		logger.ErrorWithContext(req.Context(), logger.MessageFormat("liveness check failed: %v", err))
		return *dto.NewResponse().WithCode(http.StatusServiceUnavailable).WithMessage("Database is unreachable")
	}

	return *dto.NewResponse().WithCode(http.StatusOK).WithMessage(message)
}
