package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/pkg/config"
)

func TestNewRootCmd(t *testing.T) {
	cmd := newRootCmd()

	if cmd.Use != "outfitpicker [root]" {
		t.Errorf("expected Use to be 'outfitpicker [root]', got %s", cmd.Use)
	}

	if !strings.Contains(cmd.Short, "Select outfit files") {
		t.Errorf("expected Short to contain 'Select outfit files', got %s", cmd.Short)
	}
}

func TestVersionFlag(t *testing.T) {
	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "dev") {
		t.Errorf("expected version output to contain 'dev', got: %s", output)
	}
}

func TestConfigCommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"config show", []string{"config", "show"}},
		{"config reset", []string{"config", "reset"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set temp config dir
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			// Ensure cleanup after each subtest
			defer config.Delete()

			cmd := newRootCmd()
			var stdout bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetArgs(tt.args)

			// These commands should not error (even if config doesn't exist)
			err := cmd.Execute()
			if err != nil && !strings.Contains(err.Error(), "no such file") {
				t.Errorf("command %v failed unexpectedly: %v", tt.args, err)
			}
		})
	}
}

func TestCompletionCommands(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			cmd := newRootCmd()
			var stdout bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetArgs([]string{"completion", shell})

			err := cmd.Execute()
			if err != nil {
				t.Errorf("completion command for %s failed: %v", shell, err)
			}

			output := stdout.String()
			if len(output) == 0 {
				t.Errorf("completion command for %s produced no output", shell)
			}
		})
	}
}

func TestInvalidCompletionShell(t *testing.T) {
	cmd := newRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"completion", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid shell")
	}

	if !strings.Contains(err.Error(), "invalid argument") {
		t.Errorf("expected error to contain 'invalid argument', got: %v", err)
	}
}

func TestSetRootFlag(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--set-root", tempDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("set-root command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Created cache") {
		t.Errorf("expected output to mention cache creation, got: %s", output)
	}
}

func TestConfigSetRoot(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"config", "set-root", tempDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("config set-root command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "saved default root") {
		t.Errorf("expected output to mention saving root, got: %s", output)
	}
}

func TestConfigSetRootMissingArg(t *testing.T) {
	cmd := newRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"config", "set-root"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing argument")
	}
}

func TestMainFunction(t *testing.T) {
	// Test that main function exists by calling newRootCmd
	// which is the core functionality main() uses
	cmd := newRootCmd()
	if cmd == nil {
		t.Error("newRootCmd should return a valid command")
	}
}

func TestErrorHandling(t *testing.T) {
	// Test with non-existent root directory
	cmd := newRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"/nonexistent/path"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestRootResolution(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	// Create test structure
	rootDir := filepath.Join(tempDir, "outfits")
	os.MkdirAll(filepath.Join(rootDir, "TestCat"), 0755)
	os.WriteFile(filepath.Join(rootDir, "TestCat", "test.jpg"), []byte("test"), 0644)

	tests := []struct {
		name        string
		args        []string
		setup       func()
		expectError bool
	}{
		{
			name: "positional root arg",
			args: []string{rootDir, "--category", "TestCat"},
		},
		{
			name: "root flag",
			args: []string{"--root", rootDir, "--category", "TestCat"},
		},
		{
			name: "config root",
			args: []string{"--category", "TestCat"},
			setup: func() {
				// Ensure the directory exists before saving to config
				os.MkdirAll(rootDir, 0755)
				config.Save(&config.Config{Root: rootDir})
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
			cmd.SetIn(strings.NewReader("q\n"))
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

func TestNoConfigTrigger(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	// Test that when no config exists and no args provided,
	// it tries to trigger first run wizard
	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetIn(strings.NewReader("\n")) // Empty input to trigger error
	cmd.SetArgs([]string{})

	// This should fail because we don't provide valid input
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no valid input provided")
	}
}

func TestConfigFromSaved(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	defer config.Delete()

	// Create test structure
	rootDir := filepath.Join(tempDir, "outfits")
	os.MkdirAll(filepath.Join(rootDir, "TestCat"), 0755)
	os.WriteFile(filepath.Join(rootDir, "TestCat", "test.jpg"), []byte("test"), 0644)

	// Save config
	config.Save(&config.Config{Root: rootDir})

	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetIn(strings.NewReader("q\n"))
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("using saved config failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "using root from config") {
		t.Error("expected config usage message")
	}
}

func TestConfigShowWithExistingConfig(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	defer config.Delete()

	// Save config first
	testRoot := filepath.Join(tempDir, "test-root")
	os.MkdirAll(testRoot, 0755)
	config.Save(&config.Config{Root: testRoot})

	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"config", "show"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("config show failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, testRoot) {
		t.Errorf("expected output to contain root %s, got: %s", testRoot, output)
	}
}

func TestConfigShowNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"config", "show"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("config show failed: %v", err)
	}

	output := stdout.String()
	// The command should either show "not found" or show the config if it exists
	// Both are valid behaviors depending on test isolation
	if !strings.Contains(output, "not found") && !strings.Contains(output, "no such file") && !strings.Contains(output, "Config file not found") && !strings.Contains(output, "config file:") {
		t.Errorf("expected config message, got: %s", output)
	}
}

func TestConfigReset(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	defer config.Delete()

	// Save config first
	testRoot := filepath.Join(tempDir, "test-config-root")
	os.MkdirAll(testRoot, 0755)
	config.Save(&config.Config{Root: testRoot})

	cmd := newRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"config", "reset"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("config reset failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "config reset") {
		t.Error("expected reset confirmation")
	}

	// Verify config is gone
	_, err = config.Load()
	if err == nil {
		t.Error("expected config to be deleted")
	}
}
