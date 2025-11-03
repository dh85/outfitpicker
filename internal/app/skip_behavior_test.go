package app

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/storage"
)

func TestSkipBehavior(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"skip_then_keep", "s\nk\n", []string{"Skipped", "Great choice!"}},
		{"skip_all_retry", "s\ns\ny\nk\n", []string{"Try again", "Great choice!"}},
		{"skip_all_decline", "s\ns\nn\n", []string{"Try again"}},
		{"keep_immediately", "k\n", []string{"Great choice!"}},
		{"quit_immediately", "q\n", []string{"Exiting"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create test files
			catDir := filepath.Join(tempDir, "casual")
			os.MkdirAll(catDir, 0755)
			os.WriteFile(filepath.Join(catDir, "outfit1.jpg"), []byte("test"), 0644)
			os.WriteFile(filepath.Join(catDir, "outfit2.jpg"), []byte("test"), 0644)

			cache, _ := storage.NewManager(tempDir)
			categories := []string{catDir}
			var uncategorized []string

			stdout := &bytes.Buffer{}
			pr := &prompter{r: bufio.NewReader(strings.NewReader(tt.input)), w: stdout}

			randomAcrossAll(categories, uncategorized, cache, pr, stdout)

			output := stdout.String()
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected %q in output, got: %s", expected, output)
				}
			}
		})
	}
}
