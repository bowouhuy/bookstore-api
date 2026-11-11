package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bowouhuy/bookstore-api/internal/entities/repository"
	"github.com/bowouhuy/bookstore-api/internal/usecases"
)

var errInternal = errors.New("error")

func TestHealthUC_Liveness(t *testing.T) {
	tests := map[string]struct {
		repoErr error
		want    string
		wantErr error
	}{
		"Given healthy database, Return message":   {want: "Service is healthy"},
		"Given unreachable database, Return error": {repoErr: errInternal, wantErr: errInternal},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()

			repo := repository.NewGoMockHealthRepo(ctrl)
			repo.EXPECT().GetLiveness(ctx).Return(tt.repoErr)

			got, err := usecases.NewHealthUsecase(repo).Liveness(ctx)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}
