/*
 * Copyright © 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/3/3 13:18
 */

// Package oss
package oss

import "github.com/minio/minio-go/v7"

type (
	DownloadOptions   minio.GetObjectOptions
	RemoveOptions     minio.RemoveObjectOptions
	UploadOptions     minio.PutObjectOptions
	StatObjectOptions minio.StatObjectOptions
)
