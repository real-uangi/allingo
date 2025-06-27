/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/19 14:36
 */

// Package storage

package storage

import (
	"github.com/real-uangi/allingo/common/env"
	"github.com/real-uangi/allingo/common/oss"
	"io"
)

type S3Storage struct {
	manager *oss.Manager
}

func newS3Storage(endpoint string) (*S3Storage, error) {
	manager, err := oss.NewManager(endpoint, env.GetOrDefault("MINIO_BUCKET", "peerflux"))
	if err != nil {
		return nil, err
	}
	return &S3Storage{
		manager: manager,
	}, nil
}

func (s *S3Storage) Put(key string, reader io.Reader) error {
	_, err := s.manager.Upload(key, reader, oss.UploadOptions{})
	return err
}

func (s *S3Storage) Get(key string) (io.ReadCloser, error) {
	return s.manager.Download(key, oss.DownloadOptions{})
}

func (s *S3Storage) Delete(key string) error {
	return s.manager.Remove(key, oss.RemoveOptions{})
}
