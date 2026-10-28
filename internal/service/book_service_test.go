package service

import (
	"context"
	"errors"
	"testing"

	"github.com/bowouhuy/bookstore-api/internal/model"
)

type fakeBookRepo struct {
	createCalled bool
	lastLimit    int
	lastOffset   int
}

func (f *fakeBookRepo) Create(ctx context.Context, input model.BookInput) (*model.Book, error) {
	f.createCalled = true
	return &model.Book{
		ID:            1,
		Title:         input.Title,
		Author:        input.Author,
		PublishedYear: input.PublishedYear,
	}, nil
}

func (f *fakeBookRepo) GetByID(ctx context.Context, id int64) (*model.Book, error) {
	return nil, model.ErrNotFound
}

func (f *fakeBookRepo) List(ctx context.Context, limit, offset int) ([]model.Book, error) {
	f.lastLimit = limit
	f.lastOffset = offset
	return []model.Book{}, nil
}

func (f *fakeBookRepo) Update(ctx context.Context, id int64, input model.BookInput) (*model.Book, error) {
	return nil, model.ErrNotFound
}

func (f *fakeBookRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

func intPtr(v int) *int {
	return &v
}

func TestBookService_Create(t *testing.T) {
	tests := []struct {
		name      string
		input     model.BookInput
		wantField string // field expected to fail; empty means valid input
	}{
		{
			name:      "valid input",
			input:     model.BookInput{Title: "Laskar Pelangi", Author: "Andrea Hirata", PublishedYear: intPtr(2005)},
			wantField: "",
		},
		{
			name:      "valid input without year",
			input:     model.BookInput{Title: "Bumi Manusia", Author: "Pramoedya Ananta Toer"},
			wantField: "",
		},
		{
			name:      "empty title",
			input:     model.BookInput{Title: "", Author: "Andrea Hirata"},
			wantField: "title",
		},
		{
			name:      "whitespace-only title",
			input:     model.BookInput{Title: "   ", Author: "Andrea Hirata"},
			wantField: "title",
		},
		{
			name:      "empty author",
			input:     model.BookInput{Title: "Laskar Pelangi", Author: ""},
			wantField: "author",
		},
		{
			name:      "year in the future",
			input:     model.BookInput{Title: "Laskar Pelangi", Author: "Andrea Hirata", PublishedYear: intPtr(3000)},
			wantField: "published_year",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeBookRepo{}
			svc := NewBookService(repo)

			_, err := svc.Create(context.Background(), tt.input)

			if tt.wantField == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if !repo.createCalled {
					t.Fatal("expected repository Create to be called")
				}
				return
			}

			var vErr *model.ValidationError
			if !errors.As(err, &vErr) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
			if _, ok := vErr.Fields[tt.wantField]; !ok {
				t.Errorf("expected error on field %q, got %v", tt.wantField, vErr.Fields)
			}
			if repo.createCalled {
				t.Error("repository must not be called when input is invalid")
			}
		})
	}
}

func TestBookService_List(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		pageSize   int
		wantLimit  int
		wantOffset int
	}{
		{name: "default values", page: 0, pageSize: 0, wantLimit: 10, wantOffset: 0},
		{name: "third page", page: 3, pageSize: 20, wantLimit: 20, wantOffset: 40},
		{name: "page size capped", page: 1, pageSize: 500, wantLimit: 100, wantOffset: 0},
		{name: "negative page", page: -5, pageSize: 10, wantLimit: 10, wantOffset: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeBookRepo{}
			svc := NewBookService(repo)

			result, err := svc.List(context.Background(), tt.page, tt.pageSize)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if repo.lastLimit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", repo.lastLimit, tt.wantLimit)
			}
			if repo.lastOffset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", repo.lastOffset, tt.wantOffset)
			}
			if result.PageSize != tt.wantLimit {
				t.Errorf("result.PageSize = %d, want %d", result.PageSize, tt.wantLimit)
			}
		})
	}
}
