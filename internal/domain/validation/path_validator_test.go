package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/errors"
)

func TestPathValidator_ValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{
			name:    "valid path",
			path:    "/Users/user/outfits",
			wantErr: nil,
		},
		{
			name:    "path with spaces",
			path:    "/Users/user/my outfits",
			wantErr: nil,
		},
		{
			name:    "path with unicode",
			path:    "/Users/jose/outfits/cafe-éclair",
			wantErr: nil,
		},
		{
			name:    "path segment containing two dots",
			path:    "/Users/user/archive..old/outfits",
			wantErr: nil,
		},
		{
			name:    "path traversal with ..",
			path:    "/Users/user/../../../etc",
			wantErr: errors.ErrPathTraversal,
		},
		{
			name:    "excessive slashes are normalized by filesystem APIs",
			path:    "/outfitpicker-test////user/////outfits",
			wantErr: nil,
		},
		{
			name:    "path too long",
			path:    "/" + strings.Repeat("a", 5000),
			wantErr: errors.ErrPathTooLong,
		},
		{
			name:    "restricted path /etc",
			path:    "/etc/config",
			wantErr: errors.ErrRestrictedPath,
		},
		{
			name:    "restricted path /usr",
			path:    "/usr/local",
			wantErr: errors.ErrRestrictedPath,
		},
		{
			name:    "path with restricted prefix but different segment",
			path:    "/Users/user/usrname/outfits",
			wantErr: nil,
		},
		{
			name:    "path with tmp prefix but different segment",
			path:    "/Users/user/tmpfiles/outfits",
			wantErr: nil,
		},
		{
			name:    "invalid characters",
			path:    "/Users/user\x00/outfits",
			wantErr: errors.ErrInvalidCharacters,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidatePath() error = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("ValidatePath() error = nil, want %v", tt.wantErr)
				} else if err != tt.wantErr {
					t.Errorf("ValidatePath() error = %v, want %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestPathValidator_MaxPathLength(t *testing.T) {
	if got := MaxPathLength(); got != 4096 {
		t.Errorf("MaxPathLength() = %v, want 4096", got)
	}
}

func TestPathValidator_ValidatePath_RejectsSymlinkComponents(t *testing.T) {
	workspaceDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	baseDir, err := os.MkdirTemp(workspaceDir, "path-validator-")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})

	targetDir := filepath.Join(baseDir, "target")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	symlinkPath := filepath.Join(baseDir, "linked")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Skipf("os.Symlink() unavailable: %v", err)
	}

	err = ValidatePath(filepath.Join(symlinkPath, "outfits"))
	if err != errors.ErrSymlinkNotAllowed {
		t.Fatalf("ValidatePath() error = %v, want %v", err, errors.ErrSymlinkNotAllowed)
	}
}

func TestPathValidator_RestrictedPaths(t *testing.T) {
	paths := RestrictedPaths()
	if len(paths) == 0 {
		t.Error("RestrictedPaths() returned empty set")
	}

	// Check for common restricted paths
	found := false
	for _, p := range paths {
		if p == "/etc" || p == "/usr" {
			found = true
			break
		}
	}
	if !found {
		t.Error("RestrictedPaths() should contain /etc or /usr")
	}
}
