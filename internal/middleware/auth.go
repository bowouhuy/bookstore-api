package middleware

import (
	"net/http"
	"strings"

	"github.com/bowouhuy/bookstore-api/internal/auth"
)

type TokenVerifier interface {
	Verify(token string) (int64, error)
}

// RequireAuth rejects requests without a valid Bearer token
// and stores the user ID in the request context.
func RequireAuth(verifier TokenVerifier) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				unauthorized(w, "missing or malformed authorization header")
				return
			}

			userID, err := verifier.Verify(token)
			if err != nil {
				unauthorized(w, "invalid or expired token")
				return
			}

			ctx := auth.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")

	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", false
	}

	return token, true
}

// unauthorized sets WWW-Authenticate, which RFC 7235 requires on 401 responses.
func unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="bookstore-api"`)
	writeJSONError(w, http.StatusUnauthorized, message)
}
