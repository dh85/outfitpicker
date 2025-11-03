package cli

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFirstRunWizardEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		setup func(*testing.T) string
	}{
		{
			name:  "empty input",
			input: "\n\nq\n",
			setup: func(t *testing.T) string { return "" },
		},
		{
			name:  "whitespace only",
			input: "   \n\nq\n",
			setup: func(t *testing.T) string { return "" },
		},
		{
			name:  "very long path",
			input: strings.Repeat("a", 500) + "\ny\n",
			setup: func(t *testing.T) string { return "" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())

			stdin := strings.NewReader(tt.input)
			var stdout bytes.Buffer

			_, err := FirstRunWizard(stdin, &stdout)
			// These should either succeed or fail gracefully
			if err != nil && !strings.Contains(err.Error(), "no input provided") {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestExpandUserHomeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		setup    func(*testing.T)
	}{
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "no tilde",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:  "tilde with path",
			input: "~/Documents",
			setup: func(t *testing.T) {
				if runtime.GOOS != "windows" {
					t.Setenv("HOME", "/home/test")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			result, err := ExpandUserHome(tt.input)
			if err != nil {
				t.Errorf("ExpandUserHome failed: %v", err)
			}

			if tt.expected != "" && result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestWindowsSpecificBehavior(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test")
	}

	// Test that tilde expansion is skipped on Windows
	result, err := ExpandUserHome("~/test")
	if err != nil {
		t.Errorf("ExpandUserHome failed on Windows: %v", err)
	}

	if result != "~/test" {
		t.Errorf("expected tilde to be preserved on Windows, got: %s", result)
	}
}

func TestEnsureCacheAtRootErrors(t *testing.T) {
	// Test with invalid root path - this may not always error
	// depending on system permissions, so just ensure it doesn't crash
	var stdout bytes.Buffer
	EnsureCacheAtRoot("/root/invalid/path", &stdout)
	// Test passes if no panic occurs
}

func TestReadLineEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "empty input",
			input:    "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "only newline",
			input:    "\n",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "no newline at end",
			input:    "test",
			expected: "test",
			wantErr:  false,
		},
		{
			name:     "multiple newlines",
			input:    "test\n\n\n",
			expected: "test",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			result, err := readLine(reader)

			if tt.wantErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHandlePathPermissionErrors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tempDir := t.TempDir()
	restrictedDir := filepath.Join(tempDir, "restricted")
	_ = os.MkdirAll(restrictedDir, 0000)                 // No permissions
	defer func() { _ = os.Chmod(restrictedDir, 0755) }() // Restore for cleanup

	// Test that wizard handles permission errors gracefully
	stdin := strings.NewReader(restrictedDir + "\nn\n" + tempDir + "\ny\n")
	var stdout bytes.Buffer

	_, err := FirstRunWizard(stdin, &stdout)
	if err != nil {
		t.Errorf("wizard should handle permission errors gracefully: %v", err)
	}
}

func TestIsYesResponseVariations(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"y", true},
		{"Y", true},
		{"yes", true},
		{"YES", true},
		{"Yes", true},
		{"  y  ", true},
		{"  YES  ", true},
		{"n", false},
		{"no", false},
		{"", false},
		{"maybe", false},
		{"yep", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isYesResponse(tt.input)
			if result != tt.expected {
				t.Errorf("isYesResponse(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPrintWelcomeMessagePlatforms(t *testing.T) {
	var stdout bytes.Buffer
	printWelcomeMessage(&stdout)

	output := stdout.String()
	if !strings.Contains(output, "Welcome to outfitpicker") {
		t.Error("welcome message should mention first time running")
	}

	// Should contain platform-appropriate example
	if runtime.GOOS == "windows" {
		if !strings.Contains(output, "C:\\") {
			t.Error("Windows welcome message should contain Windows path example")
		}
	} else {
		if !strings.Contains(output, "/Users/") {
			t.Error("Unix welcome message should contain Unix path example")
		}
	}
}
