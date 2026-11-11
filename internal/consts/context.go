package consts

import "github.com/bowouhuy/bookstore-api/pkg/logger"

type contextIntKey int

const (
	// CtxUserID holds the authenticated user's ID.
	CtxUserID contextIntKey = iota
	// CtxLang const
	CtxLang
)

// CtxAccess holds request fields that logger.*WithContext adds to each entry.
var CtxAccess = logger.AccessContextKey{}
