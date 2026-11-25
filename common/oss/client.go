package oss

import (
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/fx"
)

// ClientConfig configures an S3 client independently of environment variables
// and bucket selection. A nil Transport uses the MinIO client's default.
type ClientConfig struct {
	Endpoint     string
	AccessKey    string
	SecretKey    string
	SessionToken string
	Secure       bool
	Region       string
	BucketLookup minio.BucketLookupType
	Transport    http.RoundTripper
}

// NewClient constructs a native client without making requests or creating
// buckets. Callers use its context-aware APIs with explicit bucket names.
func NewClient(config ClientConfig) (*minio.Client, error) {
	return minio.New(config.Endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(config.AccessKey, config.SecretKey, config.SessionToken),
		Secure:       config.Secure,
		Region:       config.Region,
		BucketLookup: config.BucketLookup,
		Transport:    config.Transport,
	})
}

var Module = fx.Module("oss", fx.Provide(NewClient))
