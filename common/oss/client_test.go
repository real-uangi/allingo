package oss

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func objectResponse(request *http.Request, status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header: http.Header{
			"Content-Length": {"3"},
			"Last-Modified":  {"Wed, 25 Nov 2026 08:00:00 GMT"},
			"Etag":           {`"test-etag"`},
		},
		Body: io.NopCloser(strings.NewReader("")), Request: request,
	}
}

func TestTypedClientConstructsWithoutRequestsAndSupportsMultipleBuckets(t *testing.T) {
	var requests atomic.Int32
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		if request.URL.Scheme != "https" || request.URL.Host != "objects.example.test:9000" {
			t.Errorf("unexpected target: %s", request.URL)
		}
		if request.URL.Path != "/resources/item" && request.URL.Path != "/logs/item" {
			t.Errorf("unexpected bucket path: %s", request.URL.Path)
		}
		if !strings.Contains(request.Header.Get("Authorization"), "/eu-west-1/s3/aws4_request") {
			t.Errorf("configured signing region not applied")
		}
		return objectResponse(request, http.StatusOK), nil
	})
	client, err := NewClient(ClientConfig{Endpoint: "objects.example.test:9000", AccessKey: "access", SecretKey: "secret", Secure: true, Region: "eu-west-1", BucketLookup: minio.BucketLookupPath, Transport: transport})
	if err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 0 {
		t.Fatal("client construction performed network I/O")
	}
	for _, bucket := range []string{"resources", "logs"} {
		manager := NewManagerWithClient(client, bucket)
		if requests.Load() > 0 && bucket == "resources" {
			t.Fatal("binding manager performed network I/O")
		}
		info, err := manager.StatObjectContext(t.Context(), "item", StatObjectOptions{})
		if err != nil || info == nil || info.Size != 3 {
			t.Fatalf("StatObjectContext(%s) = (%v, %v)", bucket, info, err)
		}
		if manager.GetClient() != client {
			t.Fatal("bucket manager replaced shared client")
		}
	}
	if requests.Load() != 2 {
		t.Fatalf("requests = %d, want one per bucket", requests.Load())
	}
}

func TestContextCancellationReachesObjectRequest(t *testing.T) {
	started := make(chan struct{})
	client, err := NewClient(ClientConfig{
		Endpoint: "objects.example.test", Region: "us-east-1", BucketLookup: minio.BucketLookupPath,
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			close(started)
			<-request.Context().Done()
			return nil, request.Context().Err()
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	manager := NewManagerWithClient(client, "resources")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := manager.StatObjectContext(ctx, "item", StatObjectOptions{}); done <- err }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("object request did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled request = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("object request ignored cancellation")
	}
}

func TestLegacyStatStillReturnsNilForMissingObject(t *testing.T) {
	client, err := NewClient(ClientConfig{Endpoint: "objects.example.test", Region: "us-east-1", BucketLookup: minio.BucketLookupPath, Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return objectResponse(request, http.StatusNotFound), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	info, err := NewManagerWithClient(client, "resources").StatObject("missing", StatObjectOptions{})
	if info != nil || err != nil {
		t.Fatalf("missing legacy StatObject = (%v, %v), want nil, nil", info, err)
	}
}
