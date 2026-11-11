package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bowouhuy/bookstore-api/internal/middleware"
)

// fakeVerifier accepts exactly one token.
type fakeVerifier struct{}

func (fakeVerifier) Verify(token string) (int64, error) {
	if token == "valid-token" {
		return 42, nil
	}
	return 0, errors.New("invalid token")
}

func TestAuth(t *testing.T) {
	tests := map[string]struct {
		authHeader string
		wantErr    bool
	}{
		"no header":               {authHeader: "", wantErr: true},
		"wrong scheme":            {authHeader: "Basic dXNlcjpwYXNz", wantErr: true},
		"scheme without token":    {authHeader: "Bearer", wantErr: true},
		"scheme with empty token": {authHeader: "Bearer    ", wantErr: true},
		"invalid token":           {authHeader: "Bearer wrong-token", wantErr: true},
		"valid token":             {authHeader: "Bearer valid-token"},
		"lowercase scheme":        {authHeader: "bearer valid-token"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/books", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			got, err := middleware.Auth(fakeVerifier{})(req)

			if tt.wantErr {
				var mErr *middleware.Error
				require.ErrorAs(t, err, &mErr)
				assert.Equal(t, http.StatusUnauthorized, mErr.Code)
				assert.NotEmpty(t, mErr.Headers["WWW-Authenticate"], "401 must include WWW-Authenticate")
				return
			}

			require.NoError(t, err)
			userID, ok := middleware.UserIDFromContext(got.Context())
			assert.True(t, ok)
			assert.Equal(t, int64(42), userID)
		})
	}
}

func TestFilterFunc_StopsAtFirstErrorAndKeepsOrder(t *testing.T) {
	var order []string
	type key struct{}

	track := func(name string, fail bool) middleware.Handle {
		return func(r *http.Request) (*http.Request, error) {
			order = append(order, name)
			if fail {
				return r, errors.New(name + " failed")
			}
			return r.WithContext(context.WithValue(r.Context(), key{}, name)), nil
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	got, err := middleware.FilterFunc(req, []middleware.Handle{track("first", false), track("second", false)})
	require.NoError(t, err)
	assert.Equal(t, "second", got.Context().Value(key{}), "later middleware must see earlier context changes")

	order = nil
	_, err = middleware.FilterFunc(req, []middleware.Handle{track("first", true), track("second", false)})
	require.Error(t, err)
	assert.Equal(t, "first", strings.Join(order, ","), "middleware after a failure must not run")
}
