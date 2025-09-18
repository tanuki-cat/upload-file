package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"upload-util/internal/config"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

type HuaweiUploader struct {
	client   *obs.ObsClient
	config   *config.HuaweiOBSConfig
	settings *config.UploadSettings
}

func NewHuaweiOBSUploader(config *config.HuaweiOBSConfig, settings *config.UploadSettings) (*HuaweiUploader, error) {
	if config == nil {
		return nil, fmt.Errorf("huawei obs config is nil")
	}

	// 创建 OBS 客户端
	client, err := obs.New(config.AccessKeyID, config.SecretAccessKey, config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create huawei obs client: %w", err)
	}

	return &HuaweiUploader{
		client:   client,
		config:   config,
		settings: settings,
	}, nil
}

func (u *HuaweiUploader) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	// 验证文件
	if err := validateFile(header, u.settings); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// 生成文件名和对象键
	filename := generateFileName(header, u.settings)
	objectKey := buildObjectKey(filename, u.config.PathPrefix)

	// 上传文件
	input := &obs.PutObjectInput{}
	input.Bucket = u.config.Bucket
	input.Key = objectKey
	input.Body = file
	input.ContentType = getMimeType(filename)

	_, err := u.client.PutObject(input)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to huawei obs: %w", err)
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

func (u *HuaweiUploader) Delete(ctx context.Context, key string) error {
	input := &obs.DeleteObjectInput{}
	input.Bucket = u.config.Bucket
	input.Key = key

	_, err := u.client.DeleteObject(input)
	if err != nil {
		return fmt.Errorf("failed to delete object from huawei obs: %w", err)
	}
	return nil
}

func (u *HuaweiUploader) GetURL(ctx context.Context, key string) (string, error) {
	if u.config.Domain != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(u.config.Domain, "/"), key), nil
	}

	protocol := "https"
	if !u.config.UseSSL {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s.%s/%s", protocol, u.config.Bucket, u.config.Endpoint, key), nil
}
