package app

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/storage"
	"github.com/dh85/outfitpicker/internal/testutil"
)

func FuzzCategoryFlow(f *testing.F) {
	// Add seed inputs
	f.Add("r\nk\n")
	f.Add("s\n")
	f.Add("u\n")
	f.Add("q\n")
	f.Add("invalid\n")

	f.Fuzz(func(t *testing.T, input string) {
		fixture := testutil.NewTestFixture(t)
		catPath := fixture.CreateCategory("fuzz", "test1.jpg", "test2.jpg")

		var stdout bytes.Buffer
		pr := &prompter{r: bufio.NewReader(strings.NewReader(input))}

		// Should not panic
		_ = runCategoryFlow(catPath, fixture.Cache, pr, &stdout)
	})
}

func FuzzListCategories(f *testing.F) {
	// Add seed inputs
	f.Add("normal")
	f.Add(".hidden")
	f.Add("Downloads")
	f.Add("downloads")
	f.Add("very-long-category-name-with-special-chars-123")

	f.Fuzz(func(t *testing.T, categoryName string) {
		if len(categoryName) > 100 || strings.Contains(categoryName, "/") {
			return // Skip invalid inputs
		}

		tempDir := t.TempDir()
		catPath := filepath.Join(tempDir, categoryName)
		_ = os.MkdirAll(catPath, 0755)

		// Should not panic
		categories, _ := listCategories(tempDir)
		_ = categories
	})
}

func FuzzCacheOperations(f *testing.F) {
	// Add seed inputs
	f.Add("test.jpg", "/path/to/category")
	f.Add("файл.jpg", "/path/to/категория")
	f.Add("🎉.jpg", "/path/to/📂")

	f.Fuzz(func(t *testing.T, filename, categoryPath string) {
		if len(filename) > 255 || len(categoryPath) > 1000 {
			return // Skip unreasonably long inputs
		}

		tempDir := t.TempDir()
		cache, err := storage.NewManager(tempDir)
		if err != nil {
			return
		}

		// Should not panic
		cache.Add(filename, categoryPath)
		cache.Load()
		cache.Clear(categoryPath)
	})
}
