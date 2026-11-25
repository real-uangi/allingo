package common

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/real-uangi/allingo/common/env"
	"go.uber.org/fx"
)

// HTTPConfig is optionally supplied to Module. An explicit config never enables
// pprof implicitly; an empty PprofAddr disables it. Handler defaults to Gin.
type HTTPConfig struct {
	Addr              string
	Handler           http.Handler
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	TLSCertFile       string
	TLSKeyFile        string
	PprofAddr         string
}

type httpServerParams struct {
	fx.In
	Lifecycle  fx.Lifecycle
	Engine     *gin.Engine
	Shutdowner fx.Shutdowner
	Config     *HTTPConfig `optional:"true"`
}

type httpRuntime struct {
	config          HTTPConfig
	server          *http.Server
	profile         *http.Server
	listener        net.Listener
	profileListener net.Listener
	shutdowner      fx.Shutdowner
}

func legacyHTTPConfig() HTTPConfig {
	config := HTTPConfig{Addr: ":" + env.GetOrDefault("HTTP_PORT", "8080")}
	cert, key := env.Get("HTTP_CERT_FILE"), env.Get("HTTP_KEY_FILE")
	if cert != "" && key != "" {
		config.TLSCertFile, config.TLSKeyFile = cert, key
	}
	if env.Get(env.RunningMode) != env.ReleaseMode {
		config.PprofAddr = ":" + env.GetOrDefault("PPROF_PORT", "18080")
	}
	return config
}

func newHTTPRuntime(config HTTPConfig, defaultHandler http.Handler, shutdowner fx.Shutdowner) (*httpRuntime, error) {
	if config.Addr == "" {
		return nil, errors.New("HTTP address is required")
	}
	if config.ReadHeaderTimeout < 0 || config.ReadTimeout < 0 || config.WriteTimeout < 0 || config.IdleTimeout < 0 {
		return nil, errors.New("HTTP timeouts must not be negative")
	}
	if (config.TLSCertFile == "") != (config.TLSKeyFile == "") {
		return nil, errors.New("both HTTP TLS certificate and key are required")
	}
	handler := config.Handler
	if handler == nil {
		handler = defaultHandler
	}
	runtime := &httpRuntime{config: config, shutdowner: shutdowner, server: &http.Server{
		Addr: config.Addr, Handler: handler, ReadHeaderTimeout: config.ReadHeaderTimeout,
		ReadTimeout: config.ReadTimeout, WriteTimeout: config.WriteTimeout, IdleTimeout: config.IdleTimeout,
	}}
	if config.PprofAddr != "" {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		runtime.profile = &http.Server{Addr: config.PprofAddr, Handler: mux, ReadHeaderTimeout: config.ReadHeaderTimeout}
	}
	return runtime, nil
}

func (runtime *httpRuntime) start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if runtime.config.TLSCertFile != "" {
		cert, err := tls.LoadX509KeyPair(runtime.config.TLSCertFile, runtime.config.TLSKeyFile)
		if err != nil {
			return fmt.Errorf("load HTTP TLS certificate: %w", err)
		}
		runtime.server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", runtime.config.Addr)
	if err != nil {
		return fmt.Errorf("listen HTTP: %w", err)
	}
	if runtime.profile != nil {
		profileListener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", runtime.profile.Addr)
		if err != nil {
			_ = listener.Close()
			return fmt.Errorf("listen pprof: %w", err)
		}
		runtime.profileListener = profileListener
	}
	runtime.listener = listener
	go runtime.serve(runtime.server, listener, runtime.config.TLSCertFile != "")
	if runtime.profile != nil {
		go runtime.serve(runtime.profile, runtime.profileListener, false)
	}
	httpLogger.Infof("http server listening on %s", listener.Addr())
	return nil
}

func (runtime *httpRuntime) serve(server *http.Server, listener net.Listener, secure bool) {
	var err error
	if secure {
		err = server.ServeTLS(listener, "", "")
	} else {
		err = server.Serve(listener)
	}
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return
	}
	httpLogger.Error(err, "http server failed")
	if err := runtime.shutdowner.Shutdown(fx.ExitCode(1)); err != nil {
		httpLogger.Error(err, "failed to request application shutdown")
	}
}

func (runtime *httpRuntime) stop(ctx context.Context) error {
	var profileErr error
	if runtime.profile != nil {
		// Profiling requests need not consume the application's HTTP drain budget.
		profileErr = runtime.profile.Close()
		_ = runtime.profileListener.Close()
	}
	err := runtime.server.Shutdown(ctx)
	// OnStop may run before the Serve goroutine has registered its listener.
	_ = runtime.listener.Close()
	if err != nil {
		// Shutdown does not cancel active requests when its deadline expires.
		// Close their connections before dependent stores begin stopping.
		return errors.Join(err, runtime.server.Close(), profileErr)
	}
	return profileErr
}
