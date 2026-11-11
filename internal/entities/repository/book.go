package repository

//go:generate rm -f ./book_mock.go
//go:generate mockgen -destination book_mock.go -package repository -mock_names BookRepo=GoMockBookRepo -source book.go

import (
	"context"

	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/entities"
)

// BookRepo is a book repository interface.
type BookRepo interface {
	Create(ctx context.Context, input dto.BookRequest) (*entities.Book, error)
	GetByID(ctx context.Context, id int64) (*entities.Book, error)
	List(ctx context.Context, limit, offset int) ([]entities.Book, error)
	Update(ctx context.Context, id int64, input dto.BookRequest) (*entities.Book, error)
	Delete(ctx context.Context, id int64) error
}
