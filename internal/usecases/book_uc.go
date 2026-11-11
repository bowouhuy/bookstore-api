package usecases

//go:generate rm -f ./book_uc_mock.go
//go:generate mockgen -destination book_uc_mock.go -package usecases -mock_names BookUsecase=GoMockBookUsecase -source book_uc.go

import (
	"context"

	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/entities"
	"github.com/bowouhuy/bookstore-api/internal/entities/repository"
)

// BookUsecase is an interface for book usecase
type BookUsecase interface {
	Create(ctx context.Context, input dto.BookRequest) (*entities.Book, error)
	GetByID(ctx context.Context, id int64) (*entities.Book, error)
	List(ctx context.Context, page, pageSize int) (*dto.BookList, error)
	Update(ctx context.Context, id int64, input dto.BookRequest) (*entities.Book, error)
	Delete(ctx context.Context, id int64) error
}

// Compile time implementation check
var _ BookUsecase = (*bookUsecase)(nil)

// NewBookUsecase is a constructor function for book usecase
func NewBookUsecase(
	bookRepo repository.BookRepo,
) BookUsecase {
	return &bookUsecase{
		bookRepo: bookRepo,
	}
}

// bookUsecase is a struct for book usecase
type bookUsecase struct {
	bookRepo repository.BookRepo
}
