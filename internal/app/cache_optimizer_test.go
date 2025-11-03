package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheOptimizer_GetFileCount(t *testing.T) {
	optimizer := NewCacheOptimizer(time.Minute)
	tempDir := t.TempDir()

	// Create test category
	catPath := filepath.Join(tempDir, "test")
	_ = os.MkdirAll(catPath, 0755)
	_ = os.WriteFile(filepath.Join(catPath, "file1.jpg"), []byte("test"), 0644)
	_ = os.WriteFile(filepath.Join(catPath, "file2.jpg"), []byte("test"), 0644)

	// First call should compute
	count1, err := optimizer.GetFileCount(catPath)
	if err != nil {
		t.Errorf("GetFileCount failed: %v", err)
	}
	if count1 != 2 {
		t.Errorf("expected 2 files, got %d", count1)
	}

	// Second call should use cache
	count2, err := optimizer.GetFileCount(catPath)
	if err != nil {
		t.Errorf("GetFileCount failed: %v", err)
	}
	if count2 != count1 {
		t.Error("cache not working correctly")
	}
}

func TestCacheOptimizer_TTL(t *testing.T) {
	optimizer := NewCacheOptimizer(time.Millisecond) // Very short TTL
	tempDir := t.TempDir()

	catPath := filepath.Join(tempDir, "test")
	_ = os.MkdirAll(catPath, 0755)
	_ = os.WriteFile(filepath.Join(catPath, "file1.jpg"), []byte("test"), 0644)

	// Get initial count
	count1, _ := optimizer.GetFileCount(catPath)

	// Wait for TTL to expire
	time.Sleep(2 * time.Millisecond)

	// Add another file
	_ = os.WriteFile(filepath.Join(catPath, "file2.jpg"), []byte("test"), 0644)

	// Should recompute due to expired TTL
	count2, _ := optimizer.GetFileCount(catPath)
	if count2 != count1+1 {
		t.Error("TTL expiration not working correctly")
	}
}

func TestCacheOptimizer_Clear(t *testing.T) {
	optimizer := NewCacheOptimizer(time.Minute)
	tempDir := t.TempDir()

	catPath := filepath.Join(tempDir, "test")
	_ = os.MkdirAll(catPath, 0755)
	_ = os.WriteFile(filepath.Join(catPath, "file1.jpg"), []byte("test"), 0644)

	// Populate cache
	_, _ = optimizer.GetFileCount(catPath)

	// Clear cache
	optimizer.Clear()

	// Add file and check - should recompute
	_ = os.WriteFile(filepath.Join(catPath, "file2.jpg"), []byte("test"), 0644)
	count, _ := optimizer.GetFileCount(catPath)
	if count != 2 {
		t.Error("cache clear not working correctly")
	}
}

func TestCacheOptimizer_ConcurrentAccess(t *testing.T) {
	optimizer := NewCacheOptimizer(time.Minute)
	tempDir := t.TempDir()

	catPath := filepath.Join(tempDir, "test")
	_ = os.MkdirAll(catPath, 0755)
	_ = os.WriteFile(filepath.Join(catPath, "file1.jpg"), []byte("test"), 0644)

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := optimizer.GetFileCount(catPath)
			if err != nil {
				t.Errorf("concurrent access failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
