/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/17 10:18
 */

// Package main

package common

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/real-uangi/allingo/common/api"
	"github.com/real-uangi/allingo/common/auth"
	"github.com/real-uangi/allingo/common/env"
	"github.com/real-uangi/allingo/common/log"
	"github.com/real-uangi/allingo/common/ready"
	"github.com/real-uangi/allingo/common/result"
	"github.com/real-uangi/allingo/common/trace"
	"github.com/real-uangi/fxtrategy"
	"go.uber.org/fx"
)

var httpLogger = log.NewStdLogger("http")

func startHttpServer(params httpServerParams) error {
	config := legacyHTTPConfig()
	if params.Config != nil {
		config = *params.Config
	}
	runtime, err := newHTTPRuntime(config, params.Engine, params.Shutdowner)
	if err != nil {
		return err
	}
	params.Lifecycle.Append(fx.Hook{OnStart: runtime.start, OnStop: runtime.stop})
	return nil
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

func (cp *ginCheckPoint) ItemName() string {
	return "http_gin"
}

var CheckPoint = fxtrategy.ProvideItem[ready.CheckPoint](newGinCheckPoint, ready.CPGroupName)

func enableHealthCheckApi(engine *gin.Engine, manager *ready.Manager) {
	engine.GET("/health", auth.InternalOnlyMiddleware, func(c *gin.Context) {
		manager.HandleHealth(c.Writer)
	})
	engine.GET("/health/:target", auth.InternalOnlyMiddleware, func(c *gin.Context) {
		manager.HandleHealthTarget(c.Writer, c.Param("target"))
	})
}
