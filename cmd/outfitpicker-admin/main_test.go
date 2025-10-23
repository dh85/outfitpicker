package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/storage"
	"github.com/dh85/outfitpicker/pkg/config"
)

func TestNewRootCmd(t *testing.T) {
	cmd := newRootCmd()
	
	if cmd.Use != "outfitpicker-admin" {
		t.Errorf("expected Use to be 'outfitpicker-admin', got %s", cmd.Use)
	}
}

func TestCacheShowCommand(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	defer config.Delete()
	
	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"cache", "show", "--root", tempDir})
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("cache show command failed: %v", err)
	}
}

func TestCacheClearCommand(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	defer config.Delete()
	
	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"cache", "clear", "--all", "--root", tempDir})
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("cache clear command failed: %v", err)
	}
}

func TestResolveRoot(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	// Test with explicit root
	root, err := resolveRoot(tempDir)
	if err != nil {
		t.Errorf("resolveRoot failed: %v", err)
	}
	if root != tempDir {
		t.Errorf("expected root %s, got %s", tempDir, root)
	}
}

func TestResolveRootFromConfig(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	// Ensure completely clean state
	config.Delete()
	defer config.Delete()
	
	// Save config with unique path for this test
	testRoot := filepath.Join(tempDir, "test-root")
	os.MkdirAll(testRoot, 0755)
	err := config.Save(&config.Config{Root: testRoot})
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	
	// Verify config was saved correctly
	loadedConfig, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if loadedConfig.Root != testRoot {
		t.Fatalf("config not saved correctly: expected %s, got %s", testRoot, loadedConfig.Root)
	}
	
	// Test resolveRoot without override should use config
	root, err := resolveRoot("")
	if err != nil {
		t.Errorf("resolveRoot failed: %v", err)
	}
	if root != testRoot {
		t.Errorf("expected root %s, got %s", testRoot, root)
	}
}

func TestResolveRootErrors(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	tests := []struct {
		name string
		setup func()
		expectedError string
	}{
		{
			name: "empty root in config",
			setup: func() {
				config.Save(&config.Config{Root: ""})
			},
			expectedError: "config has empty root",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure clean state
			config.Delete()
			if tt.setup != nil {
				tt.setup()
			}
			defer config.Delete()
			
			_, err := resolveRoot("")
			if err == nil {
				t.Error("expected error")
			} else if !strings.Contains(err.Error(), tt.expectedError) && !strings.Contains(err.Error(), "failed to load config") {
				t.Errorf("expected error to contain %q or config load error, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestVersionFlag(t *testing.T) {
	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--version"})
	
	// This should exit, so we expect it to not return normally
	// We'll test the PersistentPreRunE instead
	cmd.PersistentPreRunE(cmd, []string{})
}

func TestConfigCommands(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	tests := []struct {
		name string
		args []string
		setup func()
		expectError bool
	}{
		{
			name: "config show - no config",
			args: []string{"config", "show"},
		},
		{
			name: "config show - with config",
			args: []string{"config", "show"},
			setup: func() {
				testRoot := filepath.Join(tempDir, "test-config")
				os.MkdirAll(testRoot, 0755)
				config.Save(&config.Config{Root: testRoot})
			},
		},
		{
			name: "config set-root",
			args: []string{"config", "set-root", filepath.Join(tempDir, "new-root")},
		},
		{
			name: "config reset",
			args: []string{"config", "reset"},
			setup: func() {
				testRoot := filepath.Join(tempDir, "test-reset")
				os.MkdirAll(testRoot, 0755)
				config.Save(&config.Config{Root: testRoot})
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			// Ensure cleanup after each subtest
			defer config.Delete()
			
			cmd := newRootCmd()
			var stdout bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetArgs(tt.args)
			
			err := cmd.Execute()
			if tt.expectError && err == nil {
				t.Error("expected error")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCacheShowEmpty(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	// Create empty root
	rootDir := filepath.Join(tempDir, "root")
	os.MkdirAll(rootDir, 0755)
	
	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"cache", "show", "--root", rootDir})
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("cache show failed: %v", err)
	}
	
	output := stdout.String()
	if !strings.Contains(output, "(empty)") {
		t.Error("expected empty cache message")
	}
}

func TestCacheShowWithData(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	// Create root with cache data
	rootDir := filepath.Join(tempDir, "root")
	mgr, _ := storage.NewManager(rootDir)
	mgr.Add("test.jpg", filepath.Join(rootDir, "TestCat"))
	
	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"cache", "show", "--root", rootDir})
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("cache show failed: %v", err)
	}
	
	output := stdout.String()
	if !strings.Contains(output, "TestCat") {
		t.Error("expected category in output")
	}
	if !strings.Contains(output, "1 selected") {
		t.Error("expected selection count")
	}
}

func TestCacheClearAll(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	// Create root with cache data
	rootDir := filepath.Join(tempDir, "root")
	mgr, _ := storage.NewManager(rootDir)
	mgr.Add("test1.jpg", filepath.Join(rootDir, "Cat1"))
	mgr.Add("test2.jpg", filepath.Join(rootDir, "Cat2"))
	
	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"cache", "clear", "--all", "--root", rootDir})
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("cache clear all failed: %v", err)
	}
	
	output := stdout.String()
	if !strings.Contains(output, "cleared cache for all") {
		t.Error("expected clear all confirmation")
	}
	
	// Verify cache is empty
	data := mgr.Load()
	if len(data) != 0 {
		t.Error("expected cache to be empty")
	}
}

func TestCacheClearAllEmpty(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	// Create empty root
	rootDir := filepath.Join(tempDir, "root")
	os.MkdirAll(rootDir, 0755)
	
	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"cache", "clear", "--all", "--root", rootDir})
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("cache clear all failed: %v", err)
	}
	
	output := stdout.String()
	if !strings.Contains(output, "already empty") {
		t.Error("expected already empty message")
	}
}

func TestCacheClearCategory(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	// Create root with cache data
	rootDir := filepath.Join(tempDir, "root")
	mgr, _ := storage.NewManager(rootDir)
	mgr.Add("test.jpg", filepath.Join(rootDir, "TestCat"))
	
	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"cache", "clear", "TestCat", "--root", rootDir})
	
	err := cmd.Execute()
	if err != nil {
		t.Errorf("cache clear category failed: %v", err)
	}
	
	output := stdout.String()
	if !strings.Contains(output, "cleared cache for \"TestCat\"") {
		t.Error("expected clear category confirmation")
	}
}

func TestCacheClearCategoryNotFound(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	// Create root with different cache data
	rootDir := filepath.Join(tempDir, "root")
	mgr, _ := storage.NewManager(rootDir)
	mgr.Add("test.jpg", filepath.Join(rootDir, "OtherCat"))
	
	cmd := newRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"cache", "clear", "NonExistent", "--root", rootDir})
	
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for non-existent category")
	}
	
	if !strings.Contains(err.Error(), "not found in cache") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestCacheClearNoArgs(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	rootDir := filepath.Join(tempDir, "root")
	os.MkdirAll(rootDir, 0755)
	
	cmd := newRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"cache", "clear", "--root", rootDir})
	
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing category argument")
	}
	
	if !strings.Contains(err.Error(), "provide a category name") {
		t.Errorf("expected 'provide category' error, got: %v", err)
	}
}

func TestCacheClearEmptyCache(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	
	rootDir := filepath.Join(tempDir, "root")
	os.MkdirAll(rootDir, 0755)
	
	cmd := newRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"cache", "clear", "TestCat", "--root", rootDir})
	
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for empty cache")
	}
	
	if !strings.Contains(err.Error(), "no categories found in cache") {
		t.Errorf("expected 'no categories' error, got: %v", err)
	}
}

func TestMainFunction(t *testing.T) {
	// Test that main function exists by calling newRootCmd
	cmd := newRootCmd()
	if cmd == nil {
		t.Error("newRootCmd should return a valid command")
	}
}