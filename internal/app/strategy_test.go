package app

import (
	"testing"
)

func TestRandomStrategy(t *testing.T) {
	strategy := RandomStrategy{}

	if strategy.Name() != "random" {
		t.Error("expected name 'random'")
	}

	// Test empty files
	result := strategy.SelectFile([]FileEntry{})
	if result.FileName != "" {
		t.Error("expected empty result for empty files")
	}

	// Test with files
	files := []FileEntry{
		{FileName: "file1.jpg"},
		{FileName: "file2.jpg"},
		{FileName: "file3.jpg"},
	}

	result = strategy.SelectFile(files)
	if result.FileName == "" {
		t.Error("expected non-empty result")
	}

	// Verify result is one of the input files
	found := false
	for _, f := range files {
		if f.FileName == result.FileName {
			found = true
			break
		}
	}
	if !found {
		t.Error("result should be one of input files")
	}
}

func TestRoundRobinStrategy(t *testing.T) {
	strategy := &RoundRobinStrategy{}

	if strategy.Name() != "round-robin" {
		t.Error("expected name 'round-robin'")
	}

	files := []FileEntry{
		{FileName: "a.jpg"},
		{FileName: "b.jpg"},
		{FileName: "c.jpg"},
	}

	// Test cycling behavior
	results := make([]string, 6)
	for i := 0; i < 6; i++ {
		result := strategy.SelectFile(files)
		results[i] = result.FileName
	}

	// Should cycle through files (order determined by sorting)
	// Just verify we get each file twice in 6 selections
	count := make(map[string]int)
	for _, result := range results {
		count[result]++
	}

	for _, file := range files {
		if count[file.FileName] != 2 {
			t.Errorf("expected file %s to appear 2 times, got %d", file.FileName, count[file.FileName])
		}
	}
}

func TestStrategyFactory(t *testing.T) {
	factory := StrategyFactory{}

	tests := []struct {
		name     string
		expected string
	}{
		{"random", "random"},
		{"round-robin", "round-robin"},
		{"weighted", "weighted"},
		{"unknown", "random"}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := factory.Create(tt.name)
			if strategy.Name() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, strategy.Name())
			}
		})
	}
}
