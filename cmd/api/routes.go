package main

import (
	"net/http"

	"github.com/bowouhuy/bookstore-api/internal/handler"
	"github.com/bowouhuy/bookstore-api/internal/middleware"
)

func newRouter(
	bookHandler *handler.BookHandler,
	authHandler *handler.AuthHandler,
	verifier middleware.TokenVerifier,
) http.Handler {
	mux := http.NewServeMux()

	requireAuth := middleware.RequireAuth(verifier)
	protected := func(h http.HandlerFunc) http.Handler {
		return requireAuth(h)
	}

	// Public routes
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// Protected routes
	mux.Handle("GET /auth/me", protected(authHandler.Me))

	mux.Handle("GET /books", protected(bookHandler.List))
	mux.Handle("POST /books", protected(bookHandler.Create))
	mux.Handle("GET /books/{id}", protected(bookHandler.GetByID))
	mux.Handle("PUT /books/{id}", protected(bookHandler.Update))
	mux.Handle("DELETE /books/{id}", protected(bookHandler.Delete))

	return middleware.Chain(mux,
		middleware.Recover,
		middleware.Logging,
	)
}
