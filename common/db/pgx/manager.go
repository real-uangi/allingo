// Package pgxdb manages native pgx pools without a database/sql or ORM adapter.
package pgxdb

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

// Open creates and verifies a pool. Config must come from pgxpool.ParseConfig.
// The caller owns the returned pool and must close it after its users stop.
func Open(ctx context.Context, config *pgxpool.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// Manager owns one pool for an Fx application. Construction does no database
// work; the pool becomes available after its OnStart hook succeeds.
type Manager struct {
	pool *pgxpool.Pool
}

func NewManager(lifecycle fx.Lifecycle, config *pgxpool.Config) *Manager {
	manager := &Manager{}
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			pool, err := Open(ctx, config)
			if err != nil {
				return err
			}
			manager.pool = pool
			return nil
		},
		OnStop: func(context.Context) error {
			// Native pgx Close waits for borrowed connections. HTTP handlers and
			// background work must release them before this hook runs.
			manager.pool.Close()
			return nil
		},
	})
	return manager
}

// GetPool returns the native pool after successful application startup.
// Consumers borrow it; only the manager closes it.
func (manager *Manager) GetPool() *pgxpool.Pool {
	return manager.pool
}

var Module = fx.Module("pgxdb", fx.Provide(NewManager))
