package generic

import (
	"os"
	"testing"
)

type testIniConfig struct {
	Field1 string `ini:"Field1"`
	Field2 string `ini:"Field2"`
	Field3 string `ini:"Field3"`
}

func writeTestIniFile(path string, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func TestIniLoader_Load_Success(t *testing.T) {
	path := "test_config.ini"
	iniContent := "[DEFAULT]\nField1 = value1\nField2 = value2\nField3 = value3\n"
	if err := writeTestIniFile(path, iniContent); err != nil {
		t.Fatalf("failed to write ini file: %v", err)
	}
	defer os.Remove(path)

	cfg := &testIniConfig{}
	loader := IniLoader[testIniConfig]{Source: path}
	if err := loader.Load(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Field1 != "value1" || cfg.Field2 != "value2" || cfg.Field3 != "value3" {
		t.Errorf("unexpected config values: %+v", cfg)
	}
}

func TestIniLoader_Load_FileNotFound(t *testing.T) {
	loader := IniLoader[testIniConfig]{Source: "nonexistent.ini"}
	cfg := &testIniConfig{}
	if err := loader.Load(cfg); err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestIniLoader_Load_InvalidFormat(t *testing.T) {
	path := "invalid_config.ini"
	iniContent := "not an ini file"
	if err := writeTestIniFile(path, iniContent); err != nil {
		t.Fatalf("failed to write ini file: %v", err)
	}
	defer os.Remove(path)

	cfg := &testIniConfig{}
	loader := IniLoader[testIniConfig]{Source: path}
	if err := loader.Load(cfg); err == nil {
		t.Error("expected error for invalid ini format, got nil")
	}
}

func TestIniLoader_Load_BytesSource(t *testing.T) {
	iniContent := []byte("[DEFAULT]\nField1 = value1\nField2 = value2\nField3 = value3\n")
	cfg := &testIniConfig{}
	loader := IniLoader[testIniConfig]{Source: iniContent}
	if err := loader.Load(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Field1 != "value1" || cfg.Field2 != "value2" || cfg.Field3 != "value3" {
		t.Errorf("unexpected config values: %+v", cfg)
	}
}
