package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/bowouhuy/bookstore-api/internal/auth"
	"github.com/bowouhuy/bookstore-api/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, name, email, passwordHash string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
}

type TokenGenerator interface {
	Generate(userID int64) (token string, expiresAt time.Time, err error)
}

const (
	maxNameLength     = 100
	maxEmailLength    = 255
	minPasswordLength = 8
	maxPasswordBytes  = 72 // bcrypt rejects longer input
)

// dummyPasswordHash is checked when the email is unknown, so Login takes
// the same time either way and does not reveal which emails exist.
var dummyPasswordHash = sync.OnceValue(func() string {
	hash, _ := auth.HashPassword("dummy-password-for-timing-protection")
	return hash
})

type AuthService struct {
	users  UserRepository
	tokens TokenGenerator
}

func NewAuthService(users UserRepository, tokens TokenGenerator) *AuthService {
	return &AuthService{
		users:  users,
		tokens: tokens,
	}
}

func (s *AuthService) Register(ctx context.Context, input model.RegisterInput) (*model.User, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = normalizeEmail(input.Email)

	if err := validateRegisterInput(input); err != nil {
		return nil, err
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	return s.users.Create(ctx, input.Name, input.Email, hash)
}

func (s *AuthService) Login(ctx context.Context, input model.LoginInput) (*model.AuthToken, error) {
	email := normalizeEmail(input.Email)
	if email == "" || input.Password == "" {
		return nil, model.ErrInvalidCredentials
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// Run bcrypt anyway to prevent timing-based email enumeration.
			auth.CheckPassword(dummyPasswordHash(), input.Password)
			return nil, model.ErrInvalidCredentials
		}
		return nil, err
	}

	if !auth.CheckPassword(user.PasswordHash, input.Password) {
		return nil, model.ErrInvalidCredentials
	}

	token, expiresAt, err := s.tokens.Generate(user.ID)
	if err != nil {
		return nil, err
	}

	return &model.AuthToken{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *AuthService) GetUser(ctx context.Context, id int64) (*model.User, error) {
	if id <= 0 {
		return nil, model.ErrNotFound
	}
	return s.users.GetByID(ctx, id)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateRegisterInput(input model.RegisterInput) error {
	fields := make(map[string]string)

	switch {
	case input.Name == "":
		fields["name"] = "is required"
	case utf8.RuneCountInString(input.Name) > maxNameLength:
		fields["name"] = fmt.Sprintf("must be at most %d characters", maxNameLength)
	}

	switch {
	case input.Email == "":
		fields["email"] = "is required"
	case utf8.RuneCountInString(input.Email) > maxEmailLength:
		fields["email"] = fmt.Sprintf("must be at most %d characters", maxEmailLength)
	case !isValidEmail(input.Email):
		fields["email"] = "must be a valid email address"
	}

	switch {
	case utf8.RuneCountInString(input.Password) < minPasswordLength:
		fields["password"] = fmt.Sprintf("must be at least %d characters", minPasswordLength)
	case len(input.Password) > maxPasswordBytes:
		fields["password"] = fmt.Sprintf("is too long (maximum %d bytes)", maxPasswordBytes)
	}

	if len(fields) > 0 {
		return &model.ValidationError{Fields: fields}
	}
	return nil
}

// isValidEmail also rejects display-name forms such as "Bob <bob@example.com>".
func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}
