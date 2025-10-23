package ui

import (
	"bytes"
	"strings"
	"testing"
)

// Integration tests for UI components working together

func TestUIWorkflow(t *testing.T) {
	var buf bytes.Buffer
	theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
	ui := NewUI(&buf, theme)

	// Simulate a complete UI workflow
	categories := []string{"/path/to/Beach", "/path/to/Formal", "/path/to/Casual"}

	// 1. Show main menu
	ui.MainMenu(categories, nil)
	output1 := buf.String()

	// Verify main menu contains expected elements
	expectedMainMenu := []string{
		"Outfit Picker", "Outfit Folders", "1", "Beach", "2", "Formal", "3", "Casual",
		"What would you like to do?", "r", "s", "u", "q",
	}

	for _, expected := range expectedMainMenu {
		if !strings.Contains(output1, expected) {
			t.Errorf("main menu missing %q", expected)
		}
	}

	buf.Reset()

	// 2. Show category info
	ui.CategoryInfo("Beach", 10, 3)
	output2 := buf.String()

	if !strings.Contains(output2, "Beach") || !strings.Contains(output2, "10") || !strings.Contains(output2, "3") {
		t.Error("category info not displayed correctly")
	}

	buf.Reset()

	// 3. Show category menu
	ui.Menu()
	output3 := buf.String()

	expectedMenu := []string{"What would you like to do?", "r", "s", "u", "q"}
	for _, expected := range expectedMenu {
		if !strings.Contains(output3, expected) {
			t.Errorf("category menu missing %q", expected)
		}
	}

	buf.Reset()

	// 4. Show random selection
	ui.RandomSelection("beach-outfit-1.jpg")
	output4 := buf.String()

	if !strings.Contains(output4, "I picked this outfit for you") || !strings.Contains(output4, "beach-outfit-1.jpg") {
		t.Error("random selection not displayed correctly")
	}

	buf.Reset()

	// 5. Show keep action
	ui.KeepAction("beach-outfit-1.jpg")
	output5 := buf.String()

	if !strings.Contains(output5, "Great choice! I've saved") || !strings.Contains(output5, "beach-outfit-1.jpg") {
		t.Error("keep action not displayed correctly")
	}

	buf.Reset()

	// 6. Show completion summary
	ui.CompletionSummary(2, 3, []string{"Beach", "Formal"})
	output6 := buf.String()

	if !strings.Contains(output6, "2/3") || !strings.Contains(output6, "Beach") || !strings.Contains(output6, "Formal") {
		t.Error("completion summary not displayed correctly")
	}
}

func TestUIThemeConsistency(t *testing.T) {
	themes := []Theme{
		{UseColors: true, UseEmojis: true, Compact: false},
		{UseColors: false, UseEmojis: true, Compact: false},
		{UseColors: true, UseEmojis: false, Compact: false},
		{UseColors: false, UseEmojis: false, Compact: false},
		{UseColors: true, UseEmojis: true, Compact: true},
		{UseColors: false, UseEmojis: false, Compact: true},
	}

	for i, theme := range themes {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			var buf bytes.Buffer
			ui := NewUI(&buf, theme)

			// Test various UI methods with this theme
			ui.Header("Test Header")
			ui.CategoryInfo("TestCategory", 5, 2)
			ui.Menu()
			ui.RandomSelection("test-file.jpg")
			ui.KeepAction("test-file.jpg")
			ui.Error("test error")
			ui.Success("test success")
			ui.Info("test info")
			ui.Warning("test warning")

			output := buf.String()

			// Verify output is not empty
			if len(output) == 0 {
				t.Error("no output generated")
			}

			// Verify emoji consistency
			hasEmojis := strings.ContainsAny(output, "📂📄✅❌ℹ️⚠️🎉🎲📋⚙️👋")
			if theme.UseEmojis && !hasEmojis {
				t.Error("expected emojis but none found")
			}
			if !theme.UseEmojis && hasEmojis {
				t.Error("unexpected emojis found")
			}

			// Verify color codes consistency
			hasColors := strings.Contains(output, "\033[")
			if theme.UseColors && !hasColors {
				t.Error("expected colors but none found")
			}
			if !theme.UseColors && hasColors {
				t.Error("unexpected colors found")
			}
		})
	}
}

func TestUIErrorHandling(t *testing.T) {
	var buf bytes.Buffer
	theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
	ui := NewUI(&buf, theme)

	// Test with empty/nil inputs
	ui.CategoryInfo("", 0, 0)
	ui.SelectedFiles("", []string{})
	ui.UnselectedFiles([]string{})
	ui.CompletionSummary(0, 0, []string{})
	ui.MainMenu([]string{}, nil)

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected some output even with empty inputs")
	}
}

func TestUILargeDatasets(t *testing.T) {
	var buf bytes.Buffer
	theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
	ui := NewUI(&buf, theme)

	// Test with large number of categories
	largeCategories := make([]string, 100)
	for i := 0; i < 100; i++ {
		largeCategories[i] = "/path/to/category" + string(rune('A'+i%26))
	}

	ui.MainMenu(largeCategories, nil)
	output1 := buf.String()

	if !strings.Contains(output1, "1") || !strings.Contains(output1, "100") {
		t.Error("large category list not handled correctly")
	}

	buf.Reset()

	// Test with large number of files
	largeFiles := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		largeFiles[i] = "file" + string(rune('0'+i%10)) + ".jpg"
	}

	ui.SelectedFiles("TestCategory", largeFiles)
	output2 := buf.String()

	if !strings.Contains(output2, "1000 outfits") {
		t.Error("large file list not handled correctly")
	}
}

func TestUIProgressBar(t *testing.T) {
	var buf bytes.Buffer
	theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
	ui := NewUI(&buf, theme)

	// Test progress bar at different percentages
	percentages := []float64{0, 25, 50, 75, 100}

	for _, pct := range percentages {
		totalFiles := 100
		selectedFiles := int(pct)

		buf.Reset()
		ui.CategoryInfo("TestCategory", totalFiles, selectedFiles)
		output := buf.String()

		if selectedFiles > 0 && !strings.Contains(output, "Progress:") {
			t.Errorf("expected progress bar for %v%% completion", pct)
		}

		if selectedFiles > 0 && !strings.Contains(output, string(rune(int('0')+int(pct/10)))) {
			t.Errorf("expected percentage %v in output", pct)
		}
	}
}

func TestUICompactMode(t *testing.T) {
	var buf1, buf2 bytes.Buffer

	// Full mode
	theme1 := Theme{UseColors: true, UseEmojis: true, Compact: false}
	ui1 := NewUI(&buf1, theme1)

	// Compact mode
	theme2 := Theme{UseColors: true, UseEmojis: true, Compact: true}
	ui2 := NewUI(&buf2, theme2)

	// Test same operations in both modes
	ui1.Header("Test Header")
	ui2.Header("Test Header")

	ui1.CategoryInfo("TestCategory", 10, 5)
	ui2.CategoryInfo("TestCategory", 10, 5)

	ui1.Menu()
	ui2.Menu()

	output1 := buf1.String()
	output2 := buf2.String()

	// Compact mode should produce less output
	if len(output2) >= len(output1) {
		t.Error("compact mode should produce shorter output")
	}

	// Both should contain essential information
	essentialInfo := []string{"Test Header", "TestCategory", "10", "5"}
	for _, info := range essentialInfo {
		if !strings.Contains(output1, info) {
			t.Errorf("full mode missing essential info: %s", info)
		}
		if !strings.Contains(output2, info) {
			t.Errorf("compact mode missing essential info: %s", info)
		}
	}
}

// Benchmark integration tests
func BenchmarkUIWorkflow(b *testing.B) {
	theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
	categories := []string{"/path/to/Beach", "/path/to/Formal", "/path/to/Casual"}
	files := []string{"file1.jpg", "file2.jpg", "file3.jpg"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		ui := NewUI(&buf, theme)

		ui.MainMenu(categories, nil)
		ui.CategoryInfo("Beach", 10, 3)
		ui.Menu()
		ui.SelectedFiles("Beach", files)
		ui.RandomSelection("beach-outfit.jpg")
		ui.KeepAction("beach-outfit.jpg")
		ui.CompletionSummary(1, 3, []string{"Beach"})
	}
}

func BenchmarkUILargeDataset(b *testing.B) {
	theme := Theme{UseColors: false, UseEmojis: false, Compact: true}

	// Create large datasets
	categories := make([]string, 50)
	for i := 0; i < 50; i++ {
		categories[i] = "/path/to/category" + string(rune('A'+i%26))
	}

	files := make([]string, 500)
	for i := 0; i < 500; i++ {
		files[i] = "file" + string(rune('0'+i%10)) + ".jpg"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		ui := NewUI(&buf, theme)

		ui.MainMenu(categories, nil)
		ui.SelectedFiles("TestCategory", files)
	}
}
