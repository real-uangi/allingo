/*
 * Copyright © 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/3/3 13:15
 */

// Package oss
package oss

import (
	"context"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/real-uangi/allingo/common/env"
	"github.com/real-uangi/allingo/common/log"
)

type Manager struct {
	client *minio.Client
	logger *log.StdLogger
	bucket string
}

func NewManager(endpoint string, bucketName string) (*Manager, error) {
	logger := log.For[Manager]()
	logger.Infof("initializing minio client")
	minioClient, err := NewClient(ClientConfig{
		Endpoint:     endpoint,
		AccessKey:    env.Get("MINIO_ID"),
		SecretKey:    env.Get("MINIO_SECRET"),
		SessionToken: env.Get("MINIO_TOKEN"),
		Secure:       false,
		Transport: &http.Transport{
			Proxy:             nil,
			DisableKeepAlives: false,
			MaxIdleConns:      10,
			IdleConnTimeout:   180 * time.Second,
		},
	})
	if err != nil {
		return nil, err
	}
	m := NewManagerWithClient(minioClient, bucketName)
	m.logger.Infof("minio client connected to %s", endpoint)
	exist, err := m.client.BucketExists(context.Background(), m.bucket)
	if err != nil {
		return m, err
	}
	if !exist {
		m.logger.Warnf("minio bucket %s doesn't exist", bucketName)
	}
	return m, nil
}

// NewManagerWithClient binds convenience operations to a bucket without
// checking the bucket or taking ownership of the shared client.
func NewManagerWithClient(client *minio.Client, bucketName string) *Manager {
	return &Manager{client: client, logger: log.For[Manager](), bucket: bucketName}
}

func (m *Manager) GetClient() *minio.Client {
	return m.client
}
