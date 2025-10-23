package app

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/storage"
	"github.com/dh85/outfitpicker/internal/testutil"
)

func TestCategoryManager_EnhancedUI(t *testing.T) {
	fixture := testutil.NewTestFixture(t)

	// Create test category with files
	beachPath := fixture.CreateCategory("Beach", "outfit1.jpg", "outfit2.jpg", "outfit3.jpg")

	cache := fixture.Cache

	var buf bytes.Buffer
	cm := NewCategoryManager(cache, &buf)

	// Test category info display
	category := &Category{
		Path:  beachPath,
		Files: []string{"outfit1.jpg", "outfit2.jpg", "outfit3.jpg"},
	}

	cm.displayCategoryInfo(category)
	output := buf.String()

	// Should contain enhanced UI elements
	if !strings.Contains(output, "Beach") {
		t.Error("expected category name in output")
	}
	if !strings.Contains(output, "3") {
		t.Error("expected file count in output")
	}

	buf.Reset()

	// Test menu display
	cm.displayMenu()
	menuOutput := buf.String()

	expectedMenuItems := []string{"r", "s", "u", "q"}
	for _, item := range expectedMenuItems {
		if !strings.Contains(menuOutput, item) {
			t.Errorf("expected menu item %q in output", item)
		}
	}

	buf.Reset()

	// Test selected files display
	cache.Add("outfit1.jpg", beachPath)
	cm.showSelectedFiles(beachPath)
	selectedOutput := buf.String()

	if !strings.Contains(selectedOutput, "outfit1.jpg") {
		t.Error("expected selected file in output")
	}

	buf.Reset()

	// Test unselected files display
	cm.showUnselectedFiles(category)
	unselectedOutput := buf.String()

	if !strings.Contains(unselectedOutput, "outfit2.jpg") || !strings.Contains(unselectedOutput, "outfit3.jpg") {
		t.Error("expected unselected files in output")
	}
}

func TestCategoryManager_CompletionSummary(t *testing.T) {
	fixture := testutil.NewTestFixture(t)

	beachPath := fixture.CreateCategory("Beach", "outfit1.jpg", "outfit2.jpg")
	fixture.CreateCategory("Formal", "suit1.jpg")

	cache := fixture.Cache

	var buf bytes.Buffer
	cm := NewCategoryManager(cache, &buf)

	// Mark Beach category as complete
	cache.Add("outfit1.jpg", beachPath)
	cache.Add("outfit2.jpg", beachPath)

	// Test completion summary
	cm.displayCompletionSummary(beachPath)
	output := buf.String()

	if !strings.Contains(output, "1/2") {
		t.Error("expected completion ratio in output")
	}
	if !strings.Contains(output, "Beach") {
		t.Error("expected completed category name in output")
	}
}

func TestCategoryManager_KeepAction(t *testing.T) {
	fixture := testutil.NewTestFixture(t)

	beachPath := fixture.CreateCategory("Beach", "outfit1.jpg")

	cache := fixture.Cache

	var buf bytes.Buffer
	cm := NewCategoryManager(cache, &buf)

	file := FileEntry{
		CategoryPath: beachPath,
		FilePath:     beachPath + "/outfit1.jpg",
		FileName:     "outfit1.jpg",
	}

	err := cm.handleKeepAction(file)
	if err != nil {
		t.Fatalf("handleKeepAction failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "outfit1.jpg") {
		t.Error("expected filename in keep action output")
	}

	// Note: File may be cleared from cache if category is complete
	// This is expected behavior when all files in category are selected
}

func TestShowSelectedAcrossAll_EnhancedUI(t *testing.T) {
	fixture := testutil.NewTestFixture(t)

	beachPath := fixture.CreateCategory("Beach", "outfit1.jpg", "outfit2.jpg")
	formalPath := fixture.CreateCategory("Formal", "suit1.jpg")

	cache := fixture.Cache

	// Add some selections
	cache.Add("outfit1.jpg", beachPath)
	cache.Add("suit1.jpg", formalPath)

	var buf bytes.Buffer
	categories := []string{beachPath, formalPath}

	err := showSelectedAcrossAll(categories, nil, cache, &buf)
	if err != nil {
		t.Fatalf("showSelectedAcrossAll failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "outfit1.jpg") {
		t.Error("expected Beach selection in output")
	}
	if !strings.Contains(output, "suit1.jpg") {
		t.Error("expected Formal selection in output")
	}
}

func TestShowUnselectedAcrossAll_EnhancedUI(t *testing.T) {
	fixture := testutil.NewTestFixture(t)

	beachPath := fixture.CreateCategory("Beach", "outfit1.jpg", "outfit2.jpg")
	formalPath := fixture.CreateCategory("Formal", "suit1.jpg")

	cache := fixture.Cache

	// Select only one file
	cache.Add("outfit1.jpg", beachPath)

	var buf bytes.Buffer
	categories := []string{beachPath, formalPath}

	err := showUnselectedAcrossAll(categories, nil, cache, &buf)
	if err != nil {
		t.Fatalf("showUnselectedAcrossAll failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "outfit2.jpg") {
		t.Error("expected unselected Beach file in output")
	}
	if !strings.Contains(output, "suit1.jpg") {
		t.Error("expected unselected Formal file in output")
	}
}

func TestShouldUseColors_CategoryUI(t *testing.T) {
	tests := []struct {
		name     string
		term     string
		noColor  string
		expected bool
	}{
		{"normal terminal", "xterm-256color", "", true},
		{"dumb terminal", "dumb", "", false},
		{"empty term", "", "", false},
		{"no color set", "xterm", "1", false},
		{"no color empty", "xterm", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origTerm := os.Getenv("TERM")
			origNoColor := os.Getenv("NO_COLOR")
			defer func() {
				os.Setenv("TERM", origTerm)
				os.Setenv("NO_COLOR", origNoColor)
			}()

			os.Setenv("TERM", tt.term)
			os.Setenv("NO_COLOR", tt.noColor)

			result := shouldUseColors()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Benchmark tests for UI performance
func BenchmarkCategoryManager_DisplayCategoryInfo(b *testing.B) {
	tempDir := b.TempDir()
	cache, _ := storage.NewManager(tempDir)
	var buf bytes.Buffer
	cm := NewCategoryManager(cache, &buf)

	category := &Category{
		Path:  tempDir + "/Beach",
		Files: []string{"outfit1.jpg", "outfit2.jpg", "outfit3.jpg"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		cm.displayCategoryInfo(category)
	}
}

func BenchmarkCategoryManager_DisplayMenu(b *testing.B) {
	cache, _ := storage.NewManager("/tmp")
	var buf bytes.Buffer
	cm := NewCategoryManager(cache, &buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		cm.displayMenu()
	}
}
