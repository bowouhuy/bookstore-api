package di

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/bowouhuy/bookstore-api/internal/adapters/httphandler"
	"github.com/bowouhuy/bookstore-api/internal/consts"
	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/internal/middleware"
	"github.com/bowouhuy/bookstore-api/pkg/locales"
	"github.com/bowouhuy/bookstore-api/pkg/logger"
)

const maxBodyBytes = 1 << 20 // 1 MB

func (r *router) handle(hfn httpHandlerFunc, handler httphandler.Handler, mdws ...middleware.Handle) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		lang := r.lang(req)
		code := http.StatusInternalServerError

		ctx := context.WithValue(req.Context(), consts.CtxAccess, map[string]interface{}{
			"path":      req.URL.Path,
			"remote_ip": req.RemoteAddr,
			"method":    req.Method,
		})
		ctx = locales.WithAcceptLanguage(ctx, lang)
		req = req.WithContext(ctx)
		req.Body = http.MaxBytesReader(w, req.Body, maxBodyBytes)

		defer func() {
			if rec := recover(); rec != nil {
				// ErrAbortHandler is a deliberate abort; let net/http handle it.
				if rec == http.ErrAbortHandler {
					panic(rec)
				}

				logger.ErrorWithContext(ctx, logger.MessageFormat("panic recovered: %v", rec), logger.String("stack", string(debug.Stack())))
				code = r.response(w, dto.HTTPResponse{Code: http.StatusInternalServerError, Lang: lang})
			}

			accessLog(ctx, code, time.Since(start))
		}()

		var err error
		if req, err = middleware.FilterFunc(req, mdws); err != nil {
			resp := middlewareErrorResponse(ctx, err)
			resp.Lang = lang
			code = r.response(w, resp)

			return
		}

		resp := hfn(req, handler)
		resp.Lang = lang
		code = r.response(w, resp)
	}
}

// response writes resp and returns the status code that was sent.
func (r *router) response(w http.ResponseWriter, resp dto.HTTPResponse) int {
	for key, value := range resp.Headers() {
		w.Header().Set(key, value)
	}

	resp.GenerateMessage()
	code := resp.GetCode()

	if resp.IsStream() {
		w.Header().Set(consts.HeaderContentTypeKey, resp.ContentType())
		w.WriteHeader(code)
		_, _ = w.Write(resp.DataStream())

		return code
	}

	w.Header().Set(consts.HeaderContentTypeKey, consts.HeaderContentTypeJSON)
	w.WriteHeader(code)

	// A 204 response must not have a body.
	if code == http.StatusNoContent {
		return code
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error(logger.MessageFormat("error %v", err))
	}

	return code
}

// fallback answers unmatched routes with the same JSON envelope as other responses.
func (r *router) fallback(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.response(w, dto.HTTPResponse{Code: code, Lang: r.lang(req)})
	})
}

func (r *router) lang(req *http.Request) string {
	lang := r.defaultLang(req.Header.Get(consts.HeaderAcceptLanguage))
	return locales.ParseIOSLang(lang, r.cfg.App.DefaultLang)
}

func (r *router) defaultLang(lang string) string {
	if len(lang) == 0 {
		return r.cfg.App.DefaultLang
	}

	return lang
}

func middlewareErrorResponse(ctx context.Context, err error) dto.HTTPResponse {
	var mErr *middleware.Error
	if !errors.As(err, &mErr) {
		logger.ErrorWithContext(ctx, logger.MessageFormat("middleware error: %v", err))
		return dto.HTTPResponse{Code: http.StatusInternalServerError}
	}

	resp := dto.NewResponse().WithCode(mErr.Code).WithMessage(mErr.Message)
	for key, value := range mErr.Headers {
		resp.WithHeader(key, value)
	}

	return *resp
}

func accessLog(ctx context.Context, code int, duration time.Duration) {
	fields := []logger.Field{logger.Any("status", code), logger.String("duration", duration.String())}

	switch {
	case code >= http.StatusInternalServerError:
		logger.ErrorWithContext(ctx, "http request", fields...)
	case code >= http.StatusBadRequest:
		logger.WarnWithContext(ctx, "http request", fields...)
	default:
		logger.InfoWithContext(ctx, "http request", fields...)
	}
}
