package upload

import (
	"testing"
)

func TestConstants(t *testing.T) {
	t.Run("FilenameStrategy", func(t *testing.T) {
		exceptedStrategies := map[string]string{
			"FilenameStrategyOriginal":  "original",
			"FilenameStrategyUUID":      "uuid",
			"FilenameStrategyTimestamp": "timestamp",
		}

		actualStrategies := map[string]string{
			"FilenameStrategyOriginal":  FilenameStrategyOriginal,
			"FilenameStrategyUUID":      FilenameStrategyUUID,
			"FilenameStrategyTimestamp": FilenameStrategyTimestamp,
		}
		for name, excepted := range exceptedStrategies {
			if actual, exists := actualStrategies[name]; !exists {
				t.Errorf("Constant %s not defined", name)
			} else if actual != excepted {
				t.Errorf("Constant %s: expected %s, got %s", name, excepted, actual)
			}
		}
	})

	t.Run("UploadTypes", func(t *testing.T) {
		exceptedTypes := map[string]string{
			"TypeLocal": "local",
			"TypeOSS":   "oss",
			"TypeMinIO": "minio",
		}

		actualTypes := map[string]string{
			"TypeLocal": TypeLocal,
			"TypeOSS":   TypeOSS,
			"TypeMinIO": TypeMinIO,
		}
		for name, excepted := range exceptedTypes {
			if actual, exists := actualTypes[name]; !exists {
				t.Errorf("Constant %s not defined", name)
			} else if actual != excepted {
				t.Errorf("Constant %s: expected %s, got %s", name, excepted, actual)
			}
		}
	})

	t.Run("OSSProviders", func(t *testing.T) {
		exceptedProviders := map[string]string{
			"ProviderAliyun":  "aliyun",
			"ProviderQCloud":  "qcloud",
			"ProviderHuawei":  "huawei",
			"ProviderAWS":     "aws",
			"ProviderTencent": "tencent",
		}
		actualProviders := map[string]string{
			"ProviderAliyun":  ProviderAliyun,
			"ProviderQCloud":  ProviderQCloud,
			"ProviderHuawei":  ProviderHuawei,
			"ProviderAWS":     ProviderAWS,
			"ProviderTencent": ProviderTencent,
		}

		for name, excepted := range exceptedProviders {
			if actual, exists := actualProviders[name]; !exists {
				t.Errorf("Constant %s not defined", name)
			} else if actual != excepted {
				t.Errorf("Constant %s: expected %s, got %s", name, excepted, actual)
			}
		}
	})

}

func Test_ConstantsUsage(t *testing.T) {
	t.Run("FilenameStrategyConstants", func(t *testing.T) {
		strategies := []string{
			FilenameStrategyOriginal,
			FilenameStrategyUUID,
			FilenameStrategyTimestamp,
		}
		for _, strategy := range strategies {
			cfg := NewConfigBuilder().
				WithLocal("/tmp", "http://localhost").
				WithFilenameStrategy(strategy).
				Build()
			if cfg.UploadSettings.FilenameStrategy != strategy {
				t.Errorf("Strategy %s not set correctly", strategy)
			}
		}
	})

	t.Run("TypeConstants", func(t *testing.T) {
		// TypeLocal
		cfg1 := NewConfigBuilder().WithLocal("/tmp", "http://localhost").Build()
		if cfg1.Upload.Type != TypeLocal {
			t.Errorf("Expected type %s for local config", TypeLocal)
		}

		// TypeOSS
		cfg2 := NewConfigBuilder().
			WithAliyunOSS("endpoint", "key", "secret", "bucket").
			Build()
		if cfg2.Upload.Type != TypeOSS {
			t.Errorf("Expected type %s for OSS config", TypeOSS)
		}

		// TypeMinIO
		cfg3 := NewConfigBuilder().
			WithMinIO("localhost:9000", "key", "secret", "bucket").
			Build()
		if cfg3.Upload.Type != TypeMinIO {
			t.Errorf("Expected type %s for MinIO config", TypeMinIO)
		}
	})

	t.Run("ProviderConstants", func(t *testing.T) {
		// Aliyun
		cfg1 := NewConfigBuilder().
			WithAliyunOSS("endpoint", "key", "secret", "bucket").
			Build()
		if cfg1.Upload.OSS.Provider != ProviderAliyun {
			t.Errorf("Expected provider %s for Aliyun config", ProviderAliyun)
		}

		// Tencent
		cfg2 := NewConfigBuilder().
			WithTencentCOS("region", "id", "key", "bucket").
			Build()
		if cfg2.Upload.OSS.Provider != ProviderTencent {
			t.Errorf("Expected provider %s for Tencent config", ProviderTencent)
		}
	})
}
