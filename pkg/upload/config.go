package upload

import (
	"upload-util/internal/config"
)

// Config 配置类型别名
type Config = config.UploadConfig

// ConfigBuilder 配置构建器
type ConfigBuilder struct {
	cfg *config.UploadConfig
}

// NewConfigBuilder 创建配置构建器
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		cfg: &config.UploadConfig{
			UploadSettings: config.UploadSettings{
				MaxFileSize:       100,
				AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".pdf"},
				FilenameStrategy:  "uuid",
				KeepOriginalName:  false,
			},
		},
	}
}

// WithLocal 配置本地存储
func (b *ConfigBuilder) WithLocal(path, urlPrefix string) *ConfigBuilder {
	b.cfg.Upload.Type = "local"
	b.cfg.Upload.Local = &config.LocalConfig{
		Path:      path,
		URLPrefix: urlPrefix,
	}
	return b
}

// WithAliyunOSS 配置阿里云 OSS
func (b *ConfigBuilder) WithAliyunOSS(endpoint, accessKeyID, accessKeySecret, bucket string) *ConfigBuilder {
	b.cfg.Upload.Type = "oss"
	b.cfg.Upload.OSS = &config.OSSConfig{
		Provider: "aliyun",
		Aliyun: &config.AliyunOSSConfig{
			Endpoint:        endpoint,
			AccessKeyID:     accessKeyID,
			AccessKeySecret: accessKeySecret,
			Bucket:          bucket,
			UseSSL:          true,
		},
	}
	return b
}

// WithTencentCOS 配置腾讯云 COS
func (b *ConfigBuilder) WithTencentCOS(region, secretID, secretKey, bucket string) *ConfigBuilder {
	b.cfg.Upload.Type = "oss"
	b.cfg.Upload.OSS = &config.OSSConfig{
		Provider: "tencent",
		Tencent: &config.TencentCOSConfig{
			Region:    region,
			SecretID:  secretID,
			SecretKey: secretKey,
			Bucket:    bucket,
			UseSSL:    true,
		},
	}
	return b
}

// WithMinIO 配置 MinIO
func (b *ConfigBuilder) WithMinIO(endpoint, accessKey, secretKey, bucket string) *ConfigBuilder {
	b.cfg.Upload.Type = "minio"
	b.cfg.Upload.MinIO = &config.MinioConfig{
		Endpoint:  endpoint,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Bucket:    bucket,
		UseSSL:    false,
	}
	return b
}

// WithPathPrefix 设置路径前缀
func (b *ConfigBuilder) WithPathPrefix(prefix string) *ConfigBuilder {
	if b.cfg.Upload.OSS != nil {
		if b.cfg.Upload.OSS.Aliyun != nil {
			b.cfg.Upload.OSS.Aliyun.PathPrefix = prefix
		}
		if b.cfg.Upload.OSS.Tencent != nil {
			b.cfg.Upload.OSS.Tencent.PathPrefix = prefix
		}
		if b.cfg.Upload.OSS.Huawei != nil {
			b.cfg.Upload.OSS.Huawei.PathPrefix = prefix
		}
	}
	if b.cfg.Upload.MinIO != nil {
		b.cfg.Upload.MinIO.PathPrefix = prefix
	}
	return b
}

// WithMaxFileSize 设置最大文件大小 (MB)
func (b *ConfigBuilder) WithMaxFileSize(size int64) *ConfigBuilder {
	b.cfg.UploadSettings.MaxFileSize = size
	return b
}

// WithAllowedExtensions 设置允许的文件扩展名
func (b *ConfigBuilder) WithAllowedExtensions(extensions []string) *ConfigBuilder {
	b.cfg.UploadSettings.AllowedExtensions = extensions
	return b
}

// WithFilenameStrategy 设置文件命名策略
func (b *ConfigBuilder) WithFilenameStrategy(strategy string) *ConfigBuilder {
	b.cfg.UploadSettings.FilenameStrategy = strategy
	return b
}

// Build 构建配置
func (b *ConfigBuilder) Build() *config.UploadConfig {
	return b.cfg
}
