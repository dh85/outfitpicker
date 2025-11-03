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

func TestRandomAcrossAll_SkipAndContinue(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	catDir := filepath.Join(tempDir, "casual")
	os.MkdirAll(catDir, 0755)
	os.WriteFile(filepath.Join(catDir, "outfit1.jpg"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(catDir, "outfit2.jpg"), []byte("test"), 0644)

	cache, _ := storage.NewManager(tempDir)
	categories := []string{catDir}
	var uncategorized []string

	// Simulate: skip first, keep second
	input := "s\nk\n"
	stdout := &bytes.Buffer{}
	pr := &prompter{r: bufio.NewReader(strings.NewReader(input)), w: stdout}

	err := randomAcrossAll(categories, uncategorized, cache, pr, stdout)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Skipped") {
		t.Error("Expected skip message in output")
	}
	if !strings.Contains(output, "Great choice!") {
		t.Error("Expected keep message in output")
	}
}

func TestRandomAcrossAll_SkipAllThenRetry(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	catDir := filepath.Join(tempDir, "casual")
	os.MkdirAll(catDir, 0755)
	os.WriteFile(filepath.Join(catDir, "outfit1.jpg"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(catDir, "outfit2.jpg"), []byte("test"), 0644)

	cache, _ := storage.NewManager(tempDir)
	categories := []string{catDir}
	var uncategorized []string

	// Simulate: skip all, then retry and keep one
	input := "s\ns\ny\nk\n"
	stdout := &bytes.Buffer{}
	pr := &prompter{r: bufio.NewReader(strings.NewReader(input)), w: stdout}

	err := randomAcrossAll(categories, uncategorized, cache, pr, stdout)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "You've skipped all available outfits") {
		t.Error("Expected skip exhaustion message")
	}
	if !strings.Contains(output, "Try again with the same outfits?") {
		t.Error("Expected retry prompt")
	}
	if !strings.Contains(output, "Great choice!") {
		t.Error("Expected keep message after retry")
	}
}

func TestRandomAcrossAll_SkipAllThenDecline(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	catDir := filepath.Join(tempDir, "casual")
	os.MkdirAll(catDir, 0755)
	os.WriteFile(filepath.Join(catDir, "outfit1.jpg"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(catDir, "outfit2.jpg"), []byte("test"), 0644)

	cache, _ := storage.NewManager(tempDir)
	categories := []string{catDir}
	var uncategorized []string

	// Simulate: skip all, then decline retry
	input := "s\ns\nn\n"
	stdout := &bytes.Buffer{}
	pr := &prompter{r: bufio.NewReader(strings.NewReader(input)), w: stdout}

	err := randomAcrossAll(categories, uncategorized, cache, pr, stdout)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "You've skipped all available outfits") {
		t.Error("Expected skip exhaustion message")
	}
	if !strings.Contains(output, "Try again with the same outfits?") {
		t.Error("Expected retry prompt")
	}
}

func TestCategoryManager_HandleRandomSelection_SkipAndContinue(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tempDir, "outfit1.jpg"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "outfit2.jpg"), []byte("test"), 0644)

	cache, _ := storage.NewManager(filepath.Dir(tempDir))
	stdout := &bytes.Buffer{}
	cm := NewCategoryManager(cache, stdout)

	category := &Category{
		Path:  tempDir,
		Files: []string{filepath.Join(tempDir, "outfit1.jpg"), filepath.Join(tempDir, "outfit2.jpg")},
	}

	// Simulate: skip first, keep second
	input := "s\nk\n"
	pr := &prompter{r: bufio.NewReader(strings.NewReader(input)), w: stdout}

	err := cm.handleRandomSelection(category, pr)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Skipped") {
		t.Error("Expected skip message in output")
	}
	if !strings.Contains(output, "Great choice!") {
		t.Error("Expected keep message in output")
	}
}

func TestCategoryManager_HandleRandomSelection_SkipAllThenRetry(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tempDir, "outfit1.jpg"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, "outfit2.jpg"), []byte("test"), 0644)

	cache, _ := storage.NewManager(filepath.Dir(tempDir))
	stdout := &bytes.Buffer{}
	cm := NewCategoryManager(cache, stdout)

	category := &Category{
		Path:  tempDir,
		Files: []string{filepath.Join(tempDir, "outfit1.jpg"), filepath.Join(tempDir, "outfit2.jpg")},
	}

	// Simulate: skip all, then retry and keep one
	input := "s\ns\ny\nk\n"
	pr := &prompter{r: bufio.NewReader(strings.NewReader(input)), w: stdout}

	err := cm.handleRandomSelection(category, pr)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "You've skipped all available outfits in this category") {
		t.Error("Expected category skip exhaustion message")
	}
	if !strings.Contains(output, "Try again with the same outfits?") {
		t.Error("Expected retry prompt")
	}
	if !strings.Contains(output, "Great choice!") {
		t.Error("Expected keep message after retry")
	}
}

func TestSkipTracking_SessionScoped(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	catDir := filepath.Join(tempDir, "casual")
	os.MkdirAll(catDir, 0755)
	os.WriteFile(filepath.Join(catDir, "outfit1.jpg"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(catDir, "outfit2.jpg"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(catDir, "outfit3.jpg"), []byte("test"), 0644)

	cache, _ := storage.NewManager(tempDir)
	categories := []string{catDir}
	var uncategorized []string

	// First session: skip two, then retry and skip one more (no keep)
	input1 := "s\ns\ny\ns\nn\n"
	stdout1 := &bytes.Buffer{}
	pr1 := &prompter{r: bufio.NewReader(strings.NewReader(input1)), w: stdout1}
	err1 := randomAcrossAll(categories, uncategorized, cache, pr1, stdout1)
	if err1 != nil {
		t.Fatalf("First session failed: %v", err1)
	}

	// Second session: should see all outfits again (skip tracking reset)
	input2 := "s\nk\n"
	stdout2 := &bytes.Buffer{}
	pr2 := &prompter{r: bufio.NewReader(strings.NewReader(input2)), w: stdout2}
	err2 := randomAcrossAll(categories, uncategorized, cache, pr2, stdout2)
	if err2 != nil {
		t.Fatalf("Second session failed: %v", err2)
	}

	// Should be able to skip and keep in second session
	output2 := stdout2.String()
	if !strings.Contains(output2, "Skipped") {
		t.Error("Expected skip in second session")
	}
	if !strings.Contains(output2, "Great choice!") {
		t.Error("Expected keep in second session")
	}
}

func TestCachePreservation_AfterSkipExhaustion(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	catDir := filepath.Join(tempDir, "casual")
	os.MkdirAll(catDir, 0755)
	os.WriteFile(filepath.Join(catDir, "outfit1.jpg"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(catDir, "outfit2.jpg"), []byte("test"), 0644)

	cache, _ := storage.NewManager(tempDir)
	categories := []string{catDir}
	var uncategorized []string

	// First keep one outfit
	input1 := "k\n"
	pr1 := &prompter{r: bufio.NewReader(strings.NewReader(input1)), w: &bytes.Buffer{}}
	randomAcrossAll(categories, uncategorized, cache, pr1, &bytes.Buffer{})

	// Verify one outfit is cached
	cached := cache.Load()
	if len(cached[catDir]) != 1 {
		t.Errorf("Expected 1 cached outfit, got %d", len(cached[catDir]))
	}

	// Skip remaining outfit and decline retry
	input2 := "s\nn\n"
	pr2 := &prompter{r: bufio.NewReader(strings.NewReader(input2)), w: &bytes.Buffer{}}
	randomAcrossAll(categories, uncategorized, cache, pr2, &bytes.Buffer{})

	// Verify cache is still preserved
	cached = cache.Load()
	if len(cached[catDir]) != 1 {
		t.Errorf("Expected cache preserved after skip exhaustion, got %d", len(cached[catDir]))
	}
}
