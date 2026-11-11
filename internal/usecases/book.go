package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/entities"
)

const (
	defaultPageSize = 10
	maxPageSize     = 100
	maxTitleLength  = 255
	maxAuthorLength = 255
)

func (u *bookUsecase) Create(ctx context.Context, input dto.BookRequest) (*entities.Book, error) {
	input = normalizeBookRequest(input)

	if err := validateBookRequest(input); err != nil {
		return nil, err
	}

	return u.bookRepo.Create(ctx, input)
}

func (u *bookUsecase) GetByID(ctx context.Context, id int64) (*entities.Book, error) {
	if id <= 0 {
		return nil, entities.ErrNotFound
	}

	return u.bookRepo.GetByID(ctx, id)
}

func (u *bookUsecase) List(ctx context.Context, page, pageSize int) (*dto.BookList, error) {
	page, pageSize = normalizePagination(page, pageSize)
	offset := (page - 1) * pageSize

	books, err := u.bookRepo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &dto.BookList{
		Books:    books,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (u *bookUsecase) Update(ctx context.Context, id int64, input dto.BookRequest) (*entities.Book, error) {
	if id <= 0 {
		return nil, entities.ErrNotFound
	}

	input = normalizeBookRequest(input)

	if err := validateBookRequest(input); err != nil {
		return nil, err
	}

	return u.bookRepo.Update(ctx, id, input)
}

func (u *bookUsecase) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return entities.ErrNotFound
	}

	return u.bookRepo.Delete(ctx, id)
}

func normalizeBookRequest(input dto.BookRequest) dto.BookRequest {
	input.Title = strings.TrimSpace(input.Title)
	input.Author = strings.TrimSpace(input.Author)
	return input
}

func validateBookRequest(input dto.BookRequest) error {
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
		return &entities.ValidationError{Fields: fields}
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
