package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bowouhuy/bookstore-api/internal/model"
)

type fakeBookService struct {
	book *model.Book
	list *model.BookList
	err  error
}

func (f *fakeBookService) Create(ctx context.Context, input model.BookInput) (*model.Book, error) {
	return f.book, f.err
}

func (f *fakeBookService) GetByID(ctx context.Context, id int64) (*model.Book, error) {
	return f.book, f.err
}

func (f *fakeBookService) List(ctx context.Context, page, pageSize int) (*model.BookList, error) {
	return f.list, f.err
}

func (f *fakeBookService) Update(ctx context.Context, id int64, input model.BookInput) (*model.Book, error) {
	return f.book, f.err
}

func (f *fakeBookService) Delete(ctx context.Context, id int64) error {
	return f.err
}

func TestBookHandler_Create(t *testing.T) {
	validBody := `{"title": "Laskar Pelangi", "author": "Andrea Hirata"}`

	tests := []struct {
		name       string
		body       string
		service    *fakeBookService
		wantStatus int
	}{
		{
			name:       "success",
			body:       validBody,
			service:    &fakeBookService{book: &model.Book{ID: 7, Title: "Laskar Pelangi"}},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "malformed JSON",
			body:       `{"title": "Laskar Pelangi",`,
			service:    &fakeBookService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body",
			body:       ``,
			service:    &fakeBookService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown field",
			body:       `{"title": "A", "author": "B", "id": 99}`,
			service:    &fakeBookService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong field type",
			body:       `{"title": "A", "author": "B", "published_year": "2005"}`,
			service:    &fakeBookService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "validation error from service",
			body: validBody,
			service: &fakeBookService{
				err: &model.ValidationError{Fields: map[string]string{"title": "is required"}},
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "unexpected error from service",
			body:       validBody,
			service:    &fakeBookService{err: errors.New("database is down")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewBookHandler(tt.service)

			req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			h.Create(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
		})
	}
}

func TestBookHandler_Create_SuccessResponse(t *testing.T) {
	svc := &fakeBookService{book: &model.Book{ID: 7, Title: "Laskar Pelangi", Author: "Andrea Hirata"}}
	h := NewBookHandler(svc)

	body := `{"title": "Laskar Pelangi", "author": "Andrea Hirata"}`
	req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if got := rec.Header().Get("Location"); got != "/books/7" {
		t.Errorf("Location = %q, want /books/7", got)
	}

	var got model.Book
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.ID != 7 || got.Title != "Laskar Pelangi" {
		t.Errorf("unexpected response body: %+v", got)
	}
}

func TestBookHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		service    *fakeBookService
		wantStatus int
	}{
		{
			name:       "found",
			id:         "1",
			service:    &fakeBookService{book: &model.Book{ID: 1}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "not found",
			id:         "99",
			service:    &fakeBookService{err: model.ErrNotFound},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "id is not a number",
			id:         "abc",
			service:    &fakeBookService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "id is zero",
			id:         "0",
			service:    &fakeBookService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewBookHandler(tt.service)

			req := httptest.NewRequest(http.MethodGet, "/books/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			rec := httptest.NewRecorder()

			h.GetByID(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestBookHandler_Delete(t *testing.T) {
	h := NewBookHandler(&fakeBookService{})

	req := httptest.NewRequest(http.MethodDelete, "/books/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body, got %q", rec.Body.String())
	}
}
