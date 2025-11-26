package service

import (
	"os"
	"testing"
)

func TestFilePath(t *testing.T) {
	t.Run("test getwd", func(t *testing.T) {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("error getting current working directory: %v", err)
		}
		t.Logf("current working directory: %s", wd)
	})
}

func TestFilePathInit(t *testing.T) {
	t.Run("test file path init", func(t *testing.T) {
		dir, err := os.Getwd()
		if err != nil {
			t.Fatalf("error getting current work dir: %v", err)
		}
		t.Logf("current working dir : %s", dir)
	})
}
