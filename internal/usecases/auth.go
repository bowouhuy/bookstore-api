package usecases

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/entities"
	"github.com/bowouhuy/bookstore-api/pkg/auth"
)

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

func (u *authUsecase) Register(ctx context.Context, input dto.RegisterRequest) (*entities.User, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = normalizeEmail(input.Email)

	if err := validateRegisterRequest(input); err != nil {
		return nil, err
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	return u.userRepo.Create(ctx, input.Name, input.Email, hash)
}

func (u *authUsecase) Login(ctx context.Context, input dto.LoginRequest) (*dto.AuthTokenResponse, error) {
	email := normalizeEmail(input.Email)
	if email == "" || input.Password == "" {
		return nil, entities.ErrInvalidCredentials
	}

	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			// Run bcrypt anyway to prevent timing-based email enumeration.
			auth.CheckPassword(dummyPasswordHash(), input.Password)
			return nil, entities.ErrInvalidCredentials
		}
		return nil, err
	}

	if !auth.CheckPassword(user.PasswordHash, input.Password) {
		return nil, entities.ErrInvalidCredentials
	}

	token, expiresAt, err := u.tokens.Generate(user.ID)
	if err != nil {
		return nil, err
	}

	return &dto.AuthTokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
	}, nil
}

func (u *authUsecase) GetUser(ctx context.Context, id int64) (*entities.User, error) {
	if id <= 0 {
		return nil, entities.ErrNotFound
	}

	return u.userRepo.GetByID(ctx, id)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateRegisterRequest(input dto.RegisterRequest) error {
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
		return &entities.ValidationError{Fields: fields}
	}

	return nil
}

// isValidEmail also rejects display-name forms such as "Bob <bob@example.com>".
func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}
