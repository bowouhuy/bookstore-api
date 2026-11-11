package httphandler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bowouhuy/bookstore-api/internal/adapters/httphandler"
	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/entities"
	"github.com/bowouhuy/bookstore-api/internal/usecases"
)

const validBookBody = `{"title": "Laskar Pelangi", "author": "Andrea Hirata"}`

func TestCreateBookHandler(t *testing.T) {
	tests := map[string]struct {
		body     string
		setup    func(uc *usecases.GoMockBookUsecase)
		wantCode int
	}{
		"success": {
			body: validBookBody,
			setup: func(uc *usecases.GoMockBookUsecase) {
				uc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&entities.Book{ID: 7, Title: "Laskar Pelangi"}, nil)
			},
			wantCode: http.StatusCreated,
		},
		"malformed JSON":   {body: `{"title": "Laskar Pelangi",`, wantCode: http.StatusBadRequest},
		"empty body":       {body: ``, wantCode: http.StatusBadRequest},
		"unknown field":    {body: `{"title": "A", "author": "B", "id": 99}`, wantCode: http.StatusBadRequest},
		"wrong field type": {body: `{"title": "A", "author": "B", "published_year": "2005"}`, wantCode: http.StatusBadRequest},
		"two JSON objects": {body: validBookBody + validBookBody, wantCode: http.StatusBadRequest},
		"validation error from usecase": {
			body: validBookBody,
			setup: func(uc *usecases.GoMockBookUsecase) {
				uc.EXPECT().Create(gomock.Any(), gomock.Any()).
					Return(nil, &entities.ValidationError{Fields: map[string]string{"title": "is required"}})
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		"unexpected error from usecase": {
			body: validBookBody,
			setup: func(uc *usecases.GoMockBookUsecase) {
				uc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("database is down"))
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			uc := usecases.NewGoMockBookUsecase(gomock.NewController(t))
			if tt.setup != nil {
				tt.setup(uc)
			}

			req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(tt.body))
			resp := httphandler.NewCreateBookHandler(uc).Handle(req)

			assert.Equal(t, tt.wantCode, resp.Code)
		})
	}
}

func TestCreateBookHandler_SuccessResponse(t *testing.T) {
	book := &entities.Book{ID: 7, Title: "Laskar Pelangi", Author: "Andrea Hirata"}

	uc := usecases.NewGoMockBookUsecase(gomock.NewController(t))
	uc.EXPECT().Create(gomock.Any(), dto.BookRequest{Title: "Laskar Pelangi", Author: "Andrea Hirata"}).Return(book, nil)

	req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(validBookBody))
	resp := httphandler.NewCreateBookHandler(uc).Handle(req)

	assert.Equal(t, "/books/7", resp.Headers()["Location"])
	assert.Equal(t, book, resp.Data)
}

func TestCreateBookHandler_BodyTooLarge(t *testing.T) {
	uc := usecases.NewGoMockBookUsecase(gomock.NewController(t))

	body := `{"title": "` + strings.Repeat("a", 100) + `", "author": "B"}`
	req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(body))
	req.Body = http.MaxBytesReader(httptest.NewRecorder(), req.Body, 16)

	resp := httphandler.NewCreateBookHandler(uc).Handle(req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.Code)
}

func TestListBooksHandler(t *testing.T) {
	t.Run("passes pagination and returns meta", func(t *testing.T) {
		uc := usecases.NewGoMockBookUsecase(gomock.NewController(t))
		uc.EXPECT().List(gomock.Any(), 2, 5).Return(&dto.BookList{Books: []entities.Book{{ID: 1}}, Page: 2, PageSize: 5}, nil)

		req := httptest.NewRequest(http.MethodGet, "/books?page=2&page_size=5", nil)
		resp := httphandler.NewListBooksHandler(uc).Handle(req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, []entities.Book{{ID: 1}}, resp.Data)
		assert.Equal(t, dto.MetaData{Page: 2, Limit: 5}, resp.Meta)
	})

	t.Run("non-numeric page", func(t *testing.T) {
		uc := usecases.NewGoMockBookUsecase(gomock.NewController(t))

		req := httptest.NewRequest(http.MethodGet, "/books?page=abc", nil)
		resp := httphandler.NewListBooksHandler(uc).Handle(req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}

func TestGetBookHandler(t *testing.T) {
	tests := map[string]struct {
		id       string
		setup    func(uc *usecases.GoMockBookUsecase)
		wantCode int
	}{
		"found": {
			id: "1",
			setup: func(uc *usecases.GoMockBookUsecase) {
				uc.EXPECT().GetByID(gomock.Any(), int64(1)).Return(&entities.Book{ID: 1}, nil)
			},
			wantCode: http.StatusOK,
		},
		"not found": {
			id: "99",
			setup: func(uc *usecases.GoMockBookUsecase) {
				uc.EXPECT().GetByID(gomock.Any(), int64(99)).Return(nil, entities.ErrNotFound)
			},
			wantCode: http.StatusNotFound,
		},
		"id is not a number": {id: "abc", wantCode: http.StatusBadRequest},
		"id is zero":         {id: "0", wantCode: http.StatusBadRequest},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			uc := usecases.NewGoMockBookUsecase(gomock.NewController(t))
			if tt.setup != nil {
				tt.setup(uc)
			}

			req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/books/"+tt.id, nil), map[string]string{"id": tt.id})
			resp := httphandler.NewGetBookHandler(uc).Handle(req)

			assert.Equal(t, tt.wantCode, resp.Code)
		})
	}
}

func TestUpdateBookHandler(t *testing.T) {
	uc := usecases.NewGoMockBookUsecase(gomock.NewController(t))
	uc.EXPECT().Update(gomock.Any(), int64(3), dto.BookRequest{Title: "Laskar Pelangi", Author: "Andrea Hirata"}).
		Return(&entities.Book{ID: 3}, nil)

	req := mux.SetURLVars(httptest.NewRequest(http.MethodPut, "/books/3", strings.NewReader(validBookBody)), map[string]string{"id": "3"})
	resp := httphandler.NewUpdateBookHandler(uc).Handle(req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestDeleteBookHandler(t *testing.T) {
	uc := usecases.NewGoMockBookUsecase(gomock.NewController(t))
	uc.EXPECT().Delete(gomock.Any(), int64(1)).Return(nil)

	req := mux.SetURLVars(httptest.NewRequest(http.MethodDelete, "/books/1", nil), map[string]string{"id": "1"})
	resp := httphandler.NewDeleteBookHandler(uc).Handle(req)

	assert.Equal(t, http.StatusNoContent, resp.Code)
}
