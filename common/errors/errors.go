package errors

import "errors"

// 定义错误类型
var (
	ErrClientNotInitialized = errors.New("minio client not initialized")
	ErrBucketNotExists      = errors.New("bucket does not exist")
	ErrObjectNotExists      = errors.New("object does not exist")
	ErrUploadFailed         = errors.New("file upload failed")
	ErrDownloadFailed       = errors.New("file download failed")
	ErrInvalidConfig        = errors.New("invalid configuration")
)
