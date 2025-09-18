package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"upload-util/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOUploader struct {
	client   *minio.Client
	config   *config.MinioConfig
	settings *config.UploadSettings
}

func NewMinIOUploader(config *config.MinioConfig, settings *config.UploadSettings) (*MinIOUploader, error) {
	if config == nil {
		return nil, fmt.Errorf("minio config is nil")
	}

	// 创建 MinIO 客户端
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &MinIOUploader{
		client:   client,
		config:   config,
		settings: settings,
	}, nil
}

func (u *MinIOUploader) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	// 验证文件
	if err := validateFile(header, u.settings); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// 生成文件名和对象键
	filename := generateFileName(header, u.settings)
	objectKey := buildObjectKey(filename, u.config.PathPrefix)

	// 上传文件
	info, err := u.client.PutObject(ctx, u.config.Bucket, objectKey, file, header.Size, minio.PutObjectOptions{
		ContentType: getMimeType(filename),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to minio: %w", err)
	}

	// 生成访问 URL
	url, err := u.GetURL(ctx, objectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get url: %w", err)
	}

	return &UploadResult{
		URL:      url,
		Key:      objectKey,
		Size:     info.Size,
		MimeType: getMimeType(filename),
	}, nil
}

func (u *MinIOUploader) Delete(ctx context.Context, key string) error {
	err := u.client.RemoveObject(ctx, u.config.Bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object from minio: %w", err)
	}
	return nil
}

func (u *MinIOUploader) GetURL(ctx context.Context, key string) (string, error) {
	if u.config.Domain != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(u.config.Domain, "/"), key), nil
	}

	protocol := "https"
	if !u.config.UseSSL {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, u.config.Endpoint, u.config.Bucket, key), nil
}
