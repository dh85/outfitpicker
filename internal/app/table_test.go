package app

import (
	"testing"
)

func TestUtilityFunctionsTable(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single path",
			input:    []string{"/path/to/file.jpg"},
			expected: []string{"file.jpg"},
		},
		{
			name:     "multiple paths",
			input:    []string{"/a/b.jpg", "/c/d.png"},
			expected: []string{"b.jpg", "d.png"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := baseNames(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
				return
			}
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("expected %s, got %s", expected, result[i])
				}
			}
		})
	}
}
