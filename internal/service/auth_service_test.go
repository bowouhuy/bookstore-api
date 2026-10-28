package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bowouhuy/bookstore-api/internal/auth"
	"github.com/bowouhuy/bookstore-api/internal/model"
)

// fakeUserRepo is an in-memory store that mimics the real repository.
type fakeUserRepo struct {
	usersByEmail map[string]*model.User
	nextID       int64
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		usersByEmail: make(map[string]*model.User),
		nextID:       1,
	}
}

func (f *fakeUserRepo) Create(ctx context.Context, name, email, passwordHash string) (*model.User, error) {
	if _, exists := f.usersByEmail[email]; exists {
		return nil, model.ErrEmailAlreadyExists
	}

	user := &model.User{
		ID:           f.nextID,
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
	}
	f.nextID++
	f.usersByEmail[email] = user
	return user, nil
}

func (f *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, ok := f.usersByEmail[email]
	if !ok {
		return nil, model.ErrNotFound
	}
	return user, nil
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	for _, user := range f.usersByEmail {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, model.ErrNotFound
}

// fakeTokenGenerator returns predictable tokens.
type fakeTokenGenerator struct{}

func (fakeTokenGenerator) Generate(userID int64) (string, time.Time, error) {
	return fmt.Sprintf("token-for-user-%d", userID), time.Now().Add(time.Hour), nil
}

func TestAuthService_Register_Success(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewAuthService(repo, fakeTokenGenerator{})

	user, err := svc.Register(context.Background(), model.RegisterInput{
		Name:     "  Budi Santoso  ",
		Email:    "  Budi@Example.COM ",
		Password: "rahasia123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Name != "Budi Santoso" {
		t.Errorf("name = %q, want trimmed %q", user.Name, "Budi Santoso")
	}
	if user.Email != "budi@example.com" {
		t.Errorf("email = %q, want normalized %q", user.Email, "budi@example.com")
	}
	if user.PasswordHash == "rahasia123" {
		t.Fatal("password must not be stored in plain text")
	}
	if !auth.CheckPassword(user.PasswordHash, "rahasia123") {
		t.Error("stored hash does not match the original password")
	}
}

func TestAuthService_Register_Validation(t *testing.T) {
	tests := []struct {
		name      string
		input     model.RegisterInput
		wantField string
	}{
		{
			name:      "empty name",
			input:     model.RegisterInput{Name: "", Email: "budi@example.com", Password: "rahasia123"},
			wantField: "name",
		},
		{
			name:      "invalid email",
			input:     model.RegisterInput{Name: "Budi", Email: "bukan-email", Password: "rahasia123"},
			wantField: "email",
		},
		{
			name:      "email with display name",
			input:     model.RegisterInput{Name: "Budi", Email: "Budi <budi@example.com>", Password: "rahasia123"},
			wantField: "email",
		},
		{
			name:      "password too short",
			input:     model.RegisterInput{Name: "Budi", Email: "budi@example.com", Password: "pendek"},
			wantField: "password",
		},
		{
			name:      "password too long for bcrypt",
			input:     model.RegisterInput{Name: "Budi", Email: "budi@example.com", Password: strings.Repeat("a", 73)},
			wantField: "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewAuthService(newFakeUserRepo(), fakeTokenGenerator{})

			_, err := svc.Register(context.Background(), tt.input)

			var vErr *model.ValidationError
			if !errors.As(err, &vErr) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
			if _, ok := vErr.Fields[tt.wantField]; !ok {
				t.Errorf("expected error on field %q, got %v", tt.wantField, vErr.Fields)
			}
		})
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	svc := NewAuthService(newFakeUserRepo(), fakeTokenGenerator{})
	ctx := context.Background()

	first := model.RegisterInput{Name: "Budi", Email: "budi@example.com", Password: "rahasia123"}
	if _, err := svc.Register(ctx, first); err != nil {
		t.Fatalf("first register: unexpected error: %v", err)
	}

	// The same email in different case must still be a duplicate.
	second := model.RegisterInput{Name: "Budi Lain", Email: "BUDI@example.com", Password: "rahasia456"}
	_, err := svc.Register(ctx, second)
	if !errors.Is(err, model.ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestAuthService_Login(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewAuthService(repo, fakeTokenGenerator{})
	ctx := context.Background()

	_, err := svc.Register(ctx, model.RegisterInput{
		Name:     "Budi",
		Email:    "budi@example.com",
		Password: "rahasia123",
	})
	if err != nil {
		t.Fatalf("setup register: unexpected error: %v", err)
	}

	tests := []struct {
		name      string
		input     model.LoginInput
		wantErr   error
		wantToken string
	}{
		{
			name:      "correct credentials",
			input:     model.LoginInput{Email: "budi@example.com", Password: "rahasia123"},
			wantToken: "token-for-user-1",
		},
		{
			name:      "email in different case",
			input:     model.LoginInput{Email: "BUDI@Example.com", Password: "rahasia123"},
			wantToken: "token-for-user-1",
		},
		{
			name:    "wrong password",
			input:   model.LoginInput{Email: "budi@example.com", Password: "salah12345"},
			wantErr: model.ErrInvalidCredentials,
		},
		{
			name:    "unknown email",
			input:   model.LoginInput{Email: "tidakada@example.com", Password: "rahasia123"},
			wantErr: model.ErrInvalidCredentials,
		},
		{
			name:    "empty password",
			input:   model.LoginInput{Email: "budi@example.com", Password: ""},
			wantErr: model.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.Login(ctx, tt.input)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.AccessToken != tt.wantToken {
				t.Errorf("token = %q, want %q", result.AccessToken, tt.wantToken)
			}
			if result.TokenType != "Bearer" {
				t.Errorf("token type = %q, want Bearer", result.TokenType)
			}
		})
	}
}
