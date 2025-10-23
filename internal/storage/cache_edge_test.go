package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadCorruptedCache(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir)

	// Create corrupted cache file
	os.WriteFile(manager.Path(), []byte("invalid json {"), 0644)

	// Should return empty map for corrupted cache
	result := manager.Load()
	if len(result) != 0 {
		t.Error("corrupted cache should return empty map")
	}
}

func TestUnicodeFilenames(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir)

	// Test with Unicode filenames
	unicodeFiles := []string{
		"файл.jpg",   // Russian
		"文件.jpg",     // Chinese
		"🎉emoji.jpg", // Emoji
	}

	category := filepath.Join(tempDir, "unicode")
	manager.Save(Map{category: unicodeFiles})

	result := manager.Load()
	if len(result[category]) != len(unicodeFiles) {
		t.Error("Unicode filenames not preserved")
	}

	for i, file := range result[category] {
		if file != unicodeFiles[i] {
			t.Errorf("Unicode filename not preserved: expected %s, got %s", unicodeFiles[i], file)
		}
	}
}

func TestAddDuplicates(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir)

	category := filepath.Join(tempDir, "test")

	// Add same file multiple times
	manager.Add("file1.jpg", category)
	manager.Add("file1.jpg", category)
	manager.Add("file1.jpg", category)

	result := manager.Load()
	if len(result[category]) != 1 {
		t.Errorf("expected 1 file, got %d", len(result[category]))
	}
}

func TestPathPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir)

	// Check that cache file has correct permissions
	manager.Save(Map{})

	info, err := os.Stat(manager.Path())
	if err != nil {
		t.Fatalf("failed to stat cache file: %v", err)
	}

	mode := info.Mode()
	if mode.Perm() != 0600 {
		t.Errorf("expected cache file permissions 0600, got %o", mode.Perm())
	}
}

func TestContainsFunction(t *testing.T) {
	tests := []struct {
		slice    []string
		item     string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{[]string{"test"}, "test", true},
		{[]string{"Test"}, "test", false}, // Case sensitive
	}

	for _, tt := range tests {
		result := contains(tt.slice, tt.item)
		if result != tt.expected {
			t.Errorf("contains(%v, %s) = %v, want %v", tt.slice, tt.item, result, tt.expected)
		}
	}
}
