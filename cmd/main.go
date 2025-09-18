package main

import (
	"log"

	"upload-util/internal/config"
	"upload-util/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("internal/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Current upload type: %s", cfg.Upload.Type)

	if cfg.Upload.Type == "oss" {
		log.Printf("Current OSS provider: %s", cfg.Upload.OSS.Provider)

		// 演示如何动态切换 OSS 厂商
		switchOSSProvider(cfg, "tencent")
		switchOSSProvider(cfg, "aliyun")
		switchOSSProvider(cfg, "aws")
	}
}

func switchOSSProvider(cfg *config.UploadConfig, newProvider string) {
	log.Printf("\n=== Switching OSS provider to: %s ===", newProvider)

	// 保存原始配置
	originalProvider := cfg.Upload.OSS.Provider

	// 切换厂商
	cfg.Upload.OSS.Provider = newProvider

	// 验证新配置
	if err := cfg.Validate(); err != nil {
		log.Printf("Configuration validation failed for %s: %v", newProvider, err)
		// 回滚到原始配置
		cfg.Upload.OSS.Provider = originalProvider
		return
	}

	// 创建新的上传器
	factory := service.NewUploadFactory(cfg)
	uploader, err := factory.CreateUploader()
	if err != nil {
		log.Printf("Failed to create uploader for %s: %v", newProvider, err)
		return
	}

	log.Printf("Successfully switched to %s provider", newProvider)
	log.Printf("Uploader type: %T", uploader)

	// 获取当前厂商配置信息
	providerConfig, providerName, err := cfg.GetCurrentOSSProvider()
	if err != nil {
		log.Printf("Failed to get current provider config: %v", err)
		return
	}

	log.Printf("Provider name: %s", providerName)
	log.Printf("Provider config: %+v", providerConfig)
}
