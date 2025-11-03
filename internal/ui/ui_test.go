package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewUI(t *testing.T) {
	var buf bytes.Buffer
	theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
	ui := NewUI(&buf, theme)

	if ui.writer != &buf {
		t.Error("writer not set correctly")
	}
	if ui.theme != theme {
		t.Error("theme not set correctly")
	}
}

func TestHeader(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		theme    Theme
		contains []string
	}{
		{
			name:     "full header with colors and emojis",
			title:    "Test Title",
			theme:    Theme{UseColors: true, UseEmojis: true, Compact: false},
			contains: []string{"Test Title", "─"},
		},
		{
			name:     "compact header",
			title:    "Test Title",
			theme:    Theme{UseColors: false, UseEmojis: true, Compact: true},
			contains: []string{"📋", "Test Title"},
		},
		{
			name:     "no emojis",
			title:    "Test Title",
			theme:    Theme{UseColors: false, UseEmojis: false, Compact: true},
			contains: []string{"Test Title"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ui := NewUI(&buf, tt.theme)
			ui.Header(tt.title)

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestCategoryInfo(t *testing.T) {
	tests := []struct {
		name          string
		categoryName  string
		totalFiles    int
		selectedFiles int
		theme         Theme
		contains      []string
	}{
		{
			name:          "full display with progress",
			categoryName:  "Beach",
			totalFiles:    10,
			selectedFiles: 5,
			theme:         Theme{UseColors: true, UseEmojis: true, Compact: false},
			contains:      []string{"Beach", "10", "5", "50.0%"},
		},
		{
			name:          "compact display",
			categoryName:  "Latex",
			totalFiles:    8,
			selectedFiles: 3,
			theme:         Theme{UseColors: false, UseEmojis: true, Compact: true},
			contains:      []string{"📂", "Latex", "(3/8)"},
		},
		{
			name:          "no files selected",
			categoryName:  "General",
			totalFiles:    5,
			selectedFiles: 0,
			theme:         Theme{UseColors: false, UseEmojis: false, Compact: false},
			contains:      []string{"General", "5", "0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ui := NewUI(&buf, tt.theme)
			ui.CategoryInfo(tt.categoryName, tt.totalFiles, tt.selectedFiles)

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestMenu(t *testing.T) {
	tests := []struct {
		name     string
		theme    Theme
		contains []string
	}{
		{
			name:     "full menu",
			theme:    Theme{UseColors: true, UseEmojis: true, Compact: false},
			contains: []string{"What would you like to do?", "r", "s", "u", "q", "🎲", "✅", "📄", "👋"},
		},
		{
			name:     "compact menu",
			theme:    Theme{UseColors: false, UseEmojis: false, Compact: true},
			contains: []string{"[r]andom", "[s]elected", "[u]nselected", "[q]uit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ui := NewUI(&buf, tt.theme)
			ui.Menu()

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestMainMenu(t *testing.T) {
	categories := []string{"/path/to/Beach", "/path/to/Latex", "/path/to/General"}

	var buf bytes.Buffer
	theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
	ui := NewUI(&buf, theme)
	ui.MainMenu(categories, nil)

	output := buf.String()
	expected := []string{
		"Outfit Picker", "Outfit Folders", "1", "Beach", "2", "Latex", "3", "General",
		"What would you like to do?", "r", "s", "u", "q",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q, got %q", exp, output)
		}
	}
}

func TestSelectedFiles(t *testing.T) {
	tests := []struct {
		name         string
		categoryName string
		files        []string
		theme        Theme
		contains     []string
	}{
		{
			name:         "no files selected",
			categoryName: "Beach",
			files:        []string{},
			theme:        Theme{UseColors: true, UseEmojis: true, Compact: false},
			contains:     []string{"You haven't picked any outfits from here yet"},
		},
		{
			name:         "files selected full display",
			categoryName: "Beach",
			files:        []string{"outfit1.jpg", "outfit2.jpg"},
			theme:        Theme{UseColors: true, UseEmojis: true, Compact: false},
			contains:     []string{"Outfits You've Already Picked", "outfit1.jpg", "outfit2.jpg", "2 outfits"},
		},
		{
			name:         "files selected compact",
			categoryName: "Beach",
			files:        []string{"outfit1.jpg"},
			theme:        Theme{UseColors: false, UseEmojis: false, Compact: true},
			contains:     []string{"You picked (1):", "1. outfit1.jpg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ui := NewUI(&buf, tt.theme)
			ui.SelectedFiles(tt.categoryName, tt.files)

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestUnselectedFiles(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		theme    Theme
		contains []string
	}{
		{
			name:     "all files selected",
			files:    []string{},
			theme:    Theme{UseColors: true, UseEmojis: true, Compact: false},
			contains: []string{"You've picked all the outfits from here!"},
		},
		{
			name:     "unselected files exist",
			files:    []string{"outfit3.jpg", "outfit4.jpg"},
			theme:    Theme{UseColors: true, UseEmojis: true, Compact: false},
			contains: []string{"Outfits You Haven't Picked Yet", "outfit3.jpg", "outfit4.jpg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ui := NewUI(&buf, tt.theme)
			ui.UnselectedFiles(tt.files)

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestRandomSelection(t *testing.T) {
	var buf bytes.Buffer
	theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
	ui := NewUI(&buf, theme)
	ui.RandomSelection("test-outfit.jpg")

	output := buf.String()
	expected := []string{"I picked this outfit for you", "test-outfit.jpg", "(k)eep", "(s)kip", "(q)uit"}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q, got %q", exp, output)
		}
	}
}

func TestKeepAction(t *testing.T) {
	var buf bytes.Buffer
	theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
	ui := NewUI(&buf, theme)
	ui.KeepAction("test-outfit.jpg")

	output := buf.String()
	if !strings.Contains(output, "Great choice! I've saved") || !strings.Contains(output, "test-outfit.jpg") {
		t.Errorf("expected keep action message, got %q", output)
	}
}

func TestSkipAction(t *testing.T) {
	var buf bytes.Buffer
	theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
	ui := NewUI(&buf, theme)
	ui.SkipAction("test-outfit.jpg")

	output := buf.String()
	if !strings.Contains(output, "Skipped") || !strings.Contains(output, "test-outfit.jpg") {
		t.Errorf("expected skip action message, got %q", output)
	}
}

func TestCompletionSummary(t *testing.T) {
	tests := []struct {
		name      string
		completed int
		total     int
		names     []string
		contains  []string
	}{
		{
			name:      "no completion",
			completed: 0,
			total:     3,
			names:     []string{},
			contains:  []string{"0/3"},
		},
		{
			name:      "partial completion",
			completed: 2,
			total:     3,
			names:     []string{"Beach", "Latex"},
			contains:  []string{"2/3", "Beach", "Latex"},
		},
		{
			name:      "full completion",
			completed: 3,
			total:     3,
			names:     []string{"Beach", "Latex", "General"},
			contains:  []string{"3/3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
			ui := NewUI(&buf, theme)
			ui.CompletionSummary(tt.completed, tt.total, tt.names)

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestMessageMethods(t *testing.T) {
	tests := []struct {
		name     string
		method   func(*UI, string)
		message  string
		contains []string
	}{
		{
			name:     "error message",
			method:   (*UI).Error,
			message:  "test error",
			contains: []string{"Error", "test error"},
		},
		{
			name:     "success message",
			method:   (*UI).Success,
			message:  "test success",
			contains: []string{"test success"},
		},
		{
			name:     "info message",
			method:   (*UI).Info,
			message:  "test info",
			contains: []string{"test info"},
		},
		{
			name:     "warning message",
			method:   (*UI).Warning,
			message:  "test warning",
			contains: []string{"test warning"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
			ui := NewUI(&buf, theme)
			tt.method(ui, tt.message)

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestSeparator(t *testing.T) {
	tests := []struct {
		name     string
		theme    Theme
		contains string
	}{
		{
			name:     "full separator",
			theme:    Theme{UseColors: true, UseEmojis: true, Compact: false},
			contains: "─",
		},
		{
			name:     "compact separator",
			theme:    Theme{UseColors: false, UseEmojis: false, Compact: true},
			contains: "---",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ui := NewUI(&buf, tt.theme)
			ui.Separator()

			output := buf.String()
			if !strings.Contains(output, tt.contains) {
				t.Errorf("expected output to contain %q, got %q", tt.contains, output)
			}
		})
	}
}

func TestColorize(t *testing.T) {
	tests := []struct {
		name      string
		useColors bool
		text      string
		color     string
		expected  string
	}{
		{
			name:      "with colors",
			useColors: true,
			text:      "test",
			color:     Red,
			expected:  Red + "test" + Reset,
		},
		{
			name:      "without colors",
			useColors: false,
			text:      "test",
			color:     Red,
			expected:  "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			theme := Theme{UseColors: tt.useColors, UseEmojis: false, Compact: false}
			ui := NewUI(&buf, theme)

			result := ui.colorize(tt.text, tt.color)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIcon(t *testing.T) {
	tests := []struct {
		name      string
		useEmojis bool
		emoji     string
		expected  string
	}{
		{
			name:      "with emojis",
			useEmojis: true,
			emoji:     IconCheck,
			expected:  IconCheck + " ",
		},
		{
			name:      "without emojis",
			useEmojis: false,
			emoji:     IconCheck,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			theme := Theme{UseColors: false, UseEmojis: tt.useEmojis, Compact: false}
			ui := NewUI(&buf, theme)

			result := ui.icon(tt.emoji)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCreateProgressBar(t *testing.T) {
	tests := []struct {
		name       string
		theme      Theme
		percentage float64
		width      int
		contains   []string
	}{
		{
			name:       "progress bar with colors",
			theme:      Theme{UseColors: true, UseEmojis: true, Compact: false},
			percentage: 50.0,
			width:      10,
			contains:   []string{"█", "░"},
		},
		{
			name:       "progress bar without colors",
			theme:      Theme{UseColors: false, UseEmojis: false, Compact: false},
			percentage: 25.0,
			width:      8,
			contains:   []string{"[", "=", "-", "]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ui := NewUI(&buf, tt.theme)

			result := ui.createProgressBar(tt.percentage, tt.width)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected progress bar to contain %q, got %q", expected, result)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkHeader(b *testing.B) {
	var buf bytes.Buffer
	theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
	ui := NewUI(&buf, theme)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		ui.Header("Benchmark Test")
	}
}

func BenchmarkMainMenu(b *testing.B) {
	var buf bytes.Buffer
	theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
	ui := NewUI(&buf, theme)
	categories := []string{"/path/to/Beach", "/path/to/Latex", "/path/to/General"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		ui.MainMenu(categories, nil)
	}
}
func TestUncategorizedOnlyMenu(t *testing.T) {
	var buf bytes.Buffer
	theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
	ui := NewUI(&buf, theme)

	ui.UncategorizedOnlyMenu(5)
	output := buf.String()

	expected := []string{
		"Outfit Picker", "Your Outfits", "5 outfits available",
		"What would you like to do?", "r", "s", "u", "m", "q",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q, got %q", exp, output)
		}
	}
}

func TestUncategorizedInfo(t *testing.T) {
	tests := []struct {
		name          string
		totalFiles    int
		selectedFiles int
		theme         Theme
		contains      []string
	}{
		{
			name:          "full display with progress",
			totalFiles:    10,
			selectedFiles: 3,
			theme:         Theme{UseColors: true, UseEmojis: true, Compact: false},
			contains:      []string{"Uncategorized Files", "10", "3", "30.0%"},
		},
		{
			name:          "compact display",
			totalFiles:    5,
			selectedFiles: 2,
			theme:         Theme{UseColors: false, UseEmojis: true, Compact: true},
			contains:      []string{"📄", "Uncategorized", "(2/5)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ui := NewUI(&buf, tt.theme)
			ui.UncategorizedInfo(tt.totalFiles, tt.selectedFiles)

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestManualSelectionMenu(t *testing.T) {
	var buf bytes.Buffer
	theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
	ui := NewUI(&buf, theme)

	ui.ManualSelectionMenu(3, 15)
	output := buf.String()

	expected := []string{
		"Choose Your Outfit", "3 outfit collections", "15 total outfits",
		"Go back", "q",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q, got %q", exp, output)
		}
	}
}

func TestDisplayFileGroup(t *testing.T) {
	tests := []struct {
		name          string
		groupName     string
		files         []string
		selectedFiles map[string]bool
		startIndex    int
		expectedIndex int
	}{
		{
			name:          "regular category",
			groupName:     "Formal",
			files:         []string{"suit.jpg", "dress.jpg"},
			selectedFiles: map[string]bool{"suit.jpg": true},
			startIndex:    1,
			expectedIndex: 3,
		},
		{
			name:          "uncategorized files",
			groupName:     "Uncategorized",
			files:         []string{"outfit1.jpg", "outfit2.jpg"},
			selectedFiles: map[string]bool{},
			startIndex:    5,
			expectedIndex: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			theme := Theme{UseColors: true, UseEmojis: true, Compact: false}
			ui := NewUI(&buf, theme)

			nextIndex := ui.DisplayFileGroup(tt.groupName, tt.files, tt.selectedFiles, tt.startIndex)

			if nextIndex != tt.expectedIndex {
				t.Errorf("expected next index %d, got %d", tt.expectedIndex, nextIndex)
			}

			output := buf.String()

			// Check group name display
			if tt.groupName == "Uncategorized" {
				if !strings.Contains(output, "Other Outfits") {
					t.Error("expected 'Other Outfits' for uncategorized group")
				}
			} else {
				if !strings.Contains(output, tt.groupName) {
					t.Errorf("expected group name %q in output", tt.groupName)
				}
			}

			// Check file listings
			for _, file := range tt.files {
				if !strings.Contains(output, file) {
					t.Errorf("expected file %q in output", file)
				}
			}

			// Check selected file indication
			if tt.selectedFiles["suit.jpg"] {
				if !strings.Contains(output, "already picked") {
					t.Error("expected 'already picked' indication for selected file")
				}
			}
		})
	}
}
