package generic

import (
	"os"
	"testing"
)

type testJSONConfig struct {
	Field1 string `json:"Field1"`
	Field2 string `json:"Field2"`
	Field3 string `json:"Field3"`
}

func writeTestJSONFile(path string, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func TestJSONLoader_Load_Success(t *testing.T) {
	path := "test_config.json"
	jsonContent := `{"Field1":"value1","Field2":"value2","Field3":"value3"}`
	if err := writeTestJSONFile(path, jsonContent); err != nil {
		t.Fatalf("failed to write json file: %v", err)
	}
	defer os.Remove(path)

	cfg := &testJSONConfig{}
	loader := JSONLoader[testJSONConfig]{Source: path}
	if err := loader.Load(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Field1 != "value1" || cfg.Field2 != "value2" || cfg.Field3 != "value3" {
		t.Errorf("unexpected config values: %+v", cfg)
	}
}

func TestJSONLoader_Load_FileNotFound(t *testing.T) {
	loader := JSONLoader[testJSONConfig]{Source: "nonexistent.json"}
	cfg := &testJSONConfig{}
	if err := loader.Load(cfg); err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestJSONLoader_Load_InvalidFormat(t *testing.T) {
	path := "invalid_config.json"
	jsonContent := "not a json file"
	if err := writeTestJSONFile(path, jsonContent); err != nil {
		t.Fatalf("failed to write json file: %v", err)
	}
	defer os.Remove(path)

	cfg := &testJSONConfig{}
	loader := JSONLoader[testJSONConfig]{Source: path}
	if err := loader.Load(cfg); err == nil {
		t.Error("expected error for invalid json format, got nil")
	}
}

func TestJSONLoader_Load_BytesSource(t *testing.T) {
	jsonContent := []byte(`{"Field1":"value1","Field2":"value2","Field3":"value3"}`)
	cfg := &testJSONConfig{}
	loader := JSONLoader[testJSONConfig]{Source: jsonContent}
	if err := loader.Load(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Field1 != "value1" || cfg.Field2 != "value2" || cfg.Field3 != "value3" {
		t.Errorf("unexpected config values: %+v", cfg)
	}
}
