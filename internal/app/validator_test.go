package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidator_ValidateRootPath(t *testing.T) {
	validator := NewValidator()
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		path      string
		setup     func() string
		wantError bool
	}{
		{"empty path", "", nil, true},
		{"whitespace path", "   ", nil, true},
		{"valid directory", tempDir, nil, false},
		{"non-existent path", "/nonexistent", nil, true},
		{"file not directory", "", func() string {
			f := filepath.Join(tempDir, "file.txt")
			_ = os.WriteFile(f, []byte("test"), 0644)
			return f
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.path
			if tt.setup != nil {
				path = tt.setup()
			}

			err := validator.ValidateRootPath(path)
			if tt.wantError && err == nil {
				t.Error("expected error")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidator_ValidateUserAction(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		action    string
		wantError bool
	}{
		{"k", false},
		{"s", false},
		{"q", false},
		{"r", false},
		{"u", false},
		{"K", false}, // case insensitive
		{"invalid", true},
		{"", true},
		{"  k  ", false}, // whitespace trimmed
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			err := validator.ValidateUserAction(tt.action)
			if tt.wantError && err == nil {
				t.Error("expected error")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidator_ValidateCategoryName(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		wantError bool
	}{
		{"ValidName", false},
		{"", true},
		{"   ", true},
		{"Name/With/Slash", true},
		{"Name\\With\\Backslash", true},
		{"Name:With:Colon", true},
		{"Name*With*Star", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCategoryName(tt.name)
			if tt.wantError && err == nil {
				t.Error("expected error")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
