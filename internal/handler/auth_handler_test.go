package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bowouhuy/bookstore-api/internal/auth"
	"github.com/bowouhuy/bookstore-api/internal/model"
)

type fakeAuthService struct {
	user  *model.User
	token *model.AuthToken
	err   error
}

func (f *fakeAuthService) Register(ctx context.Context, input model.RegisterInput) (*model.User, error) {
	return f.user, f.err
}

func (f *fakeAuthService) Login(ctx context.Context, input model.LoginInput) (*model.AuthToken, error) {
	return f.token, f.err
}

func (f *fakeAuthService) GetUser(ctx context.Context, id int64) (*model.User, error) {
	return f.user, f.err
}

func TestAuthHandler_Register(t *testing.T) {
	validBody := `{"name": "Budi", "email": "budi@example.com", "password": "rahasia123"}`

	tests := []struct {
		name       string
		body       string
		service    *fakeAuthService
		wantStatus int
	}{
		{
			name: "success",
			body: validBody,
			service: &fakeAuthService{user: &model.User{
				ID: 1, Name: "Budi", Email: "budi@example.com",
			}},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "email already exists",
			body:       validBody,
			service:    &fakeAuthService{err: model.ErrEmailAlreadyExists},
			wantStatus: http.StatusConflict,
		},
		{
			name: "validation error",
			body: validBody,
			service: &fakeAuthService{
				err: &model.ValidationError{Fields: map[string]string{"password": "must be at least 8 characters"}},
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "client tries to set password_hash",
			body:       `{"name": "Budi", "email": "budi@example.com", "password": "rahasia123", "password_hash": "x"}`,
			service:    &fakeAuthService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAuthHandler(tt.service)

			req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			h.Register(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestAuthHandler_Register_DoesNotLeakPasswordHash(t *testing.T) {
	svc := &fakeAuthService{user: &model.User{
		ID:           1,
		Name:         "Budi",
		Email:        "budi@example.com",
		PasswordHash: "$2a$12$super-secret-hash-value",
	}}
	h := NewAuthHandler(svc)

	body := `{"name": "Budi", "email": "budi@example.com", "password": "rahasia123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	responseBody := rec.Body.String()
	if strings.Contains(responseBody, "super-secret-hash-value") {
		t.Fatalf("response leaks password hash: %s", responseBody)
	}
	if strings.Contains(responseBody, "password") {
		t.Errorf("response should not contain any password field: %s", responseBody)
	}
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &fakeAuthService{token: &model.AuthToken{
			AccessToken: "abc.def.ghi",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(time.Hour),
		}}
		h := NewAuthHandler(svc)

		body := `{"email": "budi@example.com", "password": "rahasia123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, want no-store", got)
		}

		var got model.AuthToken
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if got.AccessToken != "abc.def.ghi" {
			t.Errorf("access_token = %q, want abc.def.ghi", got.AccessToken)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		h := NewAuthHandler(&fakeAuthService{err: model.ErrInvalidCredentials})

		body := `{"email": "budi@example.com", "password": "salah12345"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})
}

func TestAuthHandler_Me(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64 // 0 means no user in context
		service    *fakeAuthService
		wantStatus int
	}{
		{
			name:       "authenticated user",
			userID:     1,
			service:    &fakeAuthService{user: &model.User{ID: 1, Name: "Budi"}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no user in context",
			userID:     0,
			service:    &fakeAuthService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "user deleted after token was issued",
			userID:     1,
			service:    &fakeAuthService{err: model.ErrNotFound},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAuthHandler(tt.service)

			req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
			if tt.userID != 0 {
				ctx := auth.WithUserID(req.Context(), tt.userID)
				req = req.WithContext(ctx)
			}
			rec := httptest.NewRecorder()

			h.Me(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
