package common

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/fx"
)

func TestModuleBindsOnStartAndReportsOccupiedAddress(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	app := fx.New(fx.NopLogger, fx.Supply(&HTTPConfig{Addr: occupied.Addr().String()}), Module)
	if err := app.Err(); err != nil {
		t.Fatal(err)
	}
	if err := app.Start(context.Background()); err == nil {
		_ = app.Stop(context.Background())
		t.Fatal("expected listener bind failure during startup")
	}
}

func TestExplicitHTTPConfigUsesHandlerAndDisablesImplicitPprof(t *testing.T) {
	t.Setenv("RUN_MODE", "debug")
	config := HTTPConfig{
		Addr: "127.0.0.1:0", ReadHeaderTimeout: time.Second, ReadTimeout: time.Second,
		WriteTimeout: time.Second, IdleTimeout: time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, r.Host) }),
	}
	runtime, _ := startTestHTTP(t, config)
	if runtime.profile != nil {
		t.Fatal("explicit config unexpectedly enabled pprof")
	}
	request, err := http.NewRequest(http.MethodGet, "http://"+runtime.listener.Addr().String()+"/", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Host = "portal.example.com"
	client := &http.Client{Timeout: time.Second}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "portal.example.com" {
		t.Fatalf("custom handler lost Host: %q", body)
	}
}

func TestHTTPShutdownTimeoutCancelsActiveRequest(t *testing.T) {
	started := make(chan struct{})
	cancelled := make(chan struct{})
	runtime, _ := startTestHTTP(t, HTTPConfig{Addr: "127.0.0.1:0", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
		close(cancelled)
	})})
	done := make(chan error, 1)
	go func() {
		client := &http.Client{Timeout: time.Second}
		response, err := client.Get("http://" + runtime.listener.Addr().String())
		if response != nil {
			_ = response.Body.Close()
		}
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("request did not enter handler")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := runtime.stop(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected shutdown timeout, got %v", err)
	}
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("active request was not cancelled")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("client connection remained open")
	}
}

func TestHTTPStopClosesMainAndPprofListeners(t *testing.T) {
	runtime, _ := startTestHTTP(t, HTTPConfig{Addr: "127.0.0.1:0", PprofAddr: "127.0.0.1:0"})
	addresses := []string{runtime.listener.Addr().String(), runtime.profileListener.Addr().String()}
	if err := runtime.stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, address := range addresses {
		connection, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			_ = connection.Close()
			t.Fatalf("listener still open: %s", address)
		}
	}
}

func TestHTTPUnexpectedServeErrorRequestsFxShutdown(t *testing.T) {
	runtime, app := startTestHTTP(t, HTTPConfig{Addr: "127.0.0.1:0"})
	wait := app.Wait()
	if err := runtime.listener.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case signal := <-wait:
		if signal.ExitCode != 1 {
			t.Fatalf("expected failure exit code, got %d", signal.ExitCode)
		}
	case <-time.After(time.Second):
		t.Fatal("serve failure did not request application shutdown")
	}
}

func TestHTTPValidatesTLSBeforeListening(t *testing.T) {
	runtime, err := newHTTPRuntime(HTTPConfig{Addr: "127.0.0.1:0", TLSCertFile: filepath.Join(t.TempDir(), "missing.crt"), TLSKeyFile: "missing.key"}, http.NotFoundHandler(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.start(context.Background()); err == nil {
		t.Fatal("expected missing TLS certificate error")
	}
	if runtime.listener != nil {
		t.Fatal("listener opened before TLS validation")
	}
}

func startTestHTTP(t *testing.T, config HTTPConfig) (*httpRuntime, *fx.App) {
	t.Helper()
	var shutdowner fx.Shutdowner
	app := fx.New(fx.NopLogger, fx.Populate(&shutdowner))
	if err := app.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := app.Stop(context.Background()); err != nil {
			t.Error(err)
		}
	})
	runtime, err := newHTTPRuntime(config, http.NotFoundHandler(), shutdowner)
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := runtime.stop(ctx); err != nil {
			t.Error(err)
		}
	})
	return runtime, app
}
