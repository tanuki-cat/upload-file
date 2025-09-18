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

type TencentUpload struct {
	client   *cos.Client
	config   *config.TencentCOSConfig
	settings *config.UploadSettings
}

func NewTencentCOSUploader(config *config.TencentCOSConfig, settings *config.UploadSettings) (*TencentUpload, error) {
	if config == nil {
		return nil, fmt.Errorf("tencent oss config is nil")
	}

	// 构建 bucket URL
	bucketURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Bucket, config.Region)
	if !config.UseSSL {
		bucketURL = fmt.Sprintf("http://%s.cos.%s.myqcloud.com", config.Bucket, config.Region)
	}

	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bucket url: %w", err)
	}

	// 创建 COS 客户端
	client := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.SecretID,
			SecretKey: config.SecretKey,
		},
	})

	return &TencentUpload{
		client:   client,
		config:   config,
		settings: settings,
	}, nil

}

func (u *TencentUpload) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
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
		return nil, fmt.Errorf("failed to upload file to tencent cos: %w", err)
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

func (u *TencentUpload) Delete(ctx context.Context, key string) error {
	_, err := u.client.Object.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete object from tencent cos: %w", err)
	}
	return nil
}

func (u *TencentUpload) GetURL(ctx context.Context, key string) (string, error) {
	if u.config.Domain != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(u.config.Domain, "/"), key), nil
	}

	protocol := "https"
	if !u.config.UseSSL {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s.cos.%s.myqcloud.com/%s", protocol, u.config.Bucket, u.config.Region, key), nil
}
