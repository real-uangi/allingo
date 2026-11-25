/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/16 17:07
 */

// Package inj

package app

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/real-uangi/allingo/common/async"
	"github.com/real-uangi/allingo/common/log"
	"go.uber.org/fx"
)

type App struct {
	mu        sync.Mutex
	booted    bool
	fxOps     []fx.Option
	onFxStart []func()
	onFxStop  []func()
	logger    *log.StdLogger
}

// RuntimeContext is supplied by RunContext for long-lived application work.
// It is canceled before any stop hook, independently of the graceful-stop budget.
type RuntimeContext struct {
	Context context.Context
}

func (app *App) mustNotBooted() {
	if app.booted {
		panic("can't modify since app is booted")
	}
}

func (app *App) Option(option fx.Option) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.mustNotBooted()
	app.fxOps = append(app.fxOps, option)
}

func (app *App) Provide(constructors ...interface{}) {
	app.Option(fx.Provide(constructors...))
}

func (app *App) Invoke(functions ...interface{}) {
	app.Option(fx.Invoke(functions...))
}

func (app *App) ProvideAnnotated(annotated fx.Annotated) {
	app.Option(fx.Provide(annotated))
}

func (app *App) Run() {
	if err := app.RunContext(context.Background()); err != nil {
		panic(err)
	}
}

// RunContext waits for cancellation, an operating-system signal, or Fx Shutdown.
// StartTimeout and StopTimeout use the existing Fx options (15 seconds by default).
// Shutdown gets a fresh context so cancelling the run still permits graceful cleanup.
func (app *App) RunContext(ctx context.Context) error {
	app.mu.Lock()
	if app.booted {
		app.mu.Unlock()
		return errors.New("application already started")
	}
	app.booted = true
	options := append([]fx.Option(nil), app.fxOps...)
	app.mu.Unlock()
	runningContext, cancelRun := context.WithCancel(ctx)
	options = append(options, fx.Supply(RuntimeContext{Context: runningContext}))
	defer func() {
		cancelRun()
		async.ExitTimeout(5)
		log.ExitTimeout(5)
	}()
	fxApp := fx.New(options...)
	startCtx, cancelStart := context.WithTimeout(runningContext, fxApp.StartTimeout())
	err := fxApp.Start(startCtx)
	cancelStart()
	if err != nil {
		cancelRun()
		return err
	}
	var shutdown fx.ShutdownSignal
	select {
	case <-ctx.Done():
	case shutdown = <-fxApp.Wait():
	}
	// A long HTTP drain can exhaust Fx's shared stop budget and prevent later
	// hooks from running. Broadcast cancellation before starting that drain.
	cancelRun()
	app.logger.Info("shutting down ...")
	timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), fxApp.StopTimeout())
	defer cancelFunc()
	stopErr := fxApp.Stop(timeoutCtx)
	if shutdown.ExitCode != 0 {
		return errors.Join(fmt.Errorf("application requested exit code %d", shutdown.ExitCode), stopErr)
	}
	return stopErr
}
