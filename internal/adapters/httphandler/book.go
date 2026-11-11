package httphandler

import (
	"fmt"
	"net/http"

	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/usecases"
)

type createBookHandler struct {
	bookUsecase usecases.BookUsecase
}

func NewCreateBookHandler(bookUsecase usecases.BookUsecase) Handler {
	return &createBookHandler{bookUsecase: bookUsecase}
}

// Handle handles POST /books.
func (h *createBookHandler) Handle(req *http.Request) dto.HTTPResponse {
	var input dto.BookRequest
	if err := decodeJSON(req, &input); err != nil {
		return decodeErrorResponse(err)
	}

	book, err := h.bookUsecase.Create(req.Context(), input)
	if err != nil {
		return errorResponse(req, err)
	}

	return *dto.NewResponse().
		WithCode(http.StatusCreated).
		WithHeader("Location", fmt.Sprintf("/books/%d", book.ID)).
		WithData(book)
}

type listBooksHandler struct {
	bookUsecase usecases.BookUsecase
}

func NewListBooksHandler(bookUsecase usecases.BookUsecase) Handler {
	return &listBooksHandler{bookUsecase: bookUsecase}
}

// Handle handles GET /books.
func (h *listBooksHandler) Handle(req *http.Request) dto.HTTPResponse {
	page, err := queryInt(req, "page", 1)
	if err != nil {
		return badRequest(err.Error())
	}

	pageSize, err := queryInt(req, "page_size", 0)
	if err != nil {
		return badRequest(err.Error())
	}

	list, err := h.bookUsecase.List(req.Context(), page, pageSize)
	if err != nil {
		return errorResponse(req, err)
	}

	return *dto.NewResponse().
		WithCode(http.StatusOK).
		WithData(list.Books).
		WithMeta(dto.MetaData{Page: uint64(list.Page), Limit: uint64(list.PageSize)})
}

type getBookHandler struct {
	bookUsecase usecases.BookUsecase
}

func NewGetBookHandler(bookUsecase usecases.BookUsecase) Handler {
	return &getBookHandler{bookUsecase: bookUsecase}
}

// Handle handles GET /books/{id}.
func (h *getBookHandler) Handle(req *http.Request) dto.HTTPResponse {
	id, err := pathID(req)
	if err != nil {
		return badRequest(err.Error())
	}

	book, err := h.bookUsecase.GetByID(req.Context(), id)
	if err != nil {
		return errorResponse(req, err)
	}

	return *dto.NewResponse().WithCode(http.StatusOK).WithData(book)
}

type updateBookHandler struct {
	bookUsecase usecases.BookUsecase
}

func NewUpdateBookHandler(bookUsecase usecases.BookUsecase) Handler {
	return &updateBookHandler{bookUsecase: bookUsecase}
}

// Handle handles PUT /books/{id}.
func (h *updateBookHandler) Handle(req *http.Request) dto.HTTPResponse {
	id, err := pathID(req)
	if err != nil {
		return badRequest(err.Error())
	}

	var input dto.BookRequest
	if err := decodeJSON(req, &input); err != nil {
		return decodeErrorResponse(err)
	}

	book, err := h.bookUsecase.Update(req.Context(), id, input)
	if err != nil {
		return errorResponse(req, err)
	}

	return *dto.NewResponse().WithCode(http.StatusOK).WithData(book)
}

type deleteBookHandler struct {
	bookUsecase usecases.BookUsecase
}

func NewDeleteBookHandler(bookUsecase usecases.BookUsecase) Handler {
	return &deleteBookHandler{bookUsecase: bookUsecase}
}

// Handle handles DELETE /books/{id}.
func (h *deleteBookHandler) Handle(req *http.Request) dto.HTTPResponse {
	id, err := pathID(req)
	if err != nil {
		return badRequest(err.Error())
	}

	if err := h.bookUsecase.Delete(req.Context(), id); err != nil {
		return errorResponse(req, err)
	}

	return *dto.NewResponse().WithCode(http.StatusNoContent)
}
