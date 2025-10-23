package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestErrorPaths(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping on macOS due to permission handling differences")
	}
	
	// Test Save with invalid directory by setting invalid config home
	t.Setenv("XDG_CONFIG_HOME", "/dev/null/invalid")
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", "NUL:\\invalid")
	}
	
	err := Save(&Config{Root: "/test"})
	if err == nil {
		t.Error("expected error when saving to invalid path")
	}
}

func TestDeleteNonExistentConfig(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	// Delete non-existent config should not error
	err := Delete()
	if err != nil {
		t.Errorf("Delete of non-existent config should not error: %v", err)
	}
}

func TestPathErrors(t *testing.T) {
	// Test with invalid environment
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "")
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", "")
		t.Setenv("APPDATA", "")
	}
	
	_, err := Path()
	if err == nil && runtime.GOOS != "windows" {
		t.Error("expected error when no config directory can be determined")
	}
}

func TestCorruptedConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	// Create corrupted config file
	configPath := filepath.Join(tempDir, "outfitpicker", "config.json")
	os.MkdirAll(filepath.Dir(configPath), 0755)
	os.WriteFile(configPath, []byte("invalid json {"), 0644)
	
	_, err := Load()
	if err == nil {
		t.Error("expected error when loading corrupted config")
	}
}

func TestReadOnlyConfigDir(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("Skipping read-only test on Windows and macOS")
	}
	
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "outfitpicker")
	configFile := filepath.Join(configDir, "config.json")
	os.MkdirAll(configDir, 0755)
	
	// Create a file where config should go, then make it read-only
	os.WriteFile(configFile, []byte("existing"), 0644)
	os.Chmod(configFile, 0444)
	defer os.Chmod(configFile, 0644) // Restore for cleanup
	
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	err := Save(&Config{Root: "/test"})
	if err == nil {
		t.Error("expected error when saving to read-only directory")
	}
}

func TestWindowsSpecificPaths(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test")
	}
	
	// Test APPDATA fallback
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("APPDATA", "C:\\Users\\test\\AppData\\Roaming")
	
	path, err := Path()
	if err != nil {
		t.Errorf("Path() failed on Windows: %v", err)
	}
	
	if !strings.Contains(path, "AppData") {
		t.Errorf("expected Windows path to contain AppData, got: %s", path)
	}
}

func TestUnixSpecificPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test")
	}
	
	// Test HOME fallback
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/home/test")
	
	path, err := Path()
	if err != nil {
		t.Errorf("Path() failed on Unix: %v", err)
	}
	
	if !strings.Contains(path, "/home/test") {
		t.Errorf("expected Unix path to contain HOME, got: %s", path)
	}
}

func TestLargeConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	// Create config with very long root path
	longRoot := strings.Repeat("a", 1000)
	config := &Config{Root: longRoot}
	
	err := Save(config)
	if err != nil {
		t.Errorf("Save with long root failed: %v", err)
	}
	
	loaded, err := Load()
	if err != nil {
		t.Errorf("Load after saving long root failed: %v", err)
	}
	
	if loaded.Root != longRoot {
		t.Error("Long root path was not preserved")
	}
}

func TestConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	// Test concurrent saves
	done := make(chan bool, 2)
	
	go func() {
		Save(&Config{Root: "/test1"})
		done <- true
	}()
	
	go func() {
		Save(&Config{Root: "/test2"})
		done <- true
	}()
	
	// Wait for both to complete
	<-done
	<-done
	
	// Should be able to load without error
	_, err := Load()
	if err != nil {
		t.Errorf("Load after concurrent saves failed: %v", err)
	}
}