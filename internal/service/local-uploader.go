package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"upload-util/internal/config"
)

type LocalUploader struct {
	config   *config.LocalConfig
	settings *config.UploadSettings
}

func NewLocalUploader(config *config.LocalConfig, settings *config.UploadSettings) (*LocalUploader, error) {
	if config == nil {
		return nil, fmt.Errorf("local config is required")
	}
	uploadPath := config.Path
	if strings.HasPrefix(uploadPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home dir: %w", err)
		}
		uploadPath = filepath.Join(homeDir, uploadPath[2:])
	}
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload path: %w", err)
	}
	return &LocalUploader{
		config:   config,
		settings: settings,
	}, nil
}

func (u *LocalUploader) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	if err := validateFile(header, u.settings); err != nil {
		return nil, fmt.Errorf("failed to validate file: %w", err)
	}
	filename := generateFileName(header, u.settings)
	uploadPath := u.config.Path
	if strings.HasPrefix(uploadPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		uploadPath = filepath.Join(homeDir, uploadPath[2:])
	}

	filePath := filepath.Join(uploadPath, filename)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {
			return
		}
	}(dst)

	// 复制文件内容
	size, err := io.Copy(dst, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// 生成访问 URL
	url, _ := u.GetURL(ctx, filename)

	return &UploadResult{
		URL:      url,
		Key:      filename,
		Size:     size,
		MimeType: getMimeType(filename),
	}, nil
}

func (u *LocalUploader) Delete(ctx context.Context, key string) error {
	// 展开用户目录
	uploadPath := u.config.Path
	if strings.HasPrefix(uploadPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		uploadPath = filepath.Join(homeDir, uploadPath[2:])
	}

	filePath := filepath.Join(uploadPath, key)
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (u *LocalUploader) GetURL(ctx context.Context, key string) (string, error) {
	if u.config.URLPrefix != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(u.config.URLPrefix, "/"), key), nil
	}
	return fmt.Sprintf("file://%s", key), nil
}
