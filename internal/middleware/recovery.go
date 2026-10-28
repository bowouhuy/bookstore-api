package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			// ErrAbortHandler is a deliberate abort; let net/http handle it.
			if rec == http.ErrAbortHandler {
				panic(rec)
			}

			slog.Error("panic recovered",
				"panic", rec,
				"method", r.Method,
				"path", r.URL.Path,
				"stack", string(debug.Stack()),
			)

			w.Header().Set("Connection", "close")
			writeJSONError(w, http.StatusInternalServerError, "internal server error")
		}()

		next.ServeHTTP(w, r)
	})
}
