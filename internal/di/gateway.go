package di

import (
	"context"
	"net/http"

	"github.com/bowouhuy/bookstore-api/internal/adapters/httphandler"
	"github.com/bowouhuy/bookstore-api/internal/consts"
	"github.com/bowouhuy/bookstore-api/internal/dto"
	"github.com/bowouhuy/bookstore-api/pkg/locales"
	"github.com/bowouhuy/bookstore-api/pkg/msg"
)

// httpGateway stores the X-Lang header in the context, falling back to the
// Accept-Language value that router.handle already resolved.
func httpGateway(request *http.Request, handler httphandler.Handler) dto.HTTPResponse {
	lang := request.Header.Get(consts.HeaderLanguageKey)
	if !msg.GetAvailableLang(http.StatusOK, lang) {
		lang, _ = locales.LanguageFromContext(request.Context())
	}

	ctx := context.WithValue(request.Context(), consts.CtxLang, lang)

	return handler.Handle(request.WithContext(ctx))
}
