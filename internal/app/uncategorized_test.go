package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/storage"
	"github.com/dh85/outfitpicker/pkg/config"
)

func TestUncategorizedScenarios(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(rootDir string) error
		expectedError  string
		expectedOutput []string
	}{
		{
			name: "scenario 1 - all categorized",
			setup: func(rootDir string) error {
				// Create categories with files
				cat1 := filepath.Join(rootDir, "Shirts")
				cat2 := filepath.Join(rootDir, "Pants")
				os.MkdirAll(cat1, 0755)
				os.MkdirAll(cat2, 0755)

				// Add files to categories
				os.WriteFile(filepath.Join(cat1, "shirt1.jpg"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(cat2, "pants1.jpg"), []byte("test"), 0644)
				return nil
			},
			expectedOutput: []string{"Outfit Folders", "Shirts", "Pants"},
		},
		{
			name: "scenario 2 - no categories, only uncategorized",
			setup: func(rootDir string) error {
				// Create uncategorized files only
				os.WriteFile(filepath.Join(rootDir, "outfit1.jpg"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(rootDir, "outfit2.jpg"), []byte("test"), 0644)
				return nil
			},
			expectedOutput: []string{"Your Outfits", "2 outfits available"},
		},
		{
			name: "scenario 3 - mixed categorized and uncategorized",
			setup: func(rootDir string) error {
				// Create category with files
				cat1 := filepath.Join(rootDir, "Shirts")
				os.MkdirAll(cat1, 0755)
				os.WriteFile(filepath.Join(cat1, "shirt1.jpg"), []byte("test"), 0644)

				// Create uncategorized files
				os.WriteFile(filepath.Join(rootDir, "outfit1.jpg"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(rootDir, "outfit2.jpg"), []byte("test"), 0644)
				return nil
			},
			expectedOutput: []string{"Outfit Folders", "Shirts", "Other Outfits", "2 files"},
		},
		{
			name: "scenario 4 - empty categories with uncategorized",
			setup: func(rootDir string) error {
				// Create empty categories
				cat1 := filepath.Join(rootDir, "Shirts")
				cat2 := filepath.Join(rootDir, "Pants")
				os.MkdirAll(cat1, 0755)
				os.MkdirAll(cat2, 0755)

				// Create uncategorized files
				os.WriteFile(filepath.Join(rootDir, "outfit1.jpg"), []byte("test"), 0644)
				return nil
			},
			expectedOutput: []string{"All categories are empty", "Your Outfits"},
		},
		{
			name: "scenario 5 - no files at all",
			setup: func(rootDir string) error {
				// Create empty categories
				cat1 := filepath.Join(rootDir, "Shirts")
				os.MkdirAll(cat1, 0755)
				return nil
			},
			expectedError: "all categories are empty and no uncategorized files found",
		},
		{
			name: "scenario 5b - empty categories and no uncategorized",
			setup: func(rootDir string) error {
				// Create empty categories
				cat1 := filepath.Join(rootDir, "Shirts")
				cat2 := filepath.Join(rootDir, "Pants")
				os.MkdirAll(cat1, 0755)
				os.MkdirAll(cat2, 0755)
				return nil
			},
			expectedError: "all categories are empty and no uncategorized files found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated test environment
			tempDir := t.TempDir()
			configDir := filepath.Join(tempDir, "config")
			t.Setenv("XDG_CONFIG_HOME", configDir)
			os.MkdirAll(configDir, 0755)
			config.Delete()
			defer config.Delete()

			rootDir := filepath.Join(tempDir, "outfits")
			os.MkdirAll(rootDir, 0755)

			// Setup test scenario
			if err := tt.setup(rootDir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			// Test the scenario
			var stdout bytes.Buffer
			stdin := strings.NewReader("q\n") // Always quit to avoid hanging

			err := Run(rootDir, "", stdin, &stdout)
			output := stdout.String()

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check expected output
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestUncategorizedRandomSelection(t *testing.T) {
	// Create isolated test environment
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	t.Setenv("XDG_CONFIG_HOME", configDir)
	os.MkdirAll(configDir, 0755)
	config.Delete()
	defer config.Delete()

	rootDir := filepath.Join(tempDir, "outfits")
	os.MkdirAll(rootDir, 0755)

	// Create uncategorized files
	os.WriteFile(filepath.Join(rootDir, "outfit1.jpg"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(rootDir, "outfit2.jpg"), []byte("test"), 0644)

	var stdout bytes.Buffer
	stdin := strings.NewReader("r\nk\n") // Random, then keep

	err := Run(rootDir, "", stdin, &stdout)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "I picked this outfit for you") {
		t.Errorf("expected random selection output, got:\n%s", output)
	}
	if !strings.Contains(output, "Great choice! I've saved") {
		t.Errorf("expected keep confirmation, got:\n%s", output)
	}
}

func TestUncategorizedSelectedUnselected(t *testing.T) {
	// Create isolated test environment
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	t.Setenv("XDG_CONFIG_HOME", configDir)
	os.MkdirAll(configDir, 0755)
	config.Delete()
	defer config.Delete()

	rootDir := filepath.Join(tempDir, "outfits")
	os.MkdirAll(rootDir, 0755)

	// Create uncategorized files
	os.WriteFile(filepath.Join(rootDir, "outfit1.jpg"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(rootDir, "outfit2.jpg"), []byte("test"), 0644)

	// Pre-select one file
	cache, _ := storage.NewManager(rootDir)
	cache.Add("outfit1.jpg", "UNCATEGORIZED")

	tests := []struct {
		name           string
		input          string
		expectedOutput []string
	}{
		{
			name:           "show selected",
			input:          "s\n",
			expectedOutput: []string{"Outfits You've Already Picked", "outfit1.jpg"},
		},
		{
			name:           "show unselected",
			input:          "u\n",
			expectedOutput: []string{"Outfits You Haven't Picked Yet", "outfit2.jpg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			stdin := strings.NewReader(tt.input)

			err := Run(rootDir, "", stdin, &stdout)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			output := stdout.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestManualSelection(t *testing.T) {
	// Create isolated test environment
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	t.Setenv("XDG_CONFIG_HOME", configDir)
	os.MkdirAll(configDir, 0755)
	config.Delete()
	defer config.Delete()

	rootDir := filepath.Join(tempDir, "outfits")
	os.MkdirAll(rootDir, 0755)

	// Create mixed scenario
	cat1 := filepath.Join(rootDir, "Shirts")
	os.MkdirAll(cat1, 0755)
	os.WriteFile(filepath.Join(cat1, "shirt1.jpg"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(rootDir, "outfit1.jpg"), []byte("test"), 0644)

	// Pre-select shirt1.jpg to test "already selected" message
	cache, _ := storage.NewManager(rootDir)
	cache.Add("shirt1.jpg", cat1)

	tests := []struct {
		name           string
		input          string
		expectedOutput []string
	}{
		{
			name:           "select already selected file",
			input:          "m\n1\n",
			expectedOutput: []string{"Choose Your Outfit", "Shirts", "shirt1.jpg (already picked)", "You've already picked"},
		},
		{
			name:           "select uncategorized file",
			input:          "m\n2\n",
			expectedOutput: []string{"Choose Your Outfit", "Other Outfits", "outfit1.jpg", "Great choice! I've saved 'outfit1.jpg' from Uncategorized"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			stdin := strings.NewReader(tt.input)

			err := Run(rootDir, "", stdin, &stdout)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			output := stdout.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestDeleteFile(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput []string
		fileExists     bool
	}{
		{
			name:           "delete with confirmation",
			input:          "r\nd\nyes\n",
			expectedOutput: []string{"Are you sure you want to permanently delete", "Deleted 'test-outfit.jpg'"},
			fileExists:     false,
		},
		{
			name:           "delete cancelled",
			input:          "r\nd\nno\n",
			expectedOutput: []string{"Are you sure you want to permanently delete", "Okay, I won't delete it"},
			fileExists:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated test environment for each test
			tempDir := t.TempDir()
			configDir := filepath.Join(tempDir, "config")
			t.Setenv("XDG_CONFIG_HOME", configDir)
			os.MkdirAll(configDir, 0755)
			config.Delete()
			defer config.Delete()

			rootDir := filepath.Join(tempDir, "outfits")
			os.MkdirAll(rootDir, 0755)

			// Create test file
			testFile := filepath.Join(rootDir, "test-outfit.jpg")
			os.WriteFile(testFile, []byte("test"), 0644)

			var stdout bytes.Buffer
			stdin := strings.NewReader(tt.input)

			err := Run(rootDir, "", stdin, &stdout)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			output := stdout.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got:\n%s", expected, output)
				}
			}

			// Check file existence
			_, err = os.Stat(testFile)
			fileExists := err == nil
			if fileExists != tt.fileExists {
				t.Errorf("expected file exists=%v, got exists=%v", tt.fileExists, fileExists)
			}
		})
	}
}

func TestMixedRandomSelection(t *testing.T) {
	// Create isolated test environment
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	t.Setenv("XDG_CONFIG_HOME", configDir)
	os.MkdirAll(configDir, 0755)
	config.Delete()
	defer config.Delete()

	rootDir := filepath.Join(tempDir, "outfits")
	os.MkdirAll(rootDir, 0755)

	// Create mixed scenario
	cat1 := filepath.Join(rootDir, "Shirts")
	os.MkdirAll(cat1, 0755)
	os.WriteFile(filepath.Join(cat1, "shirt1.jpg"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(rootDir, "outfit1.jpg"), []byte("test"), 0644)

	var stdout bytes.Buffer
	stdin := strings.NewReader("r\nk\n") // Random across all, then keep

	err := Run(rootDir, "", stdin, &stdout)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "I picked this outfit for you") {
		t.Errorf("expected random selection output, got:\n%s", output)
	}

	// Should show either "From your Shirts collection" or "From your other outfits"
	hasCategory := strings.Contains(output, "From your Shirts collection") || strings.Contains(output, "From your other outfits")
	if !hasCategory {
		t.Errorf("expected category or uncategorized indication, got:\n%s", output)
	}
}

func TestCrossplatformPaths(t *testing.T) {
	// Test that file paths work correctly across different operating systems
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	t.Setenv("XDG_CONFIG_HOME", configDir)
	os.MkdirAll(configDir, 0755)
	config.Delete()
	defer config.Delete()

	rootDir := filepath.Join(tempDir, "outfits")
	os.MkdirAll(rootDir, 0755)

	// Create files with various names that might cause issues
	testFiles := []string{
		"outfit with spaces.jpg",
		"outfit-with-dashes.jpg",
		"outfit_with_underscores.jpg",
		"UPPERCASE.JPG",
		"lowercase.jpg",
	}

	for _, fileName := range testFiles {
		os.WriteFile(filepath.Join(rootDir, fileName), []byte("test"), 0644)
	}

	var stdout bytes.Buffer
	stdin := strings.NewReader("s\n") // Show selected (should be empty initially)

	err := Run(rootDir, "", stdin, &stdout)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "You haven't picked any outfits from here yet") {
		t.Errorf("expected no selected files message, got:\n%s", output)
	}
}
