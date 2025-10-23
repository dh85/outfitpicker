package cli

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestFirstRunWizard_EnhancedUI(t *testing.T) {
	// Test welcome message with enhanced UI
	var buf bytes.Buffer
	printWelcomeMessage(&buf)

	output := buf.String()
	expectedElements := []string{
		"First Time Setup",
		"Welcome to outfitpicker",
		"outfit directory",
		"Example:",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("expected welcome message to contain %q", element)
		}
	}
}

func TestGetPathInput_EnhancedUI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "valid path",
			input:    "/valid/path\n",
			expected: "/valid/path",
			hasError: false,
		},
		{
			name:     "empty path",
			input:    "\n",
			expected: "",
			hasError: false,
		},
		{
			name:     "whitespace path",
			input:    "   \n",
			expected: "",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			input := bufio.NewReader(strings.NewReader(tt.input))

			result, err := getPathInput(input, &buf)

			if tt.hasError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}

			output := buf.String()
			if !strings.Contains(output, "Root path:") {
				t.Error("expected enhanced path prompt")
			}
		})
	}
}

func TestHandleExistingPath_EnhancedUI(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "wizard_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	var buf bytes.Buffer
	info, _ := os.Stat(tmpDir)

	result, shouldContinue, err := handleExistingPath(tmpDir, info, &buf)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if shouldContinue {
		t.Error("expected shouldContinue to be false")
	}
	if result != tmpDir {
		t.Errorf("expected result to be %q, got %q", tmpDir, result)
	}

	output := buf.String()
	if !strings.Contains(output, "Found existing directory") {
		t.Error("expected success message for existing directory")
	}
}

func TestHandleExistingPath_NotDirectory(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "wizard_test")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	var buf bytes.Buffer
	info, _ := os.Stat(tmpFile.Name())

	result, shouldContinue, err := handleExistingPath(tmpFile.Name(), info, &buf)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !shouldContinue {
		t.Error("expected shouldContinue to be true")
	}
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}

	output := buf.String()
	if !strings.Contains(output, "not a directory") {
		t.Error("expected error message for file instead of directory")
	}
}

func TestHandleNonExistentPath_EnhancedUI(t *testing.T) {
	tmpDir := "/tmp/wizard_test_nonexistent"
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name           string
		input          string
		expectedResult string
		shouldContinue bool
	}{
		{
			name:           "user says yes",
			input:          "y\n",
			expectedResult: tmpDir,
			shouldContinue: false,
		},
		{
			name:           "user says no",
			input:          "n\n",
			expectedResult: "",
			shouldContinue: true,
		},
		{
			name:           "user says YES",
			input:          "YES\n",
			expectedResult: tmpDir,
			shouldContinue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			os.RemoveAll(tmpDir)

			var buf bytes.Buffer
			input := bufio.NewReader(strings.NewReader(tt.input))

			result, shouldContinue, err := handleNonExistentPath(tmpDir, input, &buf)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if shouldContinue != tt.shouldContinue {
				t.Errorf("expected shouldContinue %v, got %v", tt.shouldContinue, shouldContinue)
			}
			if result != tt.expectedResult {
				t.Errorf("expected result %q, got %q", tt.expectedResult, result)
			}

			output := buf.String()
			if !strings.Contains(output, "Path does not exist") {
				t.Error("expected warning message for non-existent path")
			}
			if !strings.Contains(output, "Create it now?") {
				t.Error("expected creation prompt")
			}

			if tt.input == "y\n" || tt.input == "YES\n" {
				if !strings.Contains(output, "Directory created successfully") {
					t.Error("expected success message for directory creation")
				}
			}
		})
	}
}

func TestFinalizeSetup_EnhancedUI(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wizard_finalize_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	var buf bytes.Buffer

	result, shouldContinue, err := finalizeSetup(tmpDir, &buf)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if shouldContinue {
		t.Error("expected shouldContinue to be false")
	}
	if result != tmpDir {
		t.Errorf("expected result to be %q, got %q", tmpDir, result)
	}

	output := buf.String()
	expectedMessages := []string{
		"Setup completed successfully!",
		"Your outfit directory is set to:",
		tmpDir,
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("expected finalize output to contain %q", msg)
		}
	}
}

func TestEnsureCacheAtRoot_EnhancedUI(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wizard_cache_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	var buf bytes.Buffer

	err = EnsureCacheAtRoot(tmpDir, &buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Created cache at") {
		t.Error("expected cache creation message")
	}
}

func TestShouldUseColors_Wizard(t *testing.T) {
	tests := []struct {
		name     string
		term     string
		noColor  string
		expected bool
	}{
		{"normal terminal", "xterm-256color", "", true},
		{"dumb terminal", "dumb", "", false},
		{"empty term", "", "", false},
		{"no color set", "xterm", "1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origTerm := os.Getenv("TERM")
			origNoColor := os.Getenv("NO_COLOR")
			defer func() {
				os.Setenv("TERM", origTerm)
				os.Setenv("NO_COLOR", origNoColor)
			}()

			os.Setenv("TERM", tt.term)
			os.Setenv("NO_COLOR", tt.noColor)

			result := shouldUseColors()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Benchmark tests for wizard UI performance
func BenchmarkPrintWelcomeMessage(b *testing.B) {
	var buf bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		printWelcomeMessage(&buf)
	}
}

func BenchmarkGetPathInput(b *testing.B) {
	var buf bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		input := bufio.NewReader(strings.NewReader("/test/path\n"))
		getPathInput(input, &buf)
	}
}
