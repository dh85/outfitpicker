package test

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/app"
	"github.com/dh85/outfitpicker/internal/storage"
)

func TestStressLargeNumberOfCategories(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	tempDir := t.TempDir()
	
	// Create 100 categories with 50 files each
	const numCategories = 100
	const filesPerCategory = 50
	
	for i := 0; i < numCategories; i++ {
		catDir := filepath.Join(tempDir, fmt.Sprintf("category_%03d", i))
		os.MkdirAll(catDir, 0755)
		
		for j := 0; j < filesPerCategory; j++ {
			filename := fmt.Sprintf("file_%03d.jpg", j)
			os.WriteFile(filepath.Join(catDir, filename), []byte("test"), 0644)
		}
	}
	
	// Test that the app can handle this many categories
	cache, err := storage.NewManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	
	var stdout bytes.Buffer
	stdin := strings.NewReader("q\n")
	
	err = app.Run(tempDir, "", bufio.NewReader(stdin), &stdout)
	if err != nil {
		t.Errorf("app should handle large number of categories: %v", err)
	}
	
	output := stdout.String()
	if !strings.Contains(output, "100 categories") {
		t.Errorf("should display correct category count, got: %s", output)
	}
}

func TestStressLargeFilenames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	tempDir := t.TempDir()
	catDir := filepath.Join(tempDir, "stress")
	os.MkdirAll(catDir, 0755)
	
	// Create files with very long names (but within filesystem limits)
	longName := strings.Repeat("a", 200) + ".jpg"
	unicodeName := strings.Repeat("🎉", 50) + ".jpg"
	specialName := "file with spaces & special chars (123).jpg"
	
	files := []string{longName, unicodeName, specialName}
	for _, filename := range files {
		os.WriteFile(filepath.Join(catDir, filename), []byte("test"), 0644)
	}
	
	cache, err := storage.NewManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	
	var stdout bytes.Buffer
	stdin := strings.NewReader("1\nu\nq\n")
	
	err = app.Run(tempDir, "", bufio.NewReader(stdin), &stdout)
	if err != nil {
		t.Errorf("app should handle long filenames: %v", err)
	}
	
	output := stdout.String()
	for _, filename := range files {
		if !strings.Contains(output, filename) {
			t.Errorf("should display filename %s in output", filename)
		}
	}
}

func TestStressDeepDirectoryStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	tempDir := t.TempDir()
	
	// Create nested directory structure (but only top level should be categories)
	for i := 0; i < 10; i++ {
		catDir := filepath.Join(tempDir, fmt.Sprintf("cat_%d", i))
		os.MkdirAll(catDir, 0755)
		
		// Create nested subdirectories (should be ignored)
		nestedDir := filepath.Join(catDir, "nested", "deep", "structure")
		os.MkdirAll(nestedDir, 0755)
		os.WriteFile(filepath.Join(nestedDir, "deep_file.jpg"), []byte("test"), 0644)
		
		// Create files at category level (should be included)
		os.WriteFile(filepath.Join(catDir, "top_level.jpg"), []byte("test"), 0644)
	}
	
	cache, err := storage.NewManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	
	var stdout bytes.Buffer
	stdin := strings.NewReader("1\nu\nq\n")
	
	err = app.Run(tempDir, "", bufio.NewReader(stdin), &stdout)
	if err != nil {
		t.Errorf("app should handle nested directories: %v", err)
	}
	
	output := stdout.String()
	if !strings.Contains(output, "top_level.jpg") {
		t.Error("should find top-level files")
	}
	if strings.Contains(output, "deep_file.jpg") {
		t.Error("should not include files from nested directories")
	}
}

func TestStressMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	tempDir := t.TempDir()
	cache, err := storage.NewManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	
	// Add many files to cache to test memory usage
	const numCategories = 50
	const filesPerCategory = 1000
	
	for i := 0; i < numCategories; i++ {
		categoryPath := fmt.Sprintf("/category_%d", i)
		for j := 0; j < filesPerCategory; j++ {
			filename := fmt.Sprintf("file_%d.jpg", j)
			cache.Add(filename, categoryPath)
		}
	}
	
	// Load and verify
	data := cache.Load()
	totalFiles := 0
	for _, files := range data {
		totalFiles += len(files)
	}
	
	expectedFiles := numCategories * filesPerCategory
	if totalFiles != expectedFiles {
		t.Errorf("expected %d files, got %d", expectedFiles, totalFiles)
	}
	
	// Save and reload to test persistence
	cache.Save(data)
	reloaded := cache.Load()
	
	reloadedTotal := 0
	for _, files := range reloaded {
		reloadedTotal += len(files)
	}
	
	if reloadedTotal != expectedFiles {
		t.Errorf("expected %d files after reload, got %d", expectedFiles, reloadedTotal)
	}
}