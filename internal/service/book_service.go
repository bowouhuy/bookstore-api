package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bowouhuy/bookstore-api/internal/model"
)

type BookRepository interface {
	Create(ctx context.Context, input model.BookInput) (*model.Book, error)
	GetByID(ctx context.Context, id int64) (*model.Book, error)
	List(ctx context.Context, limit, offset int) ([]model.Book, error)
	Update(ctx context.Context, id int64, input model.BookInput) (*model.Book, error)
	Delete(ctx context.Context, id int64) error
}

const (
	defaultPageSize = 10
	maxPageSize     = 100
	maxTitleLength  = 255
	maxAuthorLength = 255
)

type BookService struct {
	repo BookRepository
}

func NewBookService(repo BookRepository) *BookService {
	return &BookService{repo: repo}
}

func (s *BookService) Create(ctx context.Context, input model.BookInput) (*model.Book, error) {
	input = normalizeBookInput(input)

	if err := validateBookInput(input); err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, input)
}

func (s *BookService) GetByID(ctx context.Context, id int64) (*model.Book, error) {
	if id <= 0 {
		return nil, model.ErrNotFound
	}

	return s.repo.GetByID(ctx, id)
}

func (s *BookService) List(ctx context.Context, page, pageSize int) (*model.BookList, error) {
	page, pageSize = normalizePagination(page, pageSize)
	offset := (page - 1) * pageSize

	books, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &model.BookList{
		Books:    books,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *BookService) Update(ctx context.Context, id int64, input model.BookInput) (*model.Book, error) {
	if id <= 0 {
		return nil, model.ErrNotFound
	}

	input = normalizeBookInput(input)

	if err := validateBookInput(input); err != nil {
		return nil, err
	}

	return s.repo.Update(ctx, id, input)
}

func (s *BookService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return model.ErrNotFound
	}

	return s.repo.Delete(ctx, id)
}

func normalizeBookInput(input model.BookInput) model.BookInput {
	input.Title = strings.TrimSpace(input.Title)
	input.Author = strings.TrimSpace(input.Author)
	return input
}

func validateBookInput(input model.BookInput) error {
	fields := make(map[string]string)

	switch {
	case input.Title == "":
		fields["title"] = "is required"
	case utf8.RuneCountInString(input.Title) > maxTitleLength:
		fields["title"] = fmt.Sprintf("must be at most %d characters", maxTitleLength)
	}

	switch {
	case input.Author == "":
		fields["author"] = "is required"
	case utf8.RuneCountInString(input.Author) > maxAuthorLength:
		fields["author"] = fmt.Sprintf("must be at most %d characters", maxAuthorLength)
	}

	if input.PublishedYear != nil {
		year := *input.PublishedYear
		currentYear := time.Now().Year()
		if year < 1 || year > currentYear {
			fields["published_year"] = fmt.Sprintf("must be between 1 and %d", currentYear)
		}
	}

	if len(fields) > 0 {
		return &model.ValidationError{Fields: fields}
	}
	return nil
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}
