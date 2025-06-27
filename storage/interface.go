/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/19 14:35
 */

// Package storage

package storage

import (
	"github.com/real-uangi/allingo/common/env"
	"io"
)

type Storage interface {
	Put(key string, reader io.Reader) error
	Get(key string) (io.ReadCloser, error)
	Delete(key string) error
}

func InitStorage() (Storage, error) {
	minioEndpoint := env.Get("MINIO_ENDPOINT")
	if minioEndpoint != "" {
		return newS3Storage(minioEndpoint)
	}
	return newFileStorage()
}
