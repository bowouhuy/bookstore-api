package httphandler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bowouhuy/bookstore-api/internal/adapters/httphandler"
	"github.com/bowouhuy/bookstore-api/internal/consts"
	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/entities"
	"github.com/bowouhuy/bookstore-api/internal/usecases"
)

const validRegisterBody = `{"name": "Budi", "email": "budi@example.com", "password": "rahasia123"}`

func TestRegisterHandler(t *testing.T) {
	tests := map[string]struct {
		body     string
		setup    func(uc *usecases.GoMockAuthUsecase)
		wantCode int
	}{
		"success": {
			body: validRegisterBody,
			setup: func(uc *usecases.GoMockAuthUsecase) {
				uc.EXPECT().Register(gomock.Any(), gomock.Any()).Return(&entities.User{ID: 1, Name: "Budi"}, nil)
			},
			wantCode: http.StatusCreated,
		},
		"email already exists": {
			body: validRegisterBody,
			setup: func(uc *usecases.GoMockAuthUsecase) {
				uc.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil, entities.ErrEmailAlreadyExists)
			},
			wantCode: http.StatusConflict,
		},
		"validation error": {
			body: validRegisterBody,
			setup: func(uc *usecases.GoMockAuthUsecase) {
				uc.EXPECT().Register(gomock.Any(), gomock.Any()).
					Return(nil, &entities.ValidationError{Fields: map[string]string{"password": "must be at least 8 characters"}})
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		"client tries to set password_hash": {
			body:     `{"name": "Budi", "email": "budi@example.com", "password": "rahasia123", "password_hash": "x"}`,
			setup:    func(uc *usecases.GoMockAuthUsecase) {},
			wantCode: http.StatusBadRequest,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			uc := usecases.NewGoMockAuthUsecase(gomock.NewController(t))
			tt.setup(uc)

			req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tt.body))
			resp := httphandler.NewRegisterHandler(uc).Handle(req)

			assert.Equal(t, tt.wantCode, resp.Code)
		})
	}
}

func TestRegisterHandler_DoesNotLeakPasswordHash(t *testing.T) {
	uc := usecases.NewGoMockAuthUsecase(gomock.NewController(t))
	uc.EXPECT().Register(gomock.Any(), gomock.Any()).Return(&entities.User{
		ID:           1,
		Name:         "Budi",
		Email:        "budi@example.com",
		PasswordHash: "$2a$12$super-secret-hash-value",
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(validRegisterBody))
	resp := httphandler.NewRegisterHandler(uc).Handle(req)

	body, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "super-secret-hash-value")
	assert.NotContains(t, string(body), "password")
}

func TestLoginHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		token := &dto.AuthTokenResponse{AccessToken: "abc.def.ghi", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}

		uc := usecases.NewGoMockAuthUsecase(gomock.NewController(t))
		uc.EXPECT().Login(gomock.Any(), dto.LoginRequest{Email: "budi@example.com", Password: "rahasia123"}).Return(token, nil)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email": "budi@example.com", "password": "rahasia123"}`))
		resp := httphandler.NewLoginHandler(uc).Handle(req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "no-store", resp.Headers()["Cache-Control"])
		assert.Equal(t, token, resp.Data)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		uc := usecases.NewGoMockAuthUsecase(gomock.NewController(t))
		uc.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, entities.ErrInvalidCredentials)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email": "budi@example.com", "password": "salah12345"}`))
		resp := httphandler.NewLoginHandler(uc).Handle(req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}

func TestMeHandler(t *testing.T) {
	tests := map[string]struct {
		userID   int64 // 0 means no user in context
		setup    func(uc *usecases.GoMockAuthUsecase)
		wantCode int
	}{
		"authenticated user": {
			userID: 1,
			setup: func(uc *usecases.GoMockAuthUsecase) {
				uc.EXPECT().GetUser(gomock.Any(), int64(1)).Return(&entities.User{ID: 1, Name: "Budi"}, nil)
			},
			wantCode: http.StatusOK,
		},
		"no user in context": {
			setup:    func(uc *usecases.GoMockAuthUsecase) {},
			wantCode: http.StatusUnauthorized,
		},
		"user deleted after token was issued": {
			userID: 1,
			setup: func(uc *usecases.GoMockAuthUsecase) {
				uc.EXPECT().GetUser(gomock.Any(), int64(1)).Return(nil, entities.ErrNotFound)
			},
			wantCode: http.StatusUnauthorized,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			uc := usecases.NewGoMockAuthUsecase(gomock.NewController(t))
			tt.setup(uc)

			req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
			if tt.userID != 0 {
				req = req.WithContext(context.WithValue(req.Context(), consts.CtxUserID, tt.userID))
			}

			resp := httphandler.NewMeHandler(uc).Handle(req)

			assert.Equal(t, tt.wantCode, resp.Code)
		})
	}
}
