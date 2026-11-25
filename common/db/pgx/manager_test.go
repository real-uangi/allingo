package pgxdb

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestModuleConstructionDoesNotOpenPool(t *testing.T) {
	config, err := pgxpool.ParseConfig("postgres://user:password@127.0.0.1:1/test?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	var manager *Manager
	app := fx.New(Module, fx.Supply(config), fx.Populate(&manager), fx.NopLogger)
	if err := app.Err(); err != nil {
		t.Fatal(err)
	}
	if manager == nil || manager.GetPool() != nil {
		t.Fatal("construction must provide the manager without creating a pool")
	}
}

func TestOpenCanceledContextDoesNotReturnUnverifiedPool(t *testing.T) {
	config, err := pgxpool.ParseConfig("postgres://user:password@127.0.0.1:1/test?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	pool, err := Open(ctx, config)
	if pool != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("Open canceled = (%v, %v), want nil and context.Canceled", pool, err)
	}
}

func TestNativePoolLifecycle(t *testing.T) {
	dsn := os.Getenv("ALLINGO_PGX_TEST_DSN")
	if dsn == "" {
		t.Skip("ALLINGO_PGX_TEST_DSN is not configured; native PostgreSQL lifecycle is unverified")
	}
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	lifecycle := fxtest.NewLifecycle(t)
	manager := NewManager(lifecycle, config)
	ctx := t.Context()
	if err := lifecycle.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = lifecycle.Stop(context.Background()) })
	pool := manager.GetPool()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var number int
	if err := tx.QueryRow(ctx, "SELECT 1").Scan(&number); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal(err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if number != 1 {
		t.Fatalf("native transaction result = %d", number)
	}
	if err := lifecycle.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pool.Ping(ctx); err == nil {
		t.Fatal("stopped pool still accepts work")
	}
}
