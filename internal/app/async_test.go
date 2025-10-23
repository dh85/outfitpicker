package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAsyncOperations_LoadCategoriesAsync(t *testing.T) {
	ctx := context.Background()
	ao := NewAsyncOperations(ctx)
	defer ao.Cancel()
	
	// Create test structure
	tempDir := t.TempDir()
	os.MkdirAll(filepath.Join(tempDir, "Category1"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "Category2"), 0755)
	
	// Test async loading
	done := make(chan bool)
	var categories []string
	var err error
	
	ao.LoadCategoriesAsync(tempDir, func(cats []string, e error) {
		categories = cats
		err = e
		done <- true
	})
	
	select {
	case <-done:
		if err != nil {
			t.Errorf("async load failed: %v", err)
		}
		if len(categories) != 2 {
			t.Errorf("expected 2 categories, got %d", len(categories))
		}
	case <-time.After(time.Second):
		t.Error("async operation timed out")
	}
	
	ao.Wait()
}

func TestAsyncOperations_Cancellation(t *testing.T) {
	ctx := context.Background()
	ao := NewAsyncOperations(ctx)
	
	// Cancel immediately
	ao.Cancel()
	
	done := make(chan bool)
	var err error
	
	ao.LoadCategoriesAsync("/some/path", func(cats []string, e error) {
		err = e
		done <- true
	})
	
	select {
	case <-done:
		if err == nil {
			t.Error("expected cancellation error")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("cancellation didn't work")
	}
}

func TestAsyncOperations_PreloadCache(t *testing.T) {
	ctx := context.Background()
	ao := NewAsyncOperations(ctx)
	defer ao.Cancel()
	
	tempDir := t.TempDir()
	categories := []string{
		filepath.Join(tempDir, "cat1"),
		filepath.Join(tempDir, "cat2"),
	}
	
	// Create categories with files
	for _, cat := range categories {
		os.MkdirAll(cat, 0755)
		os.WriteFile(filepath.Join(cat, "test.jpg"), []byte("test"), 0644)
	}
	
	optimizer := NewCacheOptimizer(time.Minute)
	
	done := make(chan bool)
	var err error
	
	ao.PreloadCacheAsync(categories, optimizer, func(e error) {
		err = e
		done <- true
	})
	
	select {
	case <-done:
		if err != nil {
			t.Errorf("preload failed: %v", err)
		}
	case <-time.After(time.Second):
		t.Error("preload timed out")
	}
	
	ao.Wait()
}