package upload

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_ConfigBuilder(t *testing.T) {
	t.Run("WithLocal", func(t *testing.T) {
		cfg := NewConfigBuilder().
			WithLocal("/tmp/uploads", "http://localhost:8080/files").
			WithMaxFileSize(50).
			WithAllowedExtensions([]string{"jpg", "png"}).
			WithFilenameStrategy(FilenameStrategyUUID).
			Build()

		if cfg.Upload.Type != TypeLocal {
			t.Errorf("Expected type %s, got %s", TypeLocal, cfg.Upload.Type)
		}

		if cfg.Upload.Local.Path != "/tmp/uploads" {
			t.Errorf("Expected path /tmp/uploads, got %s", cfg.Upload.Local.Path)
		}

		if cfg.Upload.Local.URLPrefix != "http://localhost:8080/files" {
			t.Errorf("Expected URL prefix http://localhost:8080/files, got %s", cfg.Upload.Local.URLPrefix)
		}

		if cfg.UploadSettings.MaxFileSize != 50 {
			t.Errorf("Expected max file size 50, got %d", cfg.UploadSettings.MaxFileSize)
		}

		if len(cfg.UploadSettings.AllowedExtensions) != 2 {
			t.Errorf("Expected 2 allowed extensions, got %d", len(cfg.UploadSettings.AllowedExtensions))
		}

		if cfg.UploadSettings.FilenameStrategy != FilenameStrategyUUID {
			t.Errorf("Expected filename strategy %s, got %s", FilenameStrategyUUID, cfg.UploadSettings.FilenameStrategy)
		}
	})

	t.Run("WithAliyunOSS", func(t *testing.T) {
		cfg := NewConfigBuilder().
			WithAliyunOSS("oss-cn-hangzhou.aliyuncs.com", "test-key", "test-secret", "test-bucket").
			WithPathPrefix("uploads/").
			Build()

		if cfg.Upload.Type != TypeOSS {
			t.Errorf("Expected type %s, got %s", TypeOSS, cfg.Upload.Type)
		}

		if cfg.Upload.OSS.Provider != ProviderAliyun {
			t.Errorf("Expected provider %s, got %s", ProviderAliyun, cfg.Upload.OSS.Provider)
		}

		aliyun := cfg.Upload.OSS.Aliyun
		if aliyun.Endpoint != "oss-cn-hangzhou.aliyuncs.com" {
			t.Errorf("Expected endpoint oss-cn-hangzhou.aliyuncs.com, got %s", aliyun.Endpoint)
		}

		if aliyun.AccessKeyID != "test-key" {
			t.Errorf("Expected access key test-key, got %s", aliyun.AccessKeyID)
		}

		if aliyun.AccessKeySecret != "test-secret" {
			t.Errorf("Expected access secret test-secret, got %s", aliyun.AccessKeySecret)
		}

		if aliyun.Bucket != "test-bucket" {
			t.Errorf("Expected bucket test-bucket, got %s", aliyun.Bucket)
		}

		if aliyun.PathPrefix != "uploads/" {
			t.Errorf("Expected path prefix uploads/, got %s", aliyun.PathPrefix)
		}

		if !aliyun.UseSSL {
			t.Error("Expected UseSSL to be true by default")
		}
	})

	t.Run("WithTencentCOS", func(t *testing.T) {
		cfg := NewConfigBuilder().
			WithTencentCOS("ap-beijing", "secret-id", "secret-key", "bucket-appid").
			WithPathPrefix("files/").
			Build()

		if cfg.Upload.Type != TypeOSS {
			t.Errorf("Expected type %s, got %s", TypeOSS, cfg.Upload.Type)
		}

		if cfg.Upload.OSS.Provider != ProviderTencent {
			t.Errorf("Expected provider %s, got %s", ProviderTencent, cfg.Upload.OSS.Provider)
		}

		tencent := cfg.Upload.OSS.Tencent
		if tencent.Region != "ap-beijing" {
			t.Errorf("Expected region ap-beijing, got %s", tencent.Region)
		}

		if tencent.SecretID != "secret-id" {
			t.Errorf("Expected secret ID secret-id, got %s", tencent.SecretID)
		}

		if tencent.SecretKey != "secret-key" {
			t.Errorf("Expected secret key secret-key, got %s", tencent.SecretKey)
		}

		if tencent.Bucket != "bucket-appid" {
			t.Errorf("Expected bucket bucket-appid, got %s", tencent.Bucket)
		}

		if tencent.PathPrefix != "files/" {
			t.Errorf("Expected path prefix files/, got %s", tencent.PathPrefix)
		}
	})

	t.Run("WithMinIO", func(t *testing.T) {
		cfg := NewConfigBuilder().
			WithMinIO("localhost:9000", "minio", "minio123", "test-bucket").
			WithPathPrefix("objects/").
			Build()

		if cfg.Upload.Type != TypeMinIO {
			t.Errorf("Expected type %s, got %s", TypeMinIO, cfg.Upload.Type)
		}

		minio := cfg.Upload.MinIO
		if minio.Endpoint != "localhost:9000" {
			t.Errorf("Expected endpoint localhost:9000, got %s", minio.Endpoint)
		}

		if minio.AccessKey != "minio" {
			t.Errorf("Expected access key minio, got %s", minio.AccessKey)
		}

		if minio.SecretKey != "minio123" {
			t.Errorf("Expected secret key minio123, got %s", minio.SecretKey)
		}

		if minio.Bucket != "test-bucket" {
			t.Errorf("Expected bucket test-bucket, got %s", minio.Bucket)
		}

		if minio.PathPrefix != "objects/" {
			t.Errorf("Expected path prefix objects/, got %s", minio.PathPrefix)
		}

		if minio.UseSSL {
			t.Error("Expected UseSSL to be false by default for MinIO")
		}
	})

	t.Run("ChainedConfiguration", func(t *testing.T) {
		cfg := NewConfigBuilder().
			WithLocal("/uploads", "http://example.com/files").
			WithMaxFileSize(100).
			WithAllowedExtensions([]string{".jpg", ".png", ".gif"}).
			WithFilenameStrategy(FilenameStrategyTimestamp).
			Build()

		// 验证所有设置都正确应用
		if cfg.Upload.Type != TypeLocal {
			t.Error("Type should be local")
		}

		if cfg.Upload.Local.Path != "/uploads" {
			t.Error("Path not set correctly")
		}

		if cfg.UploadSettings.MaxFileSize != 100 {
			t.Error("Max file size not set correctly")
		}

		if len(cfg.UploadSettings.AllowedExtensions) != 3 {
			t.Error("Allowed extensions not set correctly")
		}

		if cfg.UploadSettings.FilenameStrategy != FilenameStrategyTimestamp {
			t.Error("Filename strategy not set correctly")
		}
	})

	t.Run("DefaultValues", func(t *testing.T) {
		cfg := NewConfigBuilder().Build()

		// 验证默认值
		if cfg.UploadSettings.MaxFileSize != 100 {
			t.Errorf("Expected default max file size 100, got %d", cfg.UploadSettings.MaxFileSize)
		}

		expectedExtensions := []string{".jpg", ".jpeg", ".png", ".pdf"}
		if len(cfg.UploadSettings.AllowedExtensions) != len(expectedExtensions) {
			t.Errorf("Expected %d default extensions, got %d", len(expectedExtensions), len(cfg.UploadSettings.AllowedExtensions))
		}

		if cfg.UploadSettings.FilenameStrategy != FilenameStrategyUUID {
			t.Errorf("Expected default filename strategy %s, got %s", FilenameStrategyUUID, cfg.UploadSettings.FilenameStrategy)
		}

		if cfg.UploadSettings.KeepOriginalName {
			t.Error("Expected KeepOriginalName to be false by default")
		}
	})

}

func Test_LoadConfig(t *testing.T) {
	t.Run("ValidConfigFile", func(t *testing.T) {
		configContent := `
upload:
  type: local
  local:
    path: /tmp/test-uploads
    url-prefix: http://localhost:8080/files

upload-settings:
  max-file-size: 10
  allowed-extensions:
    - .jpg
    - .png
    - .txt
  filename-strategy: uuid
  keep-original-name: false
`
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "test-config.yaml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		cfg, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if cfg.Upload.Type != "local" {
			t.Errorf("Expected type local, got %s", cfg.Upload.Type)
		}

		if cfg.Upload.Local.Path != "/tmp/test-uploads" {
			t.Errorf("Expected path /tmp/test-uploads, got %s", cfg.Upload.Local.Path)
		}

		if cfg.UploadSettings.MaxFileSize != 10 {
			t.Errorf("Expected max file size 10, got %d", cfg.UploadSettings.MaxFileSize)
		}

		if len(cfg.UploadSettings.AllowedExtensions) != 3 {
			t.Errorf("Expected 3 allowed extensions, got %d", len(cfg.UploadSettings.AllowedExtensions))
		}
	})
	t.Run("NonexistentFile", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/config.yaml")
		if err == nil {
			t.Error("Expected error for nonexistent config file")
		}
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		configContent := `
invalid: yaml: content
  - missing
    brackets
`

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid-config.yaml")

		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})

	t.Run("OSSConfig", func(t *testing.T) {
		configContent := `
upload:
  type: oss
  oss:
    provider: aliyun
    aliyun:
      endpoint: oss-cn-hangzhou.aliyuncs.com
      access-key-id: test-key
      access-key-secret: test-secret
      bucket: test-bucket
      path-prefix: uploads/
      use-ssl: true

upload-settings:
  max-file-size: 50
  allowed-extensions:
    - .jpg
    - .png
  filename-strategy: timestamp
`

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "oss-config.yaml")

		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		cfg, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if cfg.Upload.Type != "oss" {
			t.Errorf("Expected type oss, got %s", cfg.Upload.Type)
		}

		if cfg.Upload.OSS.Provider != "aliyun" {
			t.Errorf("Expected provider aliyun, got %s", cfg.Upload.OSS.Provider)
		}

		if cfg.Upload.OSS.Aliyun.Endpoint != "oss-cn-hangzhou.aliyuncs.com" {
			t.Errorf("Expected endpoint oss-cn-hangzhou.aliyuncs.com, got %s", cfg.Upload.OSS.Aliyun.Endpoint)
		}
	})
}
