package middleware

import (
	"net/http"
)

// Handle is contract for middleware and must implement this type for http if you need middleware http request.
// It returns the request so a middleware can attach context values, e.g. the authenticated user ID.
type Handle func(r *http.Request) (*http.Request, error)

// Error stops the request with the given status, message and headers.
type Error struct {
	Code    int
	Message string
	Headers map[string]string
}

func (e *Error) Error() string {
	return e.Message
}

// FilterFunc is a iterator resolver in each middleware registered
func FilterFunc(r *http.Request, mfs []Handle) (*http.Request, error) {
	for _, mf := range mfs {
		var err error
		if r, err = mf(r); err != nil {
			return r, err
		}
	}

	return r, nil
}
