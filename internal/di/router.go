package di

import (
	"net/http"

	"github.com/bowouhuy/bookstore-api/config"
	"github.com/bowouhuy/bookstore-api/pkg/routerkit"
)

type router struct {
	cfg    *config.Config
	router *routerkit.Router
}

func NewRouter() Router {
	cfg := GetConfig()
	return &router{cfg: cfg, router: routerkit.NewRouter(routerkit.WithServiceName(cfg.App.Name))}
}

func (r *router) Route() *routerkit.Router {
	r.router.NotFoundHandler = r.fallback(http.StatusNotFound)
	r.router.MethodNotAllowedHandler = r.fallback(http.StatusMethodNotAllowed)

	root := r.router.PathPrefix("/").Subrouter()
	requireAuth := GetAuthMiddleware()

	root.HandleFunc("/liveness", r.handle(httpGateway, GetHealthHandler())).Methods(http.MethodGet)

	root.HandleFunc("/auth/register", r.handle(httpGateway, GetRegisterHandler())).Methods(http.MethodPost)
	root.HandleFunc("/auth/login", r.handle(httpGateway, GetLoginHandler())).Methods(http.MethodPost)
	root.HandleFunc("/auth/me", r.handle(httpGateway, GetMeHandler(), requireAuth)).Methods(http.MethodGet)

	root.HandleFunc("/books", r.handle(httpGateway, GetListBooksHandler(), requireAuth)).Methods(http.MethodGet)
	root.HandleFunc("/books", r.handle(httpGateway, GetCreateBookHandler(), requireAuth)).Methods(http.MethodPost)
	root.HandleFunc("/books/{id}", r.handle(httpGateway, GetGetBookHandler(), requireAuth)).Methods(http.MethodGet)
	root.HandleFunc("/books/{id}", r.handle(httpGateway, GetUpdateBookHandler(), requireAuth)).Methods(http.MethodPut)
	root.HandleFunc("/books/{id}", r.handle(httpGateway, GetDeleteBookHandler(), requireAuth)).Methods(http.MethodDelete)

	return r.router
}
