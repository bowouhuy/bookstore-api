package repository

//go:generate rm -f ./user_mock.go
//go:generate mockgen -destination user_mock.go -package repository -mock_names UserRepo=GoMockUserRepo -source user.go

import (
	"context"

	"github.com/bowouhuy/bookstore-api/internal/entities"
)

// UserRepo is a user repository interface.
type UserRepo interface {
	// Create expects an already hashed password, never a plain one.
	Create(ctx context.Context, name, email, passwordHash string) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	GetByID(ctx context.Context, id int64) (*entities.User, error)
}
