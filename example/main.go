package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hellobchain/minio-sdk/common/client"
	"github.com/hellobchain/minio-sdk/common/models"
)

func main() {
	// 配置MinIO客户端
	config := &models.Config{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin123",
		UseSSL:          false,
		BucketName:      "my-bucket",
		Region:          "us-east-1",
	}

	// 创建客户端
	client, err := client.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// 示例1: 上传文件
	fmt.Println("Uploading file...")
	uploadResult, err := client.UploadFile(ctx, "test-file.txt", "./example.txt", models.UploadOptions{
		ContentType: "text/plain",
		UserMetadata: map[string]string{
			"uploaded-by": "demo-app",
		},
	})
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}
	fmt.Printf("Upload successful: %+v\n", uploadResult)

	// 示例2: 检查文件是否存在
	exists, err := client.ObjectExists(ctx, "test-file.txt")
	if err != nil {
		log.Fatalf("Check existence failed: %v", err)
	}
	fmt.Printf("File exists: %v\n", exists)

	// 示例3: 获取文件信息
	info, err := client.GetObjectInfo(ctx, "test-file.txt")
	if err != nil {
		log.Fatalf("Get info failed: %v", err)
	}
	fmt.Printf("File info: %+v\n", info)

	// 示例4: 下载文件
	fmt.Println("Downloading file...")
	err = client.DownloadFile(ctx, "test-file.txt", "./downloaded-file.txt")
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}
	fmt.Println("Download successful")

	// 示例5: 删除文件
	err = client.DeleteObject(ctx, "test-file.txt")
	if err != nil {
		log.Fatalf("Delete failed: %v", err)
	}
	fmt.Println("Delete successful")

	// 清理
	os.Remove("./downloaded-file.txt")
}
