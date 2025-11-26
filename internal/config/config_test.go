package config

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("error in load config dir")
	}
	projectPatth := filepath.Join(filepath.Dir(filename), "../..")
	configPath := filepath.Join(projectPatth, "config-example.yaml")
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("error loading config: %v", err)
	}
	t.Log(config)
	err = config.Validate()
	if err != nil {
		t.Fatalf("error validating config: %v", err)
	}
}
