package datastore

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/bowouhuy/bookstore-api/internal/entities/repository"
)

// healthRepo is a health repository.
type healthRepo struct {
	*BaseRepo
}

// NewHealthRepo returns a health repository.
func NewHealthRepo(base *BaseRepo) repository.HealthRepo {
	return &healthRepo{
		BaseRepo: base,
	}
}

// GetLiveness pings master and slave concurrently.
func (r *healthRepo) GetLiveness(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return r.dbMaster.Ping(ctx)
	})

	group.Go(func() error {
		return r.dbSlave.Ping(ctx)
	})

	return group.Wait()
}
