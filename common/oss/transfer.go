/*
 * Copyright © 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/3/3 13:32
 */

// Package oss
package oss

import (
	"context"
	"github.com/minio/minio-go/v7"
	"io"
	"mime/multipart"
	"net/http"
)

func (m *Manager) Upload(fileName string, reader io.Reader, options UploadOptions) (info *minio.UploadInfo, err error) {
	info = new(minio.UploadInfo)
	*info, err = m.GetClient().PutObject(context.Background(), m.bucket, fileName, reader, -1, minio.PutObjectOptions(options))
	return
}

func (m *Manager) Download(fileName string, options DownloadOptions) (reader io.ReadCloser, err error) {
	return m.GetClient().GetObject(context.Background(), m.bucket, fileName, minio.GetObjectOptions(options))
}

func (m *Manager) Remove(fileName string, options RemoveOptions) (err error) {
	return m.GetClient().RemoveObject(context.Background(), m.bucket, fileName, minio.RemoveObjectOptions(options))
}

func (m *Manager) UploadFromHttp(fileName string, fileHeader *multipart.FileHeader, options UploadOptions) (*minio.UploadInfo, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	return m.Upload(fileName, file, options)
}

func (m *Manager) StatObject(fileName string, options StatObjectOptions) (*minio.ObjectInfo, error) {
	info, err := m.GetClient().StatObject(context.Background(), m.bucket, fileName, minio.StatObjectOptions(options))
	// 文件不存在是正常的，不应返回err
	if minioErr, ok := err.(minio.ErrorResponse); ok && minioErr.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	return &info, err
}
