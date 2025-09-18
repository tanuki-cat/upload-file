package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"upload-util/internal/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type AWSS3Uploader struct {
	client   *s3.S3
	config   *config.AWSS3Config
	settings *config.UploadSettings
}

func NewAWSS3Uploader(config *config.AWSS3Config, settings *config.UploadSettings) (*AWSS3Uploader, error) {
	if config == nil {
		return nil, fmt.Errorf("aws s3 config is nil")
	}

	// 创建 AWS 配置
	awsConfig := &aws.Config{
		Region:      aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
	}

	// 如果指定了自定义 endpoint
	if config.Endpoint != "" {
		awsConfig.Endpoint = aws.String(fmt.Sprintf("https://%s", config.Endpoint))
		if !config.UseSSL {
			awsConfig.Endpoint = aws.String(fmt.Sprintf("http://%s", config.Endpoint))
		}
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	// 创建会话
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create aws session: %w", err)
	}

	// 创建 S3 客户端
	client := s3.New(sess)

	return &AWSS3Uploader{
		client:   client,
		config:   config,
		settings: settings,
	}, nil
}

func (u *AWSS3Uploader) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	// 验证文件
	if err := validateFile(header, u.settings); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// 生成文件名和对象键
	filename := generateFileName(header, u.settings)
	objectKey := buildObjectKey(filename, u.config.PathPrefix)

	// 上传文件
	_, err := u.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.config.Bucket),
		Key:         aws.String(objectKey),
		Body:        file,
		ContentType: aws.String(getMimeType(filename)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to aws s3: %w", err)
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

func (u *AWSS3Uploader) Delete(ctx context.Context, key string) error {
	_, err := u.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(u.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from aws s3: %w", err)
	}
	return nil
}

func (u *AWSS3Uploader) GetURL(ctx context.Context, key string) (string, error) {
	if u.config.Domain != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(u.config.Domain, "/"), key), nil
	}

	protocol := "https"
	if !u.config.UseSSL {
		protocol = "http"
	}

	if u.config.Endpoint != "" {
		return fmt.Sprintf("%s://%s/%s/%s", protocol, u.config.Endpoint, u.config.Bucket, key), nil
	}

	return fmt.Sprintf("%s://%s.s3.%s.amazonaws.com/%s", protocol, u.config.Bucket, u.config.Region, key), nil
}
