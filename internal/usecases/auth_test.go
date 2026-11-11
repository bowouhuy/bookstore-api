package usecases_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/entities"
	"github.com/bowouhuy/bookstore-api/internal/entities/repository"
	"github.com/bowouhuy/bookstore-api/internal/usecases"
	"github.com/bowouhuy/bookstore-api/pkg/auth"
)

type authFields struct {
	userRepo *repository.GoMockUserRepo
	tokens   *usecases.GoMockTokenGenerator
}

func newAuthSUT(t *testing.T) (usecases.AuthUsecase, authFields) {
	ctrl := gomock.NewController(t)
	f := authFields{
		userRepo: repository.NewGoMockUserRepo(ctrl),
		tokens:   usecases.NewGoMockTokenGenerator(ctrl),
	}

	return usecases.NewAuthUsecase(f.userRepo, f.tokens), f
}

func TestAuthUC_Register_NormalizesAndHashes(t *testing.T) {
	sut, f := newAuthSUT(t)
	ctx := context.Background()

	var storedHash string
	f.userRepo.EXPECT().
		Create(ctx, "Budi Santoso", "budi@example.com", gomock.Any()).
		DoAndReturn(func(_ context.Context, name, email, hash string) (*entities.User, error) {
			storedHash = hash
			return &entities.User{ID: 1, Name: name, Email: email, PasswordHash: hash}, nil
		})

	_, err := sut.Register(ctx, dto.RegisterRequest{
		Name:     "  Budi Santoso  ",
		Email:    "  Budi@Example.COM ",
		Password: "rahasia123",
	})

	require.NoError(t, err)
	assert.NotEqual(t, "rahasia123", storedHash, "password must not be stored in plain text")
	assert.True(t, auth.CheckPassword(storedHash, "rahasia123"), "stored hash must match the original password")
}

func TestAuthUC_Register_Validation(t *testing.T) {
	tests := map[string]struct {
		input     dto.RegisterRequest
		wantField string
	}{
		"empty name":                   {input: dto.RegisterRequest{Name: "", Email: "budi@example.com", Password: "rahasia123"}, wantField: "name"},
		"invalid email":                {input: dto.RegisterRequest{Name: "Budi", Email: "bukan-email", Password: "rahasia123"}, wantField: "email"},
		"email with display name":      {input: dto.RegisterRequest{Name: "Budi", Email: "Budi <budi@example.com>", Password: "rahasia123"}, wantField: "email"},
		"password too short":           {input: dto.RegisterRequest{Name: "Budi", Email: "budi@example.com", Password: "pendek"}, wantField: "password"},
		"password too long for bcrypt": {input: dto.RegisterRequest{Name: "Budi", Email: "budi@example.com", Password: strings.Repeat("a", 73)}, wantField: "password"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sut, _ := newAuthSUT(t)

			_, err := sut.Register(context.Background(), tt.input)

			var vErr *entities.ValidationError
			require.ErrorAs(t, err, &vErr)
			assert.Contains(t, vErr.Fields, tt.wantField)
		})
	}
}

func TestAuthUC_Register_DuplicateEmail(t *testing.T) {
	sut, f := newAuthSUT(t)
	ctx := context.Background()

	// The same email in different case must reach the repository lowercased, so the UNIQUE constraint catches it.
	f.userRepo.EXPECT().
		Create(ctx, "Budi", "budi@example.com", gomock.Any()).
		Return(nil, entities.ErrEmailAlreadyExists)

	_, err := sut.Register(ctx, dto.RegisterRequest{Name: "Budi", Email: "BUDI@example.com", Password: "rahasia123"})

	assert.ErrorIs(t, err, entities.ErrEmailAlreadyExists)
}

func TestAuthUC_Login(t *testing.T) {
	hash, err := auth.HashPassword("rahasia123")
	require.NoError(t, err)

	user := &entities.User{ID: 1, Email: "budi@example.com", PasswordHash: hash}
	expiresAt := time.Now().Add(time.Hour)

	tests := map[string]struct {
		input     dto.LoginRequest
		setup     func(ctx context.Context, f authFields)
		wantErr   error
		wantToken string
	}{
		"correct credentials": {
			input: dto.LoginRequest{Email: "budi@example.com", Password: "rahasia123"},
			setup: func(ctx context.Context, f authFields) {
				f.userRepo.EXPECT().GetByEmail(ctx, "budi@example.com").Return(user, nil)
				f.tokens.EXPECT().Generate(int64(1)).Return("token-for-user-1", expiresAt, nil)
			},
			wantToken: "token-for-user-1",
		},
		"email in different case": {
			input: dto.LoginRequest{Email: "BUDI@Example.com", Password: "rahasia123"},
			setup: func(ctx context.Context, f authFields) {
				f.userRepo.EXPECT().GetByEmail(ctx, "budi@example.com").Return(user, nil)
				f.tokens.EXPECT().Generate(int64(1)).Return("token-for-user-1", expiresAt, nil)
			},
			wantToken: "token-for-user-1",
		},
		"wrong password": {
			input: dto.LoginRequest{Email: "budi@example.com", Password: "salah12345"},
			setup: func(ctx context.Context, f authFields) {
				f.userRepo.EXPECT().GetByEmail(ctx, "budi@example.com").Return(user, nil)
			},
			wantErr: entities.ErrInvalidCredentials,
		},
		"unknown email": {
			input: dto.LoginRequest{Email: "tidakada@example.com", Password: "rahasia123"},
			setup: func(ctx context.Context, f authFields) {
				f.userRepo.EXPECT().GetByEmail(ctx, "tidakada@example.com").Return(nil, entities.ErrNotFound)
			},
			wantErr: entities.ErrInvalidCredentials,
		},
		"empty password": {
			input:   dto.LoginRequest{Email: "budi@example.com", Password: ""},
			setup:   func(ctx context.Context, f authFields) {},
			wantErr: entities.ErrInvalidCredentials,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sut, f := newAuthSUT(t)
			ctx := context.Background()
			tt.setup(ctx, f)

			got, err := sut.Login(ctx, tt.input)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantToken, got.AccessToken)
			assert.Equal(t, "Bearer", got.TokenType)
		})
	}
}
