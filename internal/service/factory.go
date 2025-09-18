package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"upload-util/internal/config"
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
type UploadFactory struct {
	config *config.UploadConfig
}

func NewUploadFactory(config *config.UploadConfig) *UploadFactory {
	return &UploadFactory{
		config: config,
	}
}

func (f *UploadFactory) CreateUploader() (Uploader, error) {
	switch f.config.Upload.Type {
	case "local":
		return NewLocalUploader(f.config.Upload.Local, &f.config.UploadSettings)
	case "oss":
		return f.createOSSUploader() // 修正函数名
	case "minio":
		return NewMinIOUploader(f.config.Upload.MinIO, &f.config.UploadSettings)
	default:
		return nil, fmt.Errorf("unsupported upload type: %s", f.config.Upload.Type)
	}
}

func (f *UploadFactory) createOSSUploader() (Uploader, error) {
	ossConfig := f.config.Upload.OSS
	if ossConfig == nil {
		return nil, fmt.Errorf("oss config is nil")
	}

	switch ossConfig.Provider {
	case "aliyun":
		return NewAliyunOSSUploader(ossConfig.Aliyun, &f.config.UploadSettings)
	case "tencent":
		return NewTencentCOSUploader(ossConfig.Tencent, &f.config.UploadSettings)
	case "huawei":
		return NewHuaweiOBSUploader(ossConfig.Huawei, &f.config.UploadSettings)
	case "aws":
		return NewAWSS3Uploader(ossConfig.AWS, &f.config.UploadSettings)
	case "qcloud":
		return NewQCloudCOSUploader(ossConfig.QCloud, &f.config.UploadSettings)
		//return nil, fmt.Errorf("qcloud oss is not supported yet")
	default:
		return nil, fmt.Errorf("unsupported oss provider: %s", ossConfig.Provider)
	}
}
