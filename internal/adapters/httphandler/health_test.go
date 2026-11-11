package httphandler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bowouhuy/bookstore-api/internal/adapters/httphandler"
	"github.com/bowouhuy/bookstore-api/internal/usecases"
)

func TestHealthHandler(t *testing.T) {
	t.Run("healthy returns version", func(t *testing.T) {
		uc := usecases.NewGoMockHealthUsecase(gomock.NewController(t))
		uc.EXPECT().Liveness(gomock.Any()).Return("Service is healthy", nil)

		resp := httphandler.NewHealthHandler(uc, "abc1234").Handle(httptest.NewRequest(http.MethodGet, "/liveness", nil))

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "Service is healthy", resp.Message)
		data, err := json.Marshal(resp.Data)
		require.NoError(t, err)
		assert.JSONEq(t, `{"version":"abc1234"}`, string(data))
	})

	t.Run("database unreachable", func(t *testing.T) {
		uc := usecases.NewGoMockHealthUsecase(gomock.NewController(t))
		uc.EXPECT().Liveness(gomock.Any()).Return("", errors.New("connection refused"))

		resp := httphandler.NewHealthHandler(uc, "abc1234").Handle(httptest.NewRequest(http.MethodGet, "/liveness", nil))

		assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
		assert.Nil(t, resp.Data, "version must not be reported when unhealthy")
	})
}
