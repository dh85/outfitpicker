package app

import (
	"math/rand"
	"testing"

	"github.com/dh85/outfitpicker/internal/testutil"
)

// Property-based tests to verify invariants

func TestPropertyCategoryFileCountInvariant(t *testing.T) {
	// Property: File count should always be non-negative and match actual files
	for i := 0; i < 100; i++ {
		f := testutil.NewTestFixture(t)
		
		// Generate random number of files (0-50)
		numFiles := rand.Intn(51)
		files := make([]string, numFiles)
		for j := 0; j < numFiles; j++ {
			files[j] = randomFilename()
		}
		
		catPath := f.CreateCategory("test", files...)
		
		count, err := categoryFileCount(catPath)
		if err != nil {
			t.Errorf("categoryFileCount should not error for valid category: %v", err)
			continue
		}
		
		if count < 0 {
			t.Errorf("file count should never be negative, got: %d", count)
			continue
		}
		
		// Allow for slight discrepancy due to duplicate filenames in random generation
		if count > numFiles || count < numFiles-5 {
			t.Errorf("file count should be close to expected files: expected ~%d, got %d", numFiles, count)
		}
	}
}

func TestPropertyCacheConsistency(t *testing.T) {
	// Property: Adding and loading should be consistent
	for i := 0; i < 50; i++ {
		f := testutil.NewTestFixture(t)
		catPath := f.CreateCategory("test", "file1.jpg", "file2.jpg")
		
		// Add random files
		numAdds := rand.Intn(10) + 1
		addedFiles := make(map[string]bool)
		
		for j := 0; j < numAdds; j++ {
			filename := randomFilename()
			f.Cache.Add(filename, catPath)
			addedFiles[filename] = true
		}
		
		// Load and verify
		loaded := f.Cache.Load()
		cachedFiles := loaded[catPath]
		
		// All added files should be in cache
		for file := range addedFiles {
			found := false
			for _, cached := range cachedFiles {
				if cached == file {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("added file %s not found in cache", file)
			}
		}
	}
}

func TestPropertyUtilityFunctions(t *testing.T) {
	// Property: Utility functions should maintain data integrity
	for i := 0; i < 100; i++ {
		// Generate random string slice
		size := rand.Intn(20)
		original := make([]string, size)
		for j := 0; j < size; j++ {
			original[j] = randomString(rand.Intn(20) + 1)
		}
		
		// Test baseNames
		paths := make([]string, len(original))
		for j, name := range original {
			paths[j] = "/path/to/" + name
		}
		bases := baseNames(paths)
		
		if len(bases) != len(original) {
			t.Errorf("baseNames should preserve length: expected %d, got %d", len(original), len(bases))
		}
		
		// Test toSet
		set := toSet(original)
		if len(set) > len(original) {
			t.Errorf("set should not have more elements than original: set=%d, original=%d", len(set), len(original))
		}
		
		// Test mapKeys
		keys := mapKeys(set)
		if len(keys) != len(set) {
			t.Errorf("mapKeys should preserve set size: expected %d, got %d", len(set), len(keys))
		}
	}
}

func randomFilename() string {
	extensions := []string{".jpg", ".png", ".gif", ".pdf", ".txt"}
	name := randomString(rand.Intn(15) + 1)
	ext := extensions[rand.Intn(len(extensions))]
	return name + ext
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}