package di

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bowouhuy/bookstore-api/config"
	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/middleware"
	"github.com/bowouhuy/bookstore-api/pkg/msg"
	"github.com/bowouhuy/bookstore-api/pkg/routerkit"
)

type handlerFunc func(req *http.Request) dto.HTTPResponse

func (f handlerFunc) Handle(req *http.Request) dto.HTTPResponse { return f(req) }

func newTestRouter(t *testing.T) *router {
	t.Helper()
	require.NoError(t, msg.Load(config.Messages))

	return &router{
		cfg:    &config.Config{App: config.App{DefaultLang: "en"}},
		router: routerkit.NewRouter(),
	}
}

func serve(r *router, h handlerFunc, req *http.Request, mdws ...middleware.Handle) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.handle(httpGateway, h, mdws...)(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	return body
}

func TestHandle_WritesEnvelopeAndHeaders(t *testing.T) {
	r := newTestRouter(t)
	h := handlerFunc(func(*http.Request) dto.HTTPResponse {
		return *dto.NewResponse().WithCode(http.StatusCreated).WithHeader("Location", "/books/7").WithData(map[string]int{"id": 7})
	})

	rec := serve(r, h, httptest.NewRequest(http.MethodPost, "/books", nil))

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, "/books/7", rec.Header().Get("Location"))

	body := decode(t, rec)
	assert.EqualValues(t, 201, body["code"])
	assert.Equal(t, "Created", body["message"])
	assert.Equal(t, map[string]any{"id": float64(7)}, body["data"])
}

func TestHandle_TranslatesMessage(t *testing.T) {
	r := newTestRouter(t)
	h := handlerFunc(func(*http.Request) dto.HTTPResponse { return *dto.NewResponse().WithCode(http.StatusNotFound) })

	req := httptest.NewRequest(http.MethodGet, "/books/9", nil)
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9")

	body := decode(t, serve(r, h, req))
	assert.Equal(t, "Data tidak ditemukan", body["message"])
}

func TestHandle_NoContentHasNoBody(t *testing.T) {
	r := newTestRouter(t)
	h := handlerFunc(func(*http.Request) dto.HTTPResponse { return *dto.NewResponse().WithCode(http.StatusNoContent) })

	rec := serve(r, h, httptest.NewRequest(http.MethodDelete, "/books/1", nil))

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
}

func TestHandle_RecoversFromPanic(t *testing.T) {
	r := newTestRouter(t)
	h := handlerFunc(func(*http.Request) dto.HTTPResponse {
		var m map[string]int
		m["boom"] = 1 // writing to a nil map panics
		return dto.HTTPResponse{}
	})

	rec := serve(r, h, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.EqualValues(t, 500, decode(t, rec)["code"])
}

func TestHandle_MiddlewareErrorStopsRequest(t *testing.T) {
	r := newTestRouter(t)
	called := false
	h := handlerFunc(func(*http.Request) dto.HTTPResponse {
		called = true
		return *dto.NewResponse().WithCode(http.StatusOK)
	})
	deny := func(req *http.Request) (*http.Request, error) {
		return req, &middleware.Error{Code: http.StatusUnauthorized, Message: "nope", Headers: map[string]string{"WWW-Authenticate": "Bearer"}}
	}
	broken := func(req *http.Request) (*http.Request, error) { return req, errors.New("boom") }

	rec := serve(r, h, httptest.NewRequest(http.MethodGet, "/auth/me", nil), deny)
	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"))
	assert.Equal(t, "nope", decode(t, rec)["message"])

	rec = serve(r, h, httptest.NewRequest(http.MethodGet, "/auth/me", nil), broken)
	assert.Equal(t, http.StatusInternalServerError, rec.Code, "unknown middleware errors must not leak details")
}

func TestFallback_ReturnsJSON(t *testing.T) {
	r := newTestRouter(t)
	r.router.NotFoundHandler = r.fallback(http.StatusNotFound)
	r.router.MethodNotAllowedHandler = r.fallback(http.StatusMethodNotAllowed)
	r.router.PathPrefix("/").Subrouter().HandleFunc("/books", func(http.ResponseWriter, *http.Request) {}).Methods(http.MethodGet)

	rec := httptest.NewRecorder()
	r.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	rec = httptest.NewRecorder()
	r.router.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/books", nil))
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}
