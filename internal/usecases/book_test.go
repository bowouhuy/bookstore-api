package usecases_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/entities"
	"github.com/bowouhuy/bookstore-api/internal/entities/repository"
	"github.com/bowouhuy/bookstore-api/internal/usecases"
)

func intPtr(v int) *int {
	return &v
}

func TestBookUC_Create(t *testing.T) {
	tests := map[string]struct {
		input     dto.BookRequest
		wantField string // field expected to fail; empty means valid input
	}{
		"valid input":           {input: dto.BookRequest{Title: "Laskar Pelangi", Author: "Andrea Hirata", PublishedYear: intPtr(2005)}},
		"valid input w/o year":  {input: dto.BookRequest{Title: "Bumi Manusia", Author: "Pramoedya Ananta Toer"}},
		"empty title":           {input: dto.BookRequest{Title: "", Author: "Andrea Hirata"}, wantField: "title"},
		"whitespace-only title": {input: dto.BookRequest{Title: "   ", Author: "Andrea Hirata"}, wantField: "title"},
		"empty author":          {input: dto.BookRequest{Title: "Laskar Pelangi", Author: ""}, wantField: "author"},
		"year in the future":    {input: dto.BookRequest{Title: "Laskar Pelangi", Author: "Andrea Hirata", PublishedYear: intPtr(3000)}, wantField: "published_year"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()
			repo := repository.NewGoMockBookRepo(ctrl)

			if tt.wantField == "" {
				repo.EXPECT().Create(ctx, tt.input).Return(&entities.Book{ID: 1}, nil)
			}
			// With no EXPECT, gomock fails the test if the repository is called for invalid input.

			_, err := usecases.NewBookUsecase(repo).Create(ctx, tt.input)

			if tt.wantField == "" {
				require.NoError(t, err)
				return
			}

			var vErr *entities.ValidationError
			require.ErrorAs(t, err, &vErr)
			assert.Contains(t, vErr.Fields, tt.wantField)
		})
	}
}

func TestBookUC_Create_TrimsInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	repo := repository.NewGoMockBookRepo(ctrl)

	repo.EXPECT().
		Create(ctx, dto.BookRequest{Title: "Laskar Pelangi", Author: "Andrea Hirata"}).
		Return(&entities.Book{ID: 1}, nil)

	_, err := usecases.NewBookUsecase(repo).Create(ctx, dto.BookRequest{Title: "  Laskar Pelangi ", Author: " Andrea Hirata  "})
	require.NoError(t, err)
}

func TestBookUC_List(t *testing.T) {
	tests := map[string]struct {
		page, pageSize        int
		wantLimit, wantOffset int
		wantPage              int
	}{
		"default values":   {page: 0, pageSize: 0, wantLimit: 10, wantOffset: 0, wantPage: 1},
		"third page":       {page: 3, pageSize: 20, wantLimit: 20, wantOffset: 40, wantPage: 3},
		"page size capped": {page: 1, pageSize: 500, wantLimit: 100, wantOffset: 0, wantPage: 1},
		"negative page":    {page: -5, pageSize: 10, wantLimit: 10, wantOffset: 0, wantPage: 1},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()
			repo := repository.NewGoMockBookRepo(ctrl)

			repo.EXPECT().List(ctx, tt.wantLimit, tt.wantOffset).Return([]entities.Book{}, nil)

			got, err := usecases.NewBookUsecase(repo).List(ctx, tt.page, tt.pageSize)

			require.NoError(t, err)
			assert.Equal(t, tt.wantPage, got.Page)
			assert.Equal(t, tt.wantLimit, got.PageSize)
		})
	}
}

func TestBookUC_InvalidIDSkipsRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	uc := usecases.NewBookUsecase(repository.NewGoMockBookRepo(ctrl))

	_, err := uc.GetByID(ctx, 0)
	assert.ErrorIs(t, err, entities.ErrNotFound)

	_, err = uc.Update(ctx, -1, dto.BookRequest{Title: "A", Author: "B"})
	assert.ErrorIs(t, err, entities.ErrNotFound)

	assert.ErrorIs(t, uc.Delete(ctx, 0), entities.ErrNotFound)
}
