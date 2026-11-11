package datastore

import (
	"context"
	"io"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bowouhuy/bookstore-api/config"
	"github.com/bowouhuy/bookstore-api/pkg/di"
	"github.com/bowouhuy/bookstore-api/pkg/logger"
)

var (
	poolMasterOnce sync.Once
	poolSlaveOnce  sync.Once
	poolMaster     *pgxpool.Pool
	poolSlave      *pgxpool.Pool
)

type wrapPool struct {
	pool *pgxpool.Pool
}

func (w *wrapPool) Close() error {
	w.pool.Close()
	return nil
}

// NewBaseRepo returns a base repository.
func NewBaseRepo(dbMaster, dbSlave *pgxpool.Pool) *BaseRepo {
	return &BaseRepo{dbMaster: dbMaster, dbSlave: dbSlave}
}

// BaseRepo holds the master pool for writes and the slave pool for reads.
type BaseRepo struct {
	dbMaster *pgxpool.Pool
	dbSlave  *pgxpool.Pool
}

// getConnString escapes credentials so passwords with special characters stay valid.
func getConnString(cfg *config.DB) string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:     cfg.DB,
		RawQuery: url.Values{"sslmode": {cfg.SSLMode}}.Encode(),
	}

	return u.String()
}

// GetDatabaseMaster returns postgresql Pool for Master.
func GetDatabaseMaster(cfg *config.DB) *pgxpool.Pool {
	poolMasterOnce.Do(func() {
		poolMaster = connect(cfg, "PostgreSQL Master Connection")
	})

	return poolMaster
}

// GetDatabaseSlave returns postgresql Pool for Slave.
func GetDatabaseSlave(cfg *config.DB) *pgxpool.Pool {
	poolSlaveOnce.Do(func() {
		poolSlave = connect(cfg, "PostgreSQL Slave Connection")
	})

	return poolSlave
}

// connect exits the process on failure; errors never include the password.
func connect(cfg *config.DB, name string) *pgxpool.Pool {
	poolCfg, err := pgxpool.ParseConfig(getConnString(cfg))
	if err != nil {
		logger.Fatal(logger.MessageFormat("%s: invalid database config for %s:%d", name, cfg.Host, cfg.Port))
	}

	poolCfg.MaxConns = 10
	poolCfg.MinConns = 2
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute
	poolCfg.HealthCheckPeriod = time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		logger.Fatal(logger.MessageFormat("%s: failed to create pool: %v", name, err))
	}

	if err := pool.Ping(ctx); err != nil {
		logger.Fatal(logger.MessageFormat("%s: failed to ping %s:%d: %v", name, cfg.Host, cfg.Port, err))
	}

	var c io.Closer = &wrapPool{pool: pool}
	di.RegisterCloser(name, c)

	return pool
}
