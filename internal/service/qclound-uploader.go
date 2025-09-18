package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"upload-util/internal/config"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type QCloudUploader struct {
	client   *cos.Client
	config   *config.QCloudCOSConfig
	settings *config.UploadSettings
}

func NewQCloudCOSUploader(config *config.QCloudCOSConfig, settings *config.UploadSettings) (*QCloudUploader, error) {
	if config == nil {
		return nil, fmt.Errorf("qcloud cos config is nil")
	}

	// 构建 bucket URL
	protocol := "https"
	if !config.UseSSL {
		protocol = "http"
	}
	bucketURL := fmt.Sprintf("%s://%s", protocol, config.Endpoint)

	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bucket url: %w", err)
	}

	// 创建 COS 客户端
	client := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.AccessKeyID,
			SecretKey: config.SecretAccessKey,
		},
	})

	return &QCloudUploader{
		client:   client,
		config:   config,
		settings: settings,
	}, nil
}

func (u *QCloudUploader) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	// 验证文件
	if err := validateFile(header, u.settings); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// 生成文件名和对象键
	filename := generateFileName(header, u.settings)
	objectKey := buildObjectKey(filename, u.config.PathPrefix)

	// 上传文件
	_, err := u.client.Object.Put(ctx, objectKey, file, &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: getMimeType(filename),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to qcloud cos: %w", err)
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

func (u *QCloudUploader) Delete(ctx context.Context, key string) error {
	_, err := u.client.Object.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete object from qcloud cos: %w", err)
	}
	return nil
}

func (u *QCloudUploader) GetURL(ctx context.Context, key string) (string, error) {
	if u.config.Domain != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(u.config.Domain, "/"), key), nil
	}

	protocol := "https"
	if !u.config.UseSSL {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s/%s", protocol, u.config.Endpoint, key), nil
}
