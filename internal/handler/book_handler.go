package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bowouhuy/bookstore-api/internal/model"
)

type BookService interface {
	Create(ctx context.Context, input model.BookInput) (*model.Book, error)
	GetByID(ctx context.Context, id int64) (*model.Book, error)
	List(ctx context.Context, page, pageSize int) (*model.BookList, error)
	Update(ctx context.Context, id int64, input model.BookInput) (*model.Book, error)
	Delete(ctx context.Context, id int64) error
}

type BookHandler struct {
	service BookService
}

func NewBookHandler(service BookService) *BookHandler {
	return &BookHandler{service: service}
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input model.BookInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeDecodeError(w, err)
		return
	}

	book, err := h.service.Create(r.Context(), input)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/books/%d", book.ID))
	writeJSON(w, http.StatusCreated, book)
}

func (h *BookHandler) List(w http.ResponseWriter, r *http.Request) {
	page, err := queryInt(r, "page", 1)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	pageSize, err := queryInt(r, "page_size", 0)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.List(r.Context(), page, pageSize)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *BookHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	book, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, book)
}

func (h *BookHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input model.BookInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeDecodeError(w, err)
		return
	}

	book, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, book)
}

func (h *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		handleServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
