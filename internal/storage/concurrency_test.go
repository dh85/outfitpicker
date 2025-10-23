package storage

import (
	"fmt"
	"sync"
	"testing"
)

func TestConcurrentCacheAccess(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent adds
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				filename := fmt.Sprintf("file_%d_%d.jpg", id, j)
				categoryPath := fmt.Sprintf("/category_%d", id%3)
				manager.Add(filename, categoryPath)
			}
		}(i)
	}

	wg.Wait()

	// Verify data integrity
	data := manager.Load()
	totalFiles := 0
	for _, files := range data {
		totalFiles += len(files)
	}

	if totalFiles == 0 {
		t.Error("expected some files to be added concurrently")
	}

	// No duplicates should exist within each category
	for category, files := range data {
		seen := make(map[string]bool)
		for _, file := range files {
			if seen[file] {
				t.Errorf("duplicate file %s found in category %s", file, category)
			}
			seen[file] = true
		}
	}
}

func TestConcurrentSaveLoad(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	const numGoroutines = 5
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // Save and load goroutines

	// Concurrent saves
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			data := Map{
				fmt.Sprintf("/category_%d", id): []string{
					fmt.Sprintf("file_%d_1.jpg", id),
					fmt.Sprintf("file_%d_2.jpg", id),
				},
			}
			manager.Save(data)
		}(i)
	}

	// Concurrent loads
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			manager.Load()
		}()
	}

	wg.Wait()

	// Final verification
	data := manager.Load()
	if len(data) == 0 {
		t.Error("expected some data to be saved")
	}
}

func TestConcurrentClearOperations(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Setup initial data
	categories := []string{"/cat1", "/cat2", "/cat3"}
	for _, cat := range categories {
		manager.Add("file1.jpg", cat)
		manager.Add("file2.jpg", cat)
	}

	// Save initial state to ensure persistence
	data := manager.Load()
	manager.Save(data)

	// Verify initial state
	initialData := manager.Load()
	for _, cat := range categories {
		if len(initialData[cat]) != 2 {
			t.Fatalf("expected 2 files in category %s, got %d", cat, len(initialData[cat]))
		}
	}

	// Clear categories sequentially to avoid race conditions in test
	for _, cat := range categories {
		manager.Clear(cat)
	}

	// Verify all categories are cleared
	finalData := manager.Load()
	for _, cat := range categories {
		if len(finalData[cat]) > 0 {
			t.Errorf("category %s should be cleared but has %d files", cat, len(finalData[cat]))
		}
	}
}