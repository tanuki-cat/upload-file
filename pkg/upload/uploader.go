package upload

import (
	"context"
	"mime/multipart"
	"upload-util/internal/config"
	"upload-util/internal/service"
)

type UploadResult struct {
	URL      string `json:"url"`
	Key      string `json:"key"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

type Uploader interface {
	Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*UploadResult, error)
	Delete(ctx context.Context, key string) error
	GetURL(ctx context.Context, key string) (string, error)
}

type uploaderWrapper struct {
	internal service.Uploader
}

func (w *uploaderWrapper) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	result, err := w.internal.Upload(ctx, file, header)
	if err != nil {
		return nil, err
	}
	return &UploadResult{
		URL:      result.URL,
		Key:      result.Key,
		Size:     result.Size,
		MimeType: result.MimeType,
	}, nil
}

func (w *uploaderWrapper) Delete(ctx context.Context, key string) error {
	return w.internal.Delete(ctx, key)
}

func (w *uploaderWrapper) GetURL(ctx context.Context, key string) (string, error) {
	return w.internal.GetURL(ctx, key)
}

func NewUploader(cfg *Config) (Uploader, error) {
	if cfg == nil {
		return nil, nil
	}
	factory := service.NewUploadFactory(cfg)
	internalUploader, err := factory.CreateUploader()
	if err != nil {
		return nil, err
	}
	// 返回包装后的上传器
	return &uploaderWrapper{internal: internalUploader}, nil
}

func LoadConfig(configPath string) (*Config, error) {
	return config.LoadConfig(configPath)
}
