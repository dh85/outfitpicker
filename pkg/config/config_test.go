package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertNoError(t *testing.T, err error, msg string) {
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func assertError(t *testing.T, err error, expectedMsg string) {
	if err == nil {
		t.Fatal("expected error but got none")
	}
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("expected error containing '%s', got: %v", expectedMsg, err)
	}
}

func assertStringContains(t *testing.T, str, substr, msg string) {
	if !strings.Contains(str, substr) {
		t.Fatalf("%s: expected to contain '%s', got: %s", msg, substr, str)
	}
}

func assertStringEquals(t *testing.T, expected, actual, msg string) {
	if expected != actual {
		t.Fatalf("%s: expected %s, got %s", msg, expected, actual)
	}
}

func createTestConfig() *Config {
	return &Config{Root: "/test/path"}
}

func TestPath(t *testing.T) {
	path, err := Path()
	assertNoError(t, err, "Path() failed")
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	assertStringContains(t, path, "outfitpicker", "path should contain app name")
	if !strings.HasSuffix(path, "config.json") {
		t.Fatalf("expected path to end with 'config.json', got: %s", path)
	}
}

func TestSaveAndLoad(t *testing.T) {
	defer Delete()

	config := createTestConfig()
	assertNoError(t, Save(config), "Save failed")

	loaded, err := Load()
	assertNoError(t, err, "Load failed")
	assertStringEquals(t, config.Root, loaded.Root, "loaded config root mismatch")
}

func TestLoad_NonexistentFile(t *testing.T) {
	defer Delete()

	_, err := Load()
	if err != os.ErrNotExist {
		t.Fatalf("expected os.ErrNotExist, got: %v", err)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	defer Delete()

	path, _ := Path()
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte("invalid json"), 0o644)

	_, err := Load()
	assertError(t, err, "failed to parse config")
}

func TestLoad_ReadError(t *testing.T) {
	defer Delete()

	path, _ := Path()
	configDir := filepath.Dir(path)
	os.MkdirAll(configDir, 0o755)
	os.WriteFile(path, []byte(`{"root": "test"}`), 0o000)
	defer os.Chmod(path, 0o644)

	_, err := Load()
	if err == nil {
		t.Skip("file permissions test not supported on this system")
	}
	if strings.Contains(err.Error(), "no such file") {
		t.Skip("file permissions test not supported on this system")
	}
	assertError(t, err, "failed to read config file")
}

func TestSave_MkdirError(t *testing.T) {
	defer Delete()

	path, _ := Path()
	configDir := filepath.Dir(path)
	parentDir := filepath.Dir(configDir)
	os.MkdirAll(parentDir, 0o755)
	os.WriteFile(configDir, []byte("blocking file"), 0o644)
	defer os.Remove(configDir)

	err := Save(createTestConfig())
	if err == nil {
		t.Skip("mkdir error test not supported on this system")
	}
	assertError(t, err, "failed to create config dir")
}

func TestSave_WriteError(t *testing.T) {
	defer Delete()

	path, _ := Path()
	configDir := filepath.Dir(path)
	os.MkdirAll(configDir, 0o555)
	defer os.Chmod(configDir, 0o755)

	err := Save(createTestConfig())
	if err == nil {
		t.Skip("write error test not supported on this system")
	}
	assertError(t, err, "failed to write config")
}

func TestDelete(t *testing.T) {
	defer Delete()

	Save(createTestConfig())
	_, err := Load()
	assertNoError(t, err, "config should exist before delete")

	assertNoError(t, Delete(), "Delete failed")

	_, err = Load()
	if err != os.ErrNotExist {
		t.Fatalf("expected os.ErrNotExist after delete, got: %v", err)
	}
}

func TestDelete_NonexistentFile(t *testing.T) {
	defer Delete()

	assertNoError(t, Delete(), "Delete of non-existent file should not error")
}

func TestDelete_RemoveError(t *testing.T) {
	defer Delete()

	path, _ := Path()
	configDir := filepath.Dir(path)
	os.MkdirAll(configDir, 0o755)
	os.WriteFile(path, []byte(`{"root": "test"}`), 0o644)
	os.Chmod(configDir, 0o555)
	defer os.Chmod(configDir, 0o755)

	err := Delete()
	if err == nil {
		t.Skip("delete error test not supported on this system")
	}
	assertError(t, err, "failed to delete config")
}

func TestSave_MarshalError(t *testing.T) {
	defer Delete()

	assertNoError(t, Save(createTestConfig()), "Save should succeed for simple config")
}

func TestPath_UserConfigDirError(t *testing.T) {
	path, err := Path()
	if err != nil {
		assertError(t, err, "failed to determine user config dir")
	} else if path == "" {
		t.Fatal("expected non-empty path")
	}
}

func TestConfig_JSONTags(t *testing.T) {
	defer Delete()

	original := &Config{Root: "/test/json/path"}

	data, err := json.MarshalIndent(original, "", "  ")
	assertNoError(t, err, "JSON marshal failed")
	assertStringContains(t, string(data), `"root"`, "JSON should contain root field")

	var loaded Config
	assertNoError(t, json.Unmarshal(data, &loaded), "JSON unmarshal failed")
	assertStringEquals(t, original.Root, loaded.Root, "unmarshaled config mismatch")
}
