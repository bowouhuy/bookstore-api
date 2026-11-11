package httphandler

import (
	"net/http"

	"github.com/bowouhuy/bookstore-api/internal/dto"
)

// Handler is an interface for http handler
type Handler interface {
	Handle(data *http.Request) dto.HTTPResponse
}
