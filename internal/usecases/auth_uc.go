package usecases

//go:generate rm -f ./auth_uc_mock.go
//go:generate mockgen -destination auth_uc_mock.go -package usecases -mock_names AuthUsecase=GoMockAuthUsecase,TokenGenerator=GoMockTokenGenerator -source auth_uc.go

import (
	"context"
	"time"

	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/entities"
	"github.com/bowouhuy/bookstore-api/internal/entities/repository"
)

// AuthUsecase is an interface for auth usecase
type AuthUsecase interface {
	Register(ctx context.Context, input dto.RegisterRequest) (*entities.User, error)
	Login(ctx context.Context, input dto.LoginRequest) (*dto.AuthTokenResponse, error)
	GetUser(ctx context.Context, id int64) (*entities.User, error)
}

// TokenGenerator issues access tokens for authenticated users.
type TokenGenerator interface {
	Generate(userID int64) (token string, expiresAt time.Time, err error)
}

// Compile time implementation check
var _ AuthUsecase = (*authUsecase)(nil)

// NewAuthUsecase is a constructor function for auth usecase
func NewAuthUsecase(
	userRepo repository.UserRepo,
	tokens TokenGenerator,
) AuthUsecase {
	return &authUsecase{
		userRepo: userRepo,
		tokens:   tokens,
	}
}

// authUsecase is a struct for auth usecase
type authUsecase struct {
	userRepo repository.UserRepo
	tokens   TokenGenerator
}
