package system

import (
	"os"
	"testing"
)

func TestDefaultFileManager_ReadDir(t *testing.T) {
	fm := NewDefaultFileManager()

	// Test with temp directory
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(tmpDir+"/test1.avatar", []byte("test"), 0644)
	os.WriteFile(tmpDir+"/test2.avatar", []byte("test"), 0644)
	os.Mkdir(tmpDir+"/subdir", 0755)

	entries, err := fm.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	// Test with non-existent directory
	_, err = fm.ReadDir("/nonexistent/path")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

func TestDefaultFileManager_FileExists(t *testing.T) {
	fm := NewDefaultFileManager()

	// Test with temp file
	tmpFile, _ := os.CreateTemp("", "test")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	if !fm.FileExists(tmpFile.Name()) {
		t.Error("expected file to exist")
	}

	// Test with non-existent file
	if fm.FileExists("/nonexistent/file") {
		t.Error("expected file to not exist")
	}
}
