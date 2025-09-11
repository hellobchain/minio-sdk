package client

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/hellobchain/minio-sdk/common/errors"
	"github.com/hellobchain/minio-sdk/common/models"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client MinIO客户端
type Client struct {
	client     *minio.Client
	config     *models.Config
	bucketName string
	logger     logger
}

func (c *Client) SetLogger(logger logger) {
	c.logger = logger
}

type logger interface {
	Print(v ...interface{})
}

type defaultLogger struct{}

func (d *defaultLogger) Print(v ...interface{}) {
	fmt.Println(v...)
}

// NewClient 创建新的MinIO客户端
func NewClient(config *models.Config) (*Client, error) {
	client := &Client{logger: &defaultLogger{}}
	client.logger.Print("Creating new MinIO client")
	if config.Endpoint == "" || config.AccessKeyID == "" || config.SecretAccessKey == "" {
		client.logger.Print("Invalid configuration")
		return nil, errors.ErrInvalidConfig
	}

	// 初始化MinIO客户端
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		client.logger.Print("Failed to create minio client:", err)
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	client = &Client{
		client:     minioClient,
		config:     config,
		bucketName: config.BucketName,
		logger:     &defaultLogger{},
	}

	// 如果指定了存储桶，检查或创建存储桶
	if config.BucketName != "" {
		if err := client.EnsureBucketExists(context.Background()); err != nil {
			client.logger.Print("Failed to ensure bucket existence:", err)
			return nil, err
		}
	}

	return client, nil
}

// EnsureBucketExists 确保存储桶存在
func (c *Client) EnsureBucketExists(ctx context.Context) error {
	if c.client == nil {
		return errors.ErrClientNotInitialized
	}

	exists, err := c.client.BucketExists(ctx, c.bucketName)
	if err != nil {
		c.logger.Print("Failed to check bucket existence:", err)
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		region := c.config.Region
		if region == "" {
			region = "us-east-1"
		}
		err = c.client.MakeBucket(ctx, c.bucketName, minio.MakeBucketOptions{Region: region})
		if err != nil {
			c.logger.Print("Failed to create bucket:", err)
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

// UploadFile 上传文件到MinIO
func (c *Client) UploadFile(ctx context.Context, objectName, filePath string, opts ...models.UploadOptions) (*models.UploadResult, error) {
	if c.client == nil {
		return nil, errors.ErrClientNotInitialized
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.logger.Print("File not found:", filePath)
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// 处理上传选项
	var uploadOpts models.UploadOptions
	if len(opts) > 0 {
		uploadOpts = opts[0]
	}

	// 设置MinIO上传选项
	minioPutOpts := minio.PutObjectOptions{}
	if uploadOpts.ContentType != "" {
		minioPutOpts.ContentType = uploadOpts.ContentType
	}
	if uploadOpts.UserMetadata != nil {
		minioPutOpts.UserMetadata = uploadOpts.UserMetadata
	}

	// 执行上传
	info, err := c.client.FPutObject(ctx, c.bucketName, objectName, filePath, minioPutOpts)
	if err != nil {
		c.logger.Print("Failed to upload file:", err)
		return nil, fmt.Errorf("%w: %v", errors.ErrUploadFailed, err)
	}

	return &models.UploadResult{
		ETag:         info.ETag,
		VersionID:    info.VersionID,
		Size:         info.Size,
		LastModified: info.LastModified.Format(time.RFC3339),
	}, nil
}

// UploadFromReader 从io.Reader上传数据
func (c *Client) UploadFromReader(ctx context.Context, objectName string, reader io.Reader, size int64, opts ...models.UploadOptions) (*models.UploadResult, error) {
	if c.client == nil {
		return nil, errors.ErrClientNotInitialized
	}

	var uploadOpts models.UploadOptions
	if len(opts) > 0 {
		uploadOpts = opts[0]
	}

	minioPutOpts := minio.PutObjectOptions{}
	if uploadOpts.ContentType != "" {
		minioPutOpts.ContentType = uploadOpts.ContentType
	}
	if uploadOpts.UserMetadata != nil {
		minioPutOpts.UserMetadata = uploadOpts.UserMetadata
	}

	info, err := c.client.PutObject(ctx, c.bucketName, objectName, reader, size, minioPutOpts)
	if err != nil {
		c.logger.Print("Failed to upload file:", err)
		return nil, fmt.Errorf("%w: %v", errors.ErrUploadFailed, err)
	}

	return &models.UploadResult{
		ETag:         info.ETag,
		VersionID:    info.VersionID,
		Size:         info.Size,
		LastModified: time.Now().Format(time.RFC3339),
	}, nil
}

// DownloadFile 从MinIO下载文件
func (c *Client) DownloadFile(ctx context.Context, objectName, filePath string, opts ...models.DownloadOptions) error {
	if c.client == nil {
		return errors.ErrClientNotInitialized
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.logger.Print("Failed to create directory:", err)
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var downloadOpts models.DownloadOptions
	if len(opts) > 0 {
		downloadOpts = opts[0]
	}

	minioGetOpts := minio.GetObjectOptions{}
	if downloadOpts.VersionID != "" {
		minioGetOpts.VersionID = downloadOpts.VersionID
	}

	// 执行下载
	err := c.client.FGetObject(ctx, c.bucketName, objectName, filePath, minioGetOpts)
	if err != nil {
		c.logger.Print("Failed to download file:", err)
		return fmt.Errorf("%w: %v", errors.ErrDownloadFailed, err)
	}

	return nil
}

// DownloadToWriter 下载文件到io.Writer
func (c *Client) DownloadToWriter(ctx context.Context, objectName string, writer io.Writer, opts ...models.DownloadOptions) error {
	if c.client == nil {
		return errors.ErrClientNotInitialized
	}

	var downloadOpts models.DownloadOptions
	if len(opts) > 0 {
		downloadOpts = opts[0]
	}

	minioGetOpts := minio.GetObjectOptions{}
	if downloadOpts.VersionID != "" {
		minioGetOpts.VersionID = downloadOpts.VersionID
	}

	object, err := c.client.GetObject(ctx, c.bucketName, objectName, minioGetOpts)
	if err != nil {
		c.logger.Print("Failed to get object:", err)
		return fmt.Errorf("%w: %v", errors.ErrDownloadFailed, err)
	}
	defer object.Close()

	_, err = io.Copy(writer, object)
	if err != nil {
		c.logger.Print("Failed to copy object data:", err)
		return fmt.Errorf("failed to copy object data: %w", err)
	}

	return nil
}

// DownloadToMemory 下载文件到memory并返回字节切片
func (c *Client) DownloadToMemory(ctx context.Context, objectName string, writer io.Writer, opts ...models.DownloadOptions) ([]byte, error) {
	if c.client == nil {
		return nil, errors.ErrClientNotInitialized
	}

	var downloadOpts models.DownloadOptions
	if len(opts) > 0 {
		downloadOpts = opts[0]
	}

	minioGetOpts := minio.GetObjectOptions{}
	if downloadOpts.VersionID != "" {
		minioGetOpts.VersionID = downloadOpts.VersionID
	}

	object, err := c.client.GetObject(ctx, c.bucketName, objectName, minioGetOpts)
	if err != nil {
		c.logger.Print("Failed to get object:", err)
		return nil, fmt.Errorf("%w: %v", errors.ErrDownloadFailed, err)
	}
	defer object.Close()
	data, err := io.ReadAll(object)
	if err != nil {
		c.logger.Print("Failed to read object data:", err)
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}
	return data, nil
}

// ObjectExists 检查对象是否存在
func (c *Client) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	if c.client == nil {
		return false, errors.ErrClientNotInitialized
	}

	_, err := c.client.StatObject(ctx, c.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		c.logger.Print("Failed to stat object:", err)
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetObjectInfo 获取对象信息
func (c *Client) GetObjectInfo(ctx context.Context, objectName string) (*models.ObjectInfo, error) {
	if c.client == nil {
		return nil, errors.ErrClientNotInitialized
	}

	info, err := c.client.StatObject(ctx, c.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		c.logger.Print("Failed to get object info:", err)
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	return &models.ObjectInfo{
		Key:          objectName,
		Size:         info.Size,
		LastModified: info.LastModified.Format(time.RFC3339),
		ETag:         info.ETag,
		ContentType:  info.ContentType,
		UserMetadata: info.UserMetadata,
	}, nil
}

// DeleteObject 删除对象
func (c *Client) DeleteObject(ctx context.Context, objectName string) error {
	if c.client == nil {
		return errors.ErrClientNotInitialized
	}

	err := c.client.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		c.logger.Print("Failed to delete object:", err)
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// SetBucket 设置当前操作的存储桶
func (c *Client) SetBucket(bucketName string) error {
	c.bucketName = bucketName
	return c.EnsureBucketExists(context.Background())
}

// GetClient 获取底层的MinIO客户端（用于高级操作）
func (c *Client) GetClient() *minio.Client {
	return c.client
}

func GetMimeType(filePath string) string {
	return mime.TypeByExtension(filepath.Ext(filePath))
}
