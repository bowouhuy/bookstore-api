package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/bowouhuy/bookstore-api/internal/consts"
)

type TokenVerifier interface {
	Verify(token string) (int64, error)
}

// Auth rejects requests without a valid Bearer token
// and stores the user ID in the request context under consts.CtxUserID.
func Auth(verifier TokenVerifier) Handle {
	return func(r *http.Request) (*http.Request, error) {
		token, ok := bearerToken(r)
		if !ok {
			return r, unauthorized("missing or malformed authorization header")
		}

		userID, err := verifier.Verify(token)
		if err != nil {
			return r, unauthorized("invalid or expired token")
		}

		return r.WithContext(context.WithValue(r.Context(), consts.CtxUserID, userID)), nil
	}
}

// UserIDFromContext returns the user ID stored by Auth.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(consts.CtxUserID).(int64)
	return userID, ok
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get(consts.HeaderAuthorization)

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
func unauthorized(message string) *Error {
	return &Error{
		Code:    http.StatusUnauthorized,
		Message: message,
		Headers: map[string]string{"WWW-Authenticate": `Bearer realm="bookstore-api"`},
	}
}
