package version

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
	
	if len(Version) == 0 {
		t.Error("Version should have non-zero length")
	}
}

func TestVersionFormat(t *testing.T) {
	// Version should be either "dev" or follow semantic versioning
	if Version != "dev" && !isValidSemVer(Version) {
		t.Errorf("Version should be 'dev' or valid semver, got: %s", Version)
	}
}

func TestGetVersion(t *testing.T) {
	// Test GetVersion function
	version := GetVersion()
	if version == "" {
		t.Error("GetVersion should not return empty string")
	}
	if !strings.Contains(version, Version) {
		t.Errorf("GetVersion should contain base version, got: %s", version)
	}
}

// Simple semver validation for testing
func isValidSemVer(v string) bool {
	if strings.HasPrefix(v, "v") {
		v = v[1:]
	}
	
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return false
	}
	
	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	
	return true
}