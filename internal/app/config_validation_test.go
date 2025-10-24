package app

import (
	"os"
	"testing"

	"github.com/dh85/outfitpicker/pkg/config"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		wantError bool
	}{
		{
			name:      "empty_root",
			config:    &config.Config{Root: ""},
			wantError: true,
		},
		{
			name:      "nonexistent_root",
			config:    &config.Config{Root: "/nonexistent/path"},
			wantError: true,
		},
		{
			name:      "path_traversal",
			config:    &config.Config{Root: "../../../etc"},
			wantError: true,
		},
		{
			name:      "invalid_language",
			config:    &config.Config{Root: os.TempDir(), Language: "invalid"},
			wantError: true,
		},
		{
			name:      "valid_config",
			config:    &config.Config{Root: os.TempDir(), Language: "en"},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Config.Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}