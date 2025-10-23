package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/storage"
)

func TestQuickModeRandom(t *testing.T) {
	tests := []struct {
		name         string
		categoryName string
		setupFunc    func(string) error
		wantErr      bool
		wantOutput   string
	}{
		{
			name:         "all categories random selection",
			categoryName: "",
			setupFunc: func(root string) error {
				return createTestFiles(root, map[string][]string{
					"shirts": {"shirt1.jpg", "shirt2.jpg"},
					"pants":  {"pants1.jpg"},
				}, []string{"loose1.jpg"})
			},
			wantOutput: "✅ Selected:",
		},
		{
			name:         "specific category selection",
			categoryName: "shirts",
			setupFunc: func(root string) error {
				return createTestFiles(root, map[string][]string{
					"shirts": {"shirt1.jpg", "shirt2.jpg"},
					"pants":  {"pants1.jpg"},
				}, nil)
			},
			wantOutput: "✅ Selected:",
		},
		{
			name:         "case insensitive category",
			categoryName: "SHIRTS",
			setupFunc: func(root string) error {
				return createTestFiles(root, map[string][]string{
					"shirts": {"shirt1.jpg"},
				}, nil)
			},
			wantOutput: "✅ Selected:",
		},
		{
			name:         "nonexistent category",
			categoryName: "nonexistent",
			setupFunc: func(root string) error {
				return createTestFiles(root, map[string][]string{
					"shirts": {"shirt1.jpg"},
				}, nil)
			},
			wantErr: true,
		},
		{
			name:         "no files available",
			categoryName: "",
			setupFunc: func(root string) error {
				return os.MkdirAll(filepath.Join(root, "empty"), 0755)
			},
			wantOutput: "No outfits available",
		},
		{
			name:         "only uncategorized files",
			categoryName: "",
			setupFunc: func(root string) error {
				return createTestFiles(root, nil, []string{"loose1.jpg", "loose2.jpg"})
			},
			wantOutput: "✅ Selected:",
		},
		{
			name:         "all files cached",
			categoryName: "",
			setupFunc: func(root string) error {
				if err := createTestFiles(root, map[string][]string{
					"shirts": {"shirt1.jpg"},
				}, []string{"loose1.jpg"}); err != nil {
					return err
				}
				// Pre-cache all files
				cache, _ := storage.NewManager(root)
				cache.Add("shirt1.jpg", filepath.Join(root, "shirts"))
				cache.Add("loose1.jpg", "UNCATEGORIZED")
				return nil
			},
			wantOutput: "No outfits available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()

			if err := tt.setupFunc(root); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			var buf bytes.Buffer
			err := QuickModeRandom(root, tt.categoryName, &buf)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			output := buf.String()
			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Expected output to contain %q, got %q", tt.wantOutput, output)
			}
		})
	}
}

func TestBuildQuickFilePool(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(string) (*storage.Manager, []string, []string, error)
		expectedCount int
		expectedFiles []string
	}{
		{
			name: "mixed categories and uncategorized",
			setupFunc: func(root string) (*storage.Manager, []string, []string, error) {
				err := createTestFiles(root, map[string][]string{
					"shirts": {"shirt1.jpg", "shirt2.jpg"},
					"pants":  {"pants1.jpg"},
				}, []string{"loose1.jpg", "loose2.jpg"})
				if err != nil {
					return nil, nil, nil, err
				}

				cache, err := storage.NewManager(root)
				categories := []string{
					filepath.Join(root, "shirts"),
					filepath.Join(root, "pants"),
				}
				uncategorized := []string{
					filepath.Join(root, "loose1.jpg"),
					filepath.Join(root, "loose2.jpg"),
				}
				return cache, categories, uncategorized, err
			},
			expectedCount: 5,
			expectedFiles: []string{"shirt1.jpg", "shirt2.jpg", "pants1.jpg", "loose1.jpg", "loose2.jpg"},
		},
		{
			name: "only categories",
			setupFunc: func(root string) (*storage.Manager, []string, []string, error) {
				err := createTestFiles(root, map[string][]string{
					"shirts": {"shirt1.jpg"},
				}, nil)
				if err != nil {
					return nil, nil, nil, err
				}

				cache, err := storage.NewManager(root)
				categories := []string{filepath.Join(root, "shirts")}
				return cache, categories, nil, err
			},
			expectedCount: 1,
			expectedFiles: []string{"shirt1.jpg"},
		},
		{
			name: "only uncategorized",
			setupFunc: func(root string) (*storage.Manager, []string, []string, error) {
				err := createTestFiles(root, nil, []string{"loose1.jpg"})
				if err != nil {
					return nil, nil, nil, err
				}

				cache, err := storage.NewManager(root)
				uncategorized := []string{filepath.Join(root, "loose1.jpg")}
				return cache, nil, uncategorized, err
			},
			expectedCount: 1,
			expectedFiles: []string{"loose1.jpg"},
		},
		{
			name: "with cached files",
			setupFunc: func(root string) (*storage.Manager, []string, []string, error) {
				err := createTestFiles(root, map[string][]string{
					"shirts": {"shirt1.jpg", "shirt2.jpg"},
				}, []string{"loose1.jpg"})
				if err != nil {
					return nil, nil, nil, err
				}

				cache, err := storage.NewManager(root)
				if err != nil {
					return nil, nil, nil, err
				}

				// Cache one file
				cache.Add("shirt1.jpg", filepath.Join(root, "shirts"))

				categories := []string{filepath.Join(root, "shirts")}
				uncategorized := []string{filepath.Join(root, "loose1.jpg")}
				return cache, categories, uncategorized, err
			},
			expectedCount: 2,
			expectedFiles: []string{"shirt2.jpg", "loose1.jpg"},
		},
		{
			name: "hidden files ignored",
			setupFunc: func(root string) (*storage.Manager, []string, []string, error) {
				err := createTestFiles(root, map[string][]string{
					"shirts": {"shirt1.jpg", ".hidden.jpg"},
				}, nil)
				if err != nil {
					return nil, nil, nil, err
				}

				cache, err := storage.NewManager(root)
				categories := []string{filepath.Join(root, "shirts")}
				return cache, categories, nil, err
			},
			expectedCount: 1,
			expectedFiles: []string{"shirt1.jpg"},
		},
		{
			name: "empty pool",
			setupFunc: func(root string) (*storage.Manager, []string, []string, error) {
				cache, err := storage.NewManager(root)
				return cache, nil, nil, err
			},
			expectedCount: 0,
			expectedFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			cache, categories, uncategorized, err := tt.setupFunc(root)
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			pool := buildQuickFilePool(categories, uncategorized, cache)

			if len(pool) != tt.expectedCount {
				t.Errorf("Expected %d files in pool, got %d", tt.expectedCount, len(pool))
			}

			// Check that expected files are present
			poolFiles := make(map[string]bool)
			for _, entry := range pool {
				poolFiles[entry.FileName] = true
			}

			for _, expectedFile := range tt.expectedFiles {
				if !poolFiles[expectedFile] {
					t.Errorf("Expected file %q not found in pool", expectedFile)
				}
			}
		})
	}
}

func TestQuickModeRandomInvalidRoot(t *testing.T) {
	var buf bytes.Buffer
	err := QuickModeRandom("/nonexistent/path", "", &buf)
	if err == nil {
		t.Error("Expected error for invalid root path")
	}
}

func TestQuickModeRandomCacheError(t *testing.T) {
	// Test with completely invalid root path that can't be made absolute
	var buf bytes.Buffer
	err := QuickModeRandom("\x00invalid", "", &buf)
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

// Helper function to create test files
func createTestFiles(root string, categories map[string][]string, uncategorized []string) error {
	// Create category directories and files
	for category, files := range categories {
		categoryPath := filepath.Join(root, category)
		if err := os.MkdirAll(categoryPath, 0755); err != nil {
			return err
		}

		for _, file := range files {
			filePath := filepath.Join(categoryPath, file)
			if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
				return err
			}
		}
	}

	// Create uncategorized files
	for _, file := range uncategorized {
		filePath := filepath.Join(root, file)
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			return err
		}
	}

	return nil
}
