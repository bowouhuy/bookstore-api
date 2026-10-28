package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/bowouhuy/bookstore-api/internal/auth"
	"github.com/bowouhuy/bookstore-api/internal/config"
	"github.com/bowouhuy/bookstore-api/internal/database"
	"github.com/bowouhuy/bookstore-api/internal/handler"
	"github.com/bowouhuy/bookstore-api/internal/repository"
	"github.com/bowouhuy/bookstore-api/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	slog.Info("config loaded",
		"env", cfg.AppEnv,
		"port", cfg.HTTPPort,
		"jwt_ttl", cfg.JWTTTL.String(),
	)

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	slog.Info("database connected", "max_conns", pool.Config().MaxConns)

	bookRepo := repository.NewBookRepository(pool)
	userRepo := repository.NewUserRepository(pool)

	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTTTL)

	bookService := service.NewBookService(bookRepo)
	authService := service.NewAuthService(userRepo, tokenManager)

	bookHandler := handler.NewBookHandler(bookService)
	authHandler := handler.NewAuthHandler(authService)

	router := newRouter(bookHandler, authHandler, tokenManager)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	slog.Info("server starting", "addr", srv.Addr)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}
