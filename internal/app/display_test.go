package app

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDisplay_CategoryInfo(t *testing.T) {
	tests := []struct {
		name          string
		config        AppConfig
		categoryName  string
		totalFiles    int
		selectedFiles int
		expectEmoji   bool
	}{
		{"with emoji", AppConfig{ShowEmojis: true}, "Beach", 10, 3, true},
		{"without emoji", AppConfig{ShowEmojis: false}, "Formal", 5, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			display := NewDisplay(&buf, tt.config)

			display.CategoryInfo(tt.categoryName, tt.totalFiles, tt.selectedFiles)

			output := buf.String()
			if tt.expectEmoji && !strings.Contains(output, "📂") {
				t.Error("expected emoji in output")
			}
			if !tt.expectEmoji && strings.Contains(output, "📂") {
				t.Error("unexpected emoji in output")
			}
			if !strings.Contains(output, tt.categoryName) {
				t.Error("expected category name in output")
			}
		})
	}
}

func TestDisplay_SelectedFiles(t *testing.T) {
	var buf bytes.Buffer
	display := NewDisplay(&buf, DefaultAppConfig())

	// Test empty files
	display.SelectedFiles("Test", []string{})
	if !strings.Contains(buf.String(), "No files have been selected yet") {
		t.Error("expected empty message")
	}

	// Test with files
	buf.Reset()
	files := []string{"file2.jpg", "file1.jpg"}
	display.SelectedFiles("Test", files)
	output := buf.String()

	if !strings.Contains(output, "Previously Selected Files") {
		t.Error("expected header")
	}
	if !strings.Contains(output, "file1.jpg") || !strings.Contains(output, "file2.jpg") {
		t.Error("expected files in output")
	}
}

func TestShouldUseColors(t *testing.T) {
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
