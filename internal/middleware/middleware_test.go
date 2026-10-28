package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bowouhuy/bookstore-api/internal/auth"
)

// fakeVerifier accepts exactly one token.
type fakeVerifier struct{}

func (fakeVerifier) Verify(token string) (int64, error) {
	if token == "valid-token" {
		return 42, nil
	}
	return 0, errors.New("invalid token")
}

func TestRequireAuth(t *testing.T) {
	tests := []struct {
		name         string
		authHeader   string
		wantStatus   int
		wantNextCall bool
	}{
		{name: "no header", authHeader: "", wantStatus: http.StatusUnauthorized},
		{name: "wrong scheme", authHeader: "Basic dXNlcjpwYXNz", wantStatus: http.StatusUnauthorized},
		{name: "scheme without token", authHeader: "Bearer", wantStatus: http.StatusUnauthorized},
		{name: "scheme with empty token", authHeader: "Bearer    ", wantStatus: http.StatusUnauthorized},
		{name: "invalid token", authHeader: "Bearer wrong-token", wantStatus: http.StatusUnauthorized},
		{name: "valid token", authHeader: "Bearer valid-token", wantStatus: http.StatusOK, wantNextCall: true},
		{name: "lowercase scheme", authHeader: "bearer valid-token", wantStatus: http.StatusOK, wantNextCall: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			var gotUserID int64

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				gotUserID, _ = auth.UserIDFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			h := RequireAuth(fakeVerifier{})(next)

			req := httptest.NewRequest(http.MethodGet, "/books", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if nextCalled != tt.wantNextCall {
				t.Errorf("next called = %v, want %v", nextCalled, tt.wantNextCall)
			}
			if tt.wantNextCall && gotUserID != 42 {
				t.Errorf("user ID in context = %d, want 42", gotUserID)
			}
			if !tt.wantNextCall && rec.Header().Get("WWW-Authenticate") == "" {
				t.Error("401 response must include WWW-Authenticate header")
			}
		})
	}
}

func TestRecover(t *testing.T) {
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m map[string]int
		m["boom"] = 1 // writing to a nil map panics
	})

	h := Recover(panicking)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rec.Body.String(), "internal server error") {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestLogging_RecordsStatus(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	h := Logging(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want %d (logging must not change the response)", rec.Code, http.StatusTeapot)
	}
}

func TestChain_Order(t *testing.T) {
	var order []string

	track := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	h := Chain(final, track("first"), track("second"), track("third"))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	got := strings.Join(order, " → ")
	want := "first → second → third → handler"
	if got != want {
		t.Errorf("order = %q, want %q", got, want)
	}
}
