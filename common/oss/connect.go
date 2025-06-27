/*
 * Copyright © 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/3/3 13:15
 */

// Package oss
package oss

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/real-uangi/allingo/common/env"
	"github.com/real-uangi/allingo/common/log"
	"net/http"
	"time"
)

type Manager struct {
	client *minio.Client
	logger *log.StdLogger
	bucket string
}

func NewManager(endpoint string, bucketName string) (*Manager, error) {
	m := &Manager{
		logger: log.For[Manager](),
		bucket: bucketName,
	}
	m.logger.Infof("initializing minio client")

	options := minio.Options{
		Creds:  credentials.NewStaticV4(env.Get("MINIO_ID"), env.Get("MINIO_SECRET"), env.Get("MINIO_TOKEN")),
		Secure: false,
		Transport: &http.Transport{
			Proxy:             nil,
			DisableKeepAlives: false,
			MaxIdleConns:      10,
			IdleConnTimeout:   180 * time.Second,
		},
	}
	minioClient, err := minio.New(endpoint, &options)
	if err != nil {
		return nil, err
	}
	m.client = minioClient
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

func (m *Manager) GetClient() *minio.Client {
	return m.client
}
