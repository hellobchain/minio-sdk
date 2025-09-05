package models

// Config MinIO客户端配置
type Config struct {
	Endpoint        string // MinIO服务器地址，格式: hostname:port
	AccessKeyID     string // 访问密钥ID
	SecretAccessKey string // 秘密访问密钥
	UseSSL          bool   // 是否使用SSL
	BucketName      string // 默认存储桶名称
	Region          string // 区域，MinIO通常使用"us-east-1"
}

// UploadOptions 文件上传选项
type UploadOptions struct {
	ContentType  string            // 文件内容类型
	UserMetadata map[string]string // 用户元数据
}

// DownloadOptions 文件下载选项
type DownloadOptions struct {
	VersionID string // 对象版本ID
}

// UploadResult 上传结果
type UploadResult struct {
	ETag         string // 对象的ETag
	VersionID    string // 对象版本ID
	Size         int64  // 文件大小
	LastModified string // 最后修改时间
}

// ObjectInfo 对象信息
type ObjectInfo struct {
	Key          string            // 对象键
	Size         int64             // 对象大小
	LastModified string            // 最后修改时间
	ETag         string            // ETag
	ContentType  string            // 内容类型
	UserMetadata map[string]string // 用户元数据
}
