/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/17 10:18
 */

// Package main

package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/real-uangi/allingo/common/api"
	"github.com/real-uangi/allingo/common/auth"
	"github.com/real-uangi/allingo/common/debug"
	"github.com/real-uangi/allingo/common/env"
	"github.com/real-uangi/allingo/common/log"
	"github.com/real-uangi/allingo/common/ready"
	"github.com/real-uangi/allingo/common/result"
	"github.com/real-uangi/allingo/common/trace"
	"github.com/real-uangi/fxtrategy"
	"go.uber.org/fx"
	"net"
	"net/http"
	"os"
	"strings"
)

var httpLogger = log.NewStdLogger("http")

func startHttpServer(lc fx.Lifecycle, engine *gin.Engine) {

	addr := fmt.Sprintf(":%s", env.GetOrDefault("HTTP_PORT", "8080"))
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	certFile := env.Get("HTTP_CERT_FILE")
	keyFile := env.Get("HTTP_KEY_FILE")

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if certFile != "" && keyFile != "" {
				go func() {
					err := srv.ListenAndServeTLS(certFile, keyFile)
					if err != nil {
						if err == http.ErrServerClosed {
							httpLogger.Warn("http server closed")
						} else {
							httpLogger.Errorf(err, "http server failed: %+v", err)
						}
					}
				}()
				httpLogger.Infof("http server started on https://0.0.0.0%s", srv.Addr)
			} else {
				go func() {
					err := srv.ListenAndServe()
					if err != nil {
						if err == http.ErrServerClosed {
							httpLogger.Warn("http server closed")
						} else {
							httpLogger.Error(err, "http server failed: "+err.Error())
						}
					}
				}()
				httpLogger.Infof("http server started on port %s", srv.Addr)
			}
			if env.Get(env.RunningMode) != env.ReleaseMode {
				pprofPort := env.GetOrDefault("PPROF_PORT", "18080")
				httpLogger.Infof("pprof server started on http://127.0.0.1:%s/debug/pprof", pprofPort)
				go debug.RunPprofHttpServer(pprofPort)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := srv.Shutdown(ctx); err != nil {
				httpLogger.Error(err, "failed to close http server")
			}
			return nil
		},
	})
}

func initGinEngine() *gin.Engine {
	if env.Get(env.RunningMode) != gin.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	engine.Use(recoverMiddleware)

	return engine
}

func recoverMiddleware(c *gin.Context) {
	defer ginRecover(c)
	c.Next()
}

func ginRecover(c *gin.Context) {
	err := recover()
	if err == nil {
		return
	}

	actualE, isErr := err.(error)
	httpLogger.Errorf(actualE, "HTTP请求处理异常 [%s]: %v\n%s", c.Request.URL.Path, err, trace.Stack(3))

	// Check for a broken connection, as it is not really a
	// condition that warrants a panic stack trace.
	var brokenPipe bool
	if ne, ok := err.(*net.OpError); ok {
		var se *os.SyscallError
		if errors.As(ne, &se) {
			seStr := strings.ToLower(se.Error())
			if strings.Contains(seStr, "broken pipe") ||
				strings.Contains(seStr, "connection reset by peer") {
				brokenPipe = true
			}
		}
	}
	if brokenPipe {
		c.Abort()
		return
	}

	if !c.Writer.Written() {
		if isErr {
			c.Render(api.HandleErr(actualE))
			return
		} else {
			c.Render(http.StatusInternalServerError, result.Custom[any](http.StatusInternalServerError, fmt.Sprint(err), nil))
			return
		}
	}

	c.AbortWithStatus(http.StatusInternalServerError)

}

func UsePlatformCloudflare(engine *gin.Engine) {
	engine.TrustedPlatform = gin.PlatformCloudflare
}

type ginCheckPoint struct {
}

func newGinCheckPoint() ready.CheckPoint {
	return &ginCheckPoint{}
}

func (cp *ginCheckPoint) Ready() error {
	return nil
}

func initGinCheckpoint() fxtrategy.Strategy[ready.CheckPoint] {
	return fxtrategy.Strategy[ready.CheckPoint]{
		NS: fxtrategy.NamedStrategy[ready.CheckPoint]{
			Name: "http-gin",
			Item: newGinCheckPoint(),
		},
	}
}

func enableHealthCheckApi(engine *gin.Engine, manager *ready.Manager) {
	engine.GET("/health", auth.InternalOnlyMiddleware, func(c *gin.Context) {
		manager.HandleHealth(c.Writer)
	})
	engine.GET("/health/:target", auth.InternalOnlyMiddleware, func(c *gin.Context) {
		manager.HandleHealthTarget(c.Writer, c.Param("target"))
	})
}
