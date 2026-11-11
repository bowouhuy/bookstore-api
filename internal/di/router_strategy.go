package di

import (
	"net/http"

	"github.com/bowouhuy/bookstore-api/internal/adapters/httphandler"
	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/pkg/routerkit"
)

// httpHandlerFunc is a contract http handler for router
type httpHandlerFunc func(request *http.Request, handler httphandler.Handler) dto.HTTPResponse

// Router is a contract router and must implement this interface
type Router interface {
	Route() *routerkit.Router
}
