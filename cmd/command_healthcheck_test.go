package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckLiveness(t *testing.T) {
	tests := map[string]struct {
		status  int
		wantErr bool
	}{
		"healthy":            {status: http.StatusOK},
		"database down":      {status: http.StatusServiceUnavailable, wantErr: true},
		"unexpected failure": {status: http.StatusInternalServerError, wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/liveness", r.URL.Path)
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			err := checkLiveness(srv.URL + "/liveness")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckLiveness_ServerUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL + "/liveness"
	srv.Close()

	assert.Error(t, checkLiveness(url))
}
