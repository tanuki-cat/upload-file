package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"upload-util/internal/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type AliyunUploader struct {
	client   *oss.Client
	bucket   *oss.Bucket
	config   *config.AliyunOSSConfig
	settings *config.UploadSettings
}

func NewAliyunOSSUploader(config *config.AliyunOSSConfig, settings *config.UploadSettings) (*AliyunUploader, error) {
	if config == nil {
		return nil, fmt.Errorf("aliyun oss config is nil")
	}

	client, err := oss.New(config.Endpoint, config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create aliyun oss client: %w", err)
	}
	bucket, err := client.Bucket(config.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to get aliyun oss bucket: %w", err)
	}
	return &AliyunUploader{
		client:   client,
		bucket:   bucket,
		config:   config,
		settings: settings,
	}, nil
}

func (u *AliyunUploader) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	// 验证文件
	if err := validateFile(header, u.settings); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// 生成文件名和对象键
	filename := generateFileName(header, u.settings)
	objectKey := buildObjectKey(filename, u.config.PathPrefix)

	// 设置上传选项
	options := []oss.Option{
		oss.ContentType(getMimeType(filename)),
		oss.ContentLength(header.Size),
	}

	// 上传文件
	err := u.bucket.PutObject(objectKey, file, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to aliyun oss: %w", err)
	}

	// 生成访问 URL
	url, err := u.GetURL(ctx, objectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get url: %w", err)
	}

	return &UploadResult{
		URL:      url,
		Key:      objectKey,
		Size:     header.Size,
		MimeType: getMimeType(filename),
	}, nil

}

func (u *AliyunUploader) Delete(ctx context.Context, key string) error {
	err := u.bucket.DeleteObject(key)
	if err != nil {
		return fmt.Errorf("failed to delete object from aliyun oss: %w", err)
	}
	return nil
}

func (u *AliyunUploader) GetURL(ctx context.Context, key string) (string, error) {
	if u.config.Domain != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(u.config.Domain, "/"), key), nil
	}
	return buildURL("", u.config.Bucket, u.config.Endpoint, key, u.config.UseSSL), nil
}
