package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/real-uangi/allingo/common/log"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

type startObserver struct{ started chan struct{} }

func (observer startObserver) LogEvent(event fxevent.Event) {
	if event, ok := event.(*fxevent.Started); ok && event.Err == nil {
		close(observer.started)
	}
}

func TestRunContextCancellationPreservesStopBudget(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{})
	stopped := make(chan time.Duration, 1)
	app := &App{logger: log.For[App]()}
	app.Option(fx.Options(
		fx.WithLogger(func() fxevent.Logger { return startObserver{started} }),
		fx.Invoke(func(lc fx.Lifecycle, running RuntimeContext) {
			lc.Append(fx.Hook{OnStop: func(ctx context.Context) error {
				if !errors.Is(running.Context.Err(), context.Canceled) {
					return errors.New("runtime was not canceled before the first stop hook")
				}
				if err := ctx.Err(); err != nil {
					return err
				}
				deadline, _ := ctx.Deadline()
				stopped <- time.Until(deadline)
				return nil
			}})
		}),
	))
	done := make(chan error, 1)
	go func() { done <- app.RunContext(ctx) }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("application did not start")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("application did not stop after cancellation")
	}
	remaining := <-stopped
	if remaining <= 14*time.Second || remaining > 15*time.Second {
		t.Fatalf("expected fresh 15-second stop budget, got %s", remaining)
	}
}

func TestRunContextObservesFxShutdownAndStopError(t *testing.T) {
	stopErr := errors.New("cleanup failed")
	app := &App{logger: log.For[App]()}
	app.Option(fx.Options(fx.NopLogger, fx.Invoke(func(lc fx.Lifecycle, shutdowner fx.Shutdowner) {
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error { return shutdowner.Shutdown(fx.ExitCode(7)) },
			OnStop:  func(context.Context) error { return stopErr },
		})
	})))
	err := app.RunContext(context.Background())
	if !errors.Is(err, stopErr) || !strings.Contains(err.Error(), "exit code 7") {
		t.Fatalf("expected shutdown code and stop failure, got %v", err)
	}
}

func TestRunContextHonorsFxStartTimeout(t *testing.T) {
	app := &App{logger: log.For[App]()}
	app.Option(fx.Options(fx.NopLogger, fx.StartTimeout(20*time.Millisecond), fx.Invoke(func(lc fx.Lifecycle) {
		lc.Append(fx.Hook{OnStart: func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }})
	})))
	if err := app.RunContext(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected startup deadline, got %v", err)
	}
}

func TestRunContextHonorsFxStopTimeout(t *testing.T) {
	app := &App{logger: log.For[App]()}
	app.Option(fx.Options(fx.NopLogger, fx.StopTimeout(20*time.Millisecond), fx.Invoke(func(lc fx.Lifecycle, shutdowner fx.Shutdowner) {
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error { return shutdowner.Shutdown() },
			OnStop:  func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() },
		})
	})))
	if err := app.RunContext(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected stop deadline, got %v", err)
	}
}

func TestFxShutdownCancelsRuntimeBeforeFirstStopConsumesBudget(t *testing.T) {
	observed := make(chan error, 1)
	stopFinished := make(chan struct{})
	application := &App{logger: log.For[App]()}
	application.Option(fx.Options(fx.NopLogger, fx.StopTimeout(20*time.Millisecond), fx.Invoke(func(lc fx.Lifecycle, running RuntimeContext, shutdowner fx.Shutdowner) {
		// This stands in for the first HTTP drain hook: it consumes the entire
		// shared budget. Runtime cancellation must precede it, not a later hook.
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error { return shutdowner.Shutdown() },
			OnStop: func(ctx context.Context) error {
				defer close(stopFinished)
				observed <- running.Context.Err()
				<-ctx.Done()
				return ctx.Err()
			},
		})
	})))
	if err := application.RunContext(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected exhausted graceful-stop budget, got %v", err)
	}
	if err := <-observed; !errors.Is(err, context.Canceled) {
		t.Fatalf("runtime was still active when HTTP drain started: %v", err)
	}
	select {
	case <-stopFinished:
	case <-time.After(time.Second):
		t.Fatal("deadline-aware stop hook did not finish")
	}
}

func TestRunContextCancelsSuppliedRuntimeWhenStartupFails(t *testing.T) {
	startErr := errors.New("startup failed")
	var running context.Context
	application := &App{logger: log.For[App]()}
	application.Option(fx.Options(fx.NopLogger, fx.Invoke(func(lc fx.Lifecycle, lifetime RuntimeContext) {
		running = lifetime.Context
		lc.Append(fx.Hook{OnStart: func(context.Context) error { return startErr }})
	})))
	if err := application.RunContext(context.Background()); !errors.Is(err, startErr) {
		t.Fatalf("expected startup failure, got %v", err)
	}
	if !errors.Is(running.Err(), context.Canceled) {
		t.Fatal("failed startup left the runtime context active")
	}
}
