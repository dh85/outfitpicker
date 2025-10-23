package app

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/testutil"
)

func TestUnicodeFilenames(t *testing.T) {
	f := testutil.NewTestFixture(t)

	// Create files with Unicode names
	unicodeFiles := []string{
		"файл.jpg",   // Russian
		"文件.jpg",     // Chinese
		"🎉party.jpg", // Emoji
	}

	catPath := f.CreateCategory("unicode", unicodeFiles...)

	var stdout bytes.Buffer
	pr := &prompter{r: bufio.NewReader(strings.NewReader("u\nq\n"))}

	err := runCategoryFlow(catPath, f.Cache, pr, &stdout)
	if err != nil {
		t.Errorf("should handle Unicode filenames: %v", err)
	}

	output := stdout.String()
	for _, file := range unicodeFiles {
		if !strings.Contains(output, file) {
			t.Errorf("Unicode filename %s not displayed correctly", file)
		}
	}
}

func TestVeryLargeCategory(t *testing.T) {
	f := testutil.NewTestFixture(t)

	// Create category with many files
	files := make([]string, 100) // Reduced for faster testing
	for i := 0; i < 100; i++ {
		files[i] = fmt.Sprintf("file%03d.jpg", i)
	}

	catPath := f.CreateCategory("large", files...)

	var stdout bytes.Buffer
	pr := &prompter{r: bufio.NewReader(strings.NewReader("u\nq\n"))}

	err := runCategoryFlow(catPath, f.Cache, pr, &stdout)
	if err != nil {
		t.Errorf("should handle large categories: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Total files in \"large\": 100") && !strings.Contains(output, "100") {
		t.Errorf("should display correct file count for large category, got: %s", output)
	}
}

func TestListCategoriesWithSpecialDirectories(t *testing.T) {
	f := testutil.NewTestFixture(t)

	// Create various directory types
	os.MkdirAll(filepath.Join(f.TempDir, "Normal"), 0755)
	os.MkdirAll(filepath.Join(f.TempDir, ".hidden"), 0755)
	os.MkdirAll(filepath.Join(f.TempDir, "Downloads"), 0755)
	os.MkdirAll(filepath.Join(f.TempDir, "downloads"), 0755) // Different case
	os.WriteFile(filepath.Join(f.TempDir, "file.txt"), []byte("test"), 0644)

	categories, err := listCategories(f.TempDir)
	if err != nil {
		t.Fatalf("listCategories failed: %v", err)
	}

	// Should only include Normal (both Downloads and downloads are excluded case-insensitively)
	expectedCount := 1
	if len(categories) != expectedCount {
		t.Errorf("expected %d categories, got %d: %v", expectedCount, len(categories), categories)
		// Let's see what we actually got
		for _, cat := range categories {
			t.Logf("Found category: %s", filepath.Base(cat))
		}
	}

	// Check that hidden and Downloads are excluded
	for _, cat := range categories {
		base := filepath.Base(cat)
		if strings.HasPrefix(base, ".") {
			t.Error("hidden directory should be excluded")
		}
		if base == "Downloads" {
			t.Error("Downloads directory should be excluded")
		}
	}
}

func TestRandomAcrossAllWithEmptyCategories(t *testing.T) {
	f := testutil.NewTestFixture(t)

	// Create mix of empty and non-empty categories
	cat1 := f.CreateCategory("empty1")
	cat2 := f.CreateCategory("nonempty", "file1.jpg")
	cat3 := f.CreateCategory("empty2")

	categories := []string{cat1, cat2, cat3}

	var stdout bytes.Buffer
	pr := &prompter{r: bufio.NewReader(strings.NewReader("k\n"))}

	err := randomAcrossAll(categories, nil, f.Cache, pr, &stdout)
	if err != nil {
		t.Errorf("should handle mix of empty/non-empty categories: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "file1.jpg") {
		t.Error("should select from non-empty category")
	}
}
