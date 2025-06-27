/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/16 17:07
 */

// Package inj

package app

import (
	"context"
	"github.com/real-uangi/allingo/common/async"
	"github.com/real-uangi/allingo/common/log"
	"go.uber.org/fx"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type App struct {
	mu        sync.Mutex
	booted    bool
	fxOps     []fx.Option
	onFxStart []func()
	onFxStop  []func()
	logger    *log.StdLogger
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

	fxApp := fx.New(app.fxOps...)
	err := fxApp.Start(context.Background())
	if err != nil {
		panic(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	app.logger.Info("shutting down ...")
	timeoutCtx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFunc()
	err = fxApp.Stop(timeoutCtx)
	if err != nil {
		app.logger.Errorf(err, "failed to stop fx container")
	}

	async.ExitTimeout(5)
}
