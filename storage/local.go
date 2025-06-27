/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/19 14:36
 */

// Package storage

package storage

import (
	"errors"
	"fmt"
	"github.com/real-uangi/allingo/common/env"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FileStorage struct {
	baseDir string
}

func newFileStorage() (*FileStorage, error) {
	baseDir := env.GetOrDefault("STORAGE_BASE_DIR", "./storage-data")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	return &FileStorage{
		baseDir: baseDir,
	}, nil
}

func (fs *FileStorage) Put(key string, reader io.Reader) error {
	path, err := fs.safePath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, reader)
	return err
}

func (fs *FileStorage) Get(key string) (io.ReadCloser, error) {
	path, err := fs.safePath(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil // 调用方负责关闭
}

func (fs *FileStorage) Delete(key string) error {
	path, err := fs.safePath(key)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (fs *FileStorage) safePath(key string) (string, error) {
	cleanKey := filepath.Clean(key)
	if strings.HasPrefix(cleanKey, "..") || filepath.IsAbs(cleanKey) {
		return "", fmt.Errorf("非法路径: %s", key)
	}
	fullPath := filepath.Join(fs.baseDir, cleanKey)
	if !strings.HasPrefix(fullPath, fs.baseDir) {
		return "", fmt.Errorf("路径越界: %s", key)
	}
	return fullPath, nil
}
