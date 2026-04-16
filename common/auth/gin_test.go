package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInternalOnlyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("deny external ip without token", func(t *testing.T) {
		t.Setenv(internalTokenEnv, "secret-token")

		w := performInternalOnlyRequest(t, requestArgs{
			remoteAddr: "8.8.8.8:12345",
			host:       "127.0.0.1:8080",
		})

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("allow external ip with valid token", func(t *testing.T) {
		t.Setenv(internalTokenEnv, "secret-token")

		w := performInternalOnlyRequest(t, requestArgs{
			remoteAddr: "8.8.8.8:12345",
			host:       "example.com",
			token:      "secret-token",
		})

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("allow internal ip with valid host", func(t *testing.T) {
		t.Setenv(internalTokenEnv, "")

		w := performInternalOnlyRequest(t, requestArgs{
			remoteAddr: "127.0.0.1:12345",
			host:       "127.0.0.1:8080",
		})

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("deny internal ip with invalid host", func(t *testing.T) {
		t.Setenv(internalTokenEnv, "")

		w := performInternalOnlyRequest(t, requestArgs{
			remoteAddr: "127.0.0.1:12345",
			host:       "example.com",
		})

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

type requestArgs struct {
	remoteAddr string
	host       string
	token      string
}

func performInternalOnlyRequest(t *testing.T, args requestArgs) *httptest.ResponseRecorder {
	t.Helper()

	router := gin.New()
	router.GET("/health", InternalOnlyMiddleware, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = args.remoteAddr
	req.Host = args.host
	if args.token != "" {
		req.Header.Set(internalTokenHeader, args.token)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}
