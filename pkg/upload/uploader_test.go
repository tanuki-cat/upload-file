package upload

import (
	"testing"
)

func Test_NewUploader(t *testing.T) {
	t.Run("LocalUploader", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := NewConfigBuilder().
			WithLocal(tmpDir, "http://localhost:8080/files").
			Build()
		uploader, err := NewUploader(cfg)
		if err != nil {
			t.Fatalf("NewUploader failed: %v", err)
		}

		if uploader == nil {
			t.Fatal("Expected non-nil uploader")
		}
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		cfg := NewConfigBuilder().Build()
		cfg.Upload.Type = "invalid"

		_, err := NewUploader(cfg)
		if err == nil {
			t.Error("Expected error for invalid config")
		}
	})

	t.Run("NilConfig", func(t *testing.T) {
		_, err := NewUploader(nil)
		if err != nil {
			t.Error("Expected error for nil config")
		}
	})

}
