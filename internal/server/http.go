package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bowouhuy/bookstore-api/internal/consts"
	"github.com/bowouhuy/bookstore-api/internal/di"
	pkgdi "github.com/bowouhuy/bookstore-api/pkg/di"
	"github.com/bowouhuy/bookstore-api/pkg/logger"
)

const shutdownTimeout = 10 * time.Second

type httpServer struct {
	srv *http.Server
}

func NewHTTPServer() Server {
	return &httpServer{}
}

// Run serves until SIGINT/SIGTERM, then drains in-flight requests and closes registered resources.
func (h *httpServer) Run() error {
	cfg := di.GetConfig()
	router := di.NewRouter()
	defer pkgdi.CloseAll()

	h.srv = &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", cfg.App.Port),
		Handler:           router.Route(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       time.Duration(cfg.App.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.App.WriteTimeout) * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- h.srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	case <-ctx.Done():
	}

	logger.Info("shutting down server", logger.EventName(consts.LogEventNameServiceTerminated))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := h.srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	logger.Info("server exited properly", logger.EventName(consts.LogEventNameServiceTerminated))

	return nil
}

func (h *httpServer) GracefulStop() {
	if h.srv != nil {
		_ = h.srv.Shutdown(context.Background())
	}
}
