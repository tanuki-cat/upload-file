package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type UploadConfig struct {
	Upload         UploadProvider `yaml:"upload"`
	UploadSettings UploadSettings `yaml:"upload-settings"`
}

type UploadProvider struct {
	Type  string       `yaml:"type"`
	Local *LocalConfig `yaml:"local,omitempty"`
	OSS   *OSSConfig   `yaml:"oss,omitempty"`
	MinIO *MinioConfig `yaml:"minio,omitempty"`
}

type LocalConfig struct {
	Path      string `yaml:"path"`
	URLPrefix string `yaml:"url-prefix,omitempty"`
}

type OSSConfig struct {
	Provider string            `yaml:"provider"`
	Aliyun   *AliyunOSSConfig  `yaml:"aliyun,omitempty"`
	Tencent  *TencentCOSConfig `yaml:"tencent,omitempty"`
	Huawei   *HuaweiOBSConfig  `yaml:"huawei,omitempty"`
	AWS      *AWSS3Config      `yaml:"aws,omitempty"`
	QCloud   *QCloudCOSConfig  `yaml:"qcloud,omitempty"`
}

type AliyunOSSConfig struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access-key-id"`
	AccessKeySecret string `yaml:"access-key-secret"`
	Bucket          string `yaml:"bucket"`
	Domain          string `yaml:"domain,omitempty"`
	PathPrefix      string `yaml:"path-prefix,omitempty"`
	UseSSL          bool   `yaml:"use-ssl"`
	UseInternal     bool   `yaml:"use-internal"`
	SignURLExpire   int64  `yaml:"sign-url-expire,omitempty"`
}

type TencentCOSConfig struct {
	Region     string `yaml:"region"`
	SecretID   string `yaml:"secret-id"`
	SecretKey  string `yaml:"secret-key"`
	Bucket     string `yaml:"bucket"`
	Domain     string `yaml:"domain,omitempty"`
	PathPrefix string `yaml:"path-prefix,omitempty"`
	UseSSL     bool   `yaml:"use-ssl"`
}

type HuaweiOBSConfig struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access-key-id"`
	SecretAccessKey string `yaml:"secret-access-key"`
	Bucket          string `yaml:"bucket"`
	Region          string `yaml:"region,omitempty"`
	Domain          string `yaml:"domain,omitempty"`
	PathPrefix      string `yaml:"path-prefix,omitempty"`
	UseSSL          bool   `yaml:"use-ssl"`
}

// AWSS3Config AWS S3 配置
type AWSS3Config struct {
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access-key-id"`
	SecretAccessKey string `yaml:"secret-access-key"`
	Bucket          string `yaml:"bucket"`
	Endpoint        string `yaml:"endpoint,omitempty"`
	Domain          string `yaml:"domain,omitempty"`
	PathPrefix      string `yaml:"path-prefix,omitempty"`
	UseSSL          bool   `yaml:"use-ssl"`
}

// QCloudCOSConfig 其他兼容 S3 的云服务配置
type QCloudCOSConfig struct {
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access-key-id"`
	SecretAccessKey string `yaml:"secret-access-key"`
	Bucket          string `yaml:"bucket"`
	Endpoint        string `yaml:"endpoint"`
	Domain          string `yaml:"domain,omitempty"`
	PathPrefix      string `yaml:"path-prefix,omitempty"`
	UseSSL          bool   `yaml:"use-ssl"`
}

type MinioConfig struct {
	Endpoint   string `yaml:"endpoint"`
	AccessKey  string `yaml:"access-key"`
	SecretKey  string `yaml:"secret-key"`
	Bucket     string `yaml:"bucket"`
	Domain     string `yaml:"domain,omitempty"`
	PathPrefix string `yaml:"path-prefix,omitempty"`
	UseSSL     bool   `yaml:"use-ssl"`
	Region     string `yaml:"region,omitempty"`
}

type UploadSettings struct {
	MaxFileSize       int64    `yaml:"max-file-size"`
	AllowedExtensions []string `yaml:"allowed-extensions"`
	FilenameStrategy  string   `yaml:"filename-strategy"`
	KeepOriginalName  bool     `yaml:"keep-original-name"`
}

func LoadConfig(configPath string) (*UploadConfig, error) {
	fileData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var config UploadConfig
	if err := yaml.Unmarshal(fileData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return &config, nil
}

func (c *UploadConfig) Validate() error {
	switch c.Upload.Type {
	case "local":
		if c.Upload.Local == nil {
			return fmt.Errorf("local config is required when type is local")
		}
		if c.Upload.Local.Path == "" {
			return fmt.Errorf("local path is required when type is local")
		}
	case "oss":
		if c.Upload.OSS == nil {
			return fmt.Errorf("oss config is required when type is oss")
		}
		return c.validateOSS()

	}
	return nil
}

func (c *UploadConfig) validateOSS() error {
	oss := c.Upload.OSS
	switch oss.Provider {
	case "aliyun":
		if oss.Aliyun == nil {
			return fmt.Errorf("aliyun oss config is required when provider is aliyun")
		}
		return c.validateAliyunOSS()
	case "tencent":
		if oss.Tencent == nil {
			return fmt.Errorf("tencent oss config is required when provider is tencent")
		}
		return c.validateTencentCOS()
	case "huawei":
		if oss.Huawei == nil {
			return fmt.Errorf("huawei oss config is required when provider is huawei")
		}
		return c.validateHuaweiOSS()
	case "aws":
		if oss.AWS == nil {
			return fmt.Errorf("aws oss config is required when provider is aws")
		}
		return c.validateAWSS3()
	default:
		return fmt.Errorf("unsupported oss provider: %s", oss.Provider)
	}
}

func (c *UploadConfig) validateAliyunOSS() error {
	oss := c.Upload.OSS.Aliyun
	if oss.Endpoint == "" {
		return fmt.Errorf("aliyun oss endpoint is required")
	}
	if oss.AccessKeyID == "" {
		return fmt.Errorf("aliyun oss access key id is required")
	}
	if oss.AccessKeySecret == "" {
		return fmt.Errorf("aliyun oss access key secret is required")
	}
	if oss.Bucket == "" {
		return fmt.Errorf("aliyun oss bucket is required")
	}
	return nil
}

func (c *UploadConfig) validateTencentCOS() error {
	oss := c.Upload.OSS.Tencent
	if oss.Region == "" {
		return fmt.Errorf("tencent oss region is required")
	}
	if oss.SecretID == "" {
		return fmt.Errorf("tencent oss secret id is required")
	}
	if oss.SecretKey == "" {
		return fmt.Errorf("tencent oss secret key is required")
	}
	if oss.Bucket == "" {
		return fmt.Errorf("tencent oss bucket is required")
	}
	return nil
}

func (c *UploadConfig) validateHuaweiOSS() error {
	oss := c.Upload.OSS.Huawei
	if oss.Endpoint == "" {
		fmt.Errorf("huawei oss endpoint is required")
	}
	if oss.AccessKeyID == "" {
		fmt.Errorf("huawei oss access key id is required")
	}
	if oss.SecretAccessKey == "" {
		fmt.Errorf("huawei oss access key secret is required")
	}
	if oss.Bucket == "" {
		fmt.Errorf("huawei oss bucket is required")
	}
	return nil
}

func (c *UploadConfig) validateAWSS3() error {
	oss := c.Upload.OSS.AWS
	if oss.Region == "" {
		return fmt.Errorf("aws oss region is required")
	}
	if oss.AccessKeyID == "" {
		return fmt.Errorf("aws oss access key id is required")
	}
	if oss.SecretAccessKey == "" {
		return fmt.Errorf("aws oss access key secret is required")
	}
	if oss.Bucket == "" {
		return fmt.Errorf("aws oss bucket is required")
	}

	return nil
}

func (c *UploadConfig) validateQCloudCOS() error {
	oss := c.Upload.OSS.QCloud
	if oss.Region == "" {
		return fmt.Errorf("qcloud oss region is required")
	}
	if oss.AccessKeyID == "" {
		return fmt.Errorf("qcloud oss access key id is required")
	}
	if oss.SecretAccessKey == "" {
		return fmt.Errorf("qcloud oss access key secret is required")
	}
	if oss.Bucket == "" {
		return fmt.Errorf("qcloud oss bucket is required")
	}
	if oss.Endpoint == "" {
		return fmt.Errorf("qcloud oss endpoint is required")
	}
	return nil
}

func (c *UploadConfig) validateMinIO() error {
	minio := c.Upload.MinIO
	if minio.Endpoint == "" {
		return fmt.Errorf("minio endpoint is required")
	}
	if minio.AccessKey == "" {
		return fmt.Errorf("minio access key is required")
	}
	if minio.SecretKey == "" {
		return fmt.Errorf("minio secret key is required")
	}
	if minio.Bucket == "" {
		return fmt.Errorf("minio bucket is required")
	}
	return nil
}

func (c *UploadConfig) GetCurrentOSSProvider() (any, string, error) {
	if c.Upload.Type != "oss" || c.Upload.OSS == nil {
		return nil, "", fmt.Errorf("not using oss upload")
	}
	switch c.Upload.OSS.Provider {
	case "aliyun":
		return c.Upload.OSS.Aliyun, "aliyun", nil
	case "tencent":
		return c.Upload.OSS.Tencent, "tencent", nil
	case "huawei":
		return c.Upload.OSS.Huawei, "huawei", nil
	case "aws":
		return c.Upload.OSS.AWS, "aws", nil
	case "qcloud":
		return c.Upload.OSS.QCloud, "qcloud", nil
	default:
		return nil, "", fmt.Errorf("unsupported oss provider: %s", c.Upload.OSS.Provider)
	}
}
