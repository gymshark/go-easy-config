package generic

import (
	"os"
	"testing"
)

type testYAMLConfig struct {
	Field1 string `yaml:"Field1"`
	Field2 string `yaml:"Field2"`
	Field3 string `yaml:"Field3"`
}

func writeTestYAMLFile(path string, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func TestYAMLLoader_Load_Success(t *testing.T) {
	path := "test_config.yaml"
	yamlContent := "Field1: value1\nField2: value2\nField3: value3\n"
	if err := writeTestYAMLFile(path, yamlContent); err != nil {
		t.Fatalf("failed to write yaml file: %v", err)
	}
	defer os.Remove(path)

	cfg := &testYAMLConfig{}
	loader := YAMLLoader[testYAMLConfig]{Source: path}
	if err := loader.Load(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Field1 != "value1" || cfg.Field2 != "value2" || cfg.Field3 != "value3" {
		t.Errorf("unexpected config values: %+v", cfg)
	}
}

func TestYAMLLoader_Load_FileNotFound(t *testing.T) {
	loader := YAMLLoader[testYAMLConfig]{Source: "nonexistent.yaml"}
	cfg := &testYAMLConfig{}
	if err := loader.Load(cfg); err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestYAMLLoader_Load_InvalidFormat(t *testing.T) {
	path := "invalid_config.yaml"
	yamlContent := "not a yaml file"
	if err := writeTestYAMLFile(path, yamlContent); err != nil {
		t.Fatalf("failed to write yaml file: %v", err)
	}
	defer os.Remove(path)

	cfg := &testYAMLConfig{}
	loader := YAMLLoader[testYAMLConfig]{Source: path}
	if err := loader.Load(cfg); err == nil {
		t.Error("expected error for invalid yaml format, got nil")
	}
}

func TestYAMLLoader_Load_BytesSource(t *testing.T) {
	yamlContent := []byte("Field1: value1\nField2: value2\nField3: value3\n")
	cfg := &testYAMLConfig{}
	loader := YAMLLoader[testYAMLConfig]{Source: yamlContent}
	if err := loader.Load(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Field1 != "value1" || cfg.Field2 != "value2" || cfg.Field3 != "value3" {
		t.Errorf("unexpected config values: %+v", cfg)
	}
}
