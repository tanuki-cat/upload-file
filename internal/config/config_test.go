package config

import "testing"

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("~/GoLandProjects/xinan-company/upload-util/config-local-private.yaml")
	if err != nil {
		t.Fatalf("error loading config: %v", err)
	}
	t.Log(config)
	err = config.Validate()
	if err != nil {
		t.Fatalf("error validating config: %v", err)
	}
}
