// Package testutil provides shared testing utilities to reduce duplication across test files.
package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/storage"
)

// TestFixture provides common test setup and utilities
type TestFixture struct {
	T       *testing.T
	TempDir string
	Cache   *storage.Manager
}

// NewTestFixture creates a new test fixture with temporary directory and cache
func NewTestFixture(t *testing.T) *TestFixture {
	tempDir := t.TempDir()
	cache, err := storage.NewManager(tempDir)
	if err != nil {
		t.Fatalf("failed to create cache manager: %v", err)
	}
	
	return &TestFixture{
		T:       t,
		TempDir: tempDir,
		Cache:   cache,
	}
}

// CreateCategory creates a test category with the specified files
func (f *TestFixture) CreateCategory(name string, files ...string) string {
	catPath := filepath.Join(f.TempDir, name)
	if err := os.MkdirAll(catPath, 0755); err != nil {
		f.T.Fatalf("failed to create category %s: %v", name, err)
	}
	
	for _, file := range files {
		if err := f.CreateFile(catPath, file); err != nil {
			f.T.Fatalf("failed to create file %s: %v", file, err)
		}
	}
	
	return catPath
}

// CreateFile creates a test file with default content
func (f *TestFixture) CreateFile(dir, name string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte("test content"), 0644)
}

// AssertError checks that an error occurred and optionally contains expected text
func (f *TestFixture) AssertError(err error, expectedTexts ...string) {
	f.T.Helper()
	if err == nil {
		f.T.Fatal("expected error but got none")
	}
	
	for _, expected := range expectedTexts {
		if !strings.Contains(err.Error(), expected) {
			f.T.Errorf("expected error to contain %q, got: %v", expected, err)
		}
	}
}

// AssertNoError checks that no error occurred
func (f *TestFixture) AssertNoError(err error) {
	f.T.Helper()
	if err != nil {
		f.T.Fatalf("unexpected error: %v", err)
	}
}

// AssertOutputContains checks that output contains all expected strings
func (f *TestFixture) AssertOutputContains(output string, expected ...string) {
	f.T.Helper()
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			f.T.Errorf("expected output to contain %q, got:\n%s", exp, output)
		}
	}
}

// AssertOutputNotContains checks that output does not contain any of the specified strings
func (f *TestFixture) AssertOutputNotContains(output string, notExpected ...string) {
	f.T.Helper()
	for _, notExp := range notExpected {
		if strings.Contains(output, notExp) {
			f.T.Errorf("expected output to NOT contain %q, got:\n%s", notExp, output)
		}
	}
}

// AssertEqual checks that two values are equal
func (f *TestFixture) AssertEqual(got, want interface{}) {
	f.T.Helper()
	if got != want {
		f.T.Errorf("expected %v, got %v", want, got)
	}
}

// AssertNotEqual checks that two values are not equal
func (f *TestFixture) AssertNotEqual(got, notWant interface{}) {
	f.T.Helper()
	if got == notWant {
		f.T.Errorf("expected %v to not equal %v", got, notWant)
	}
}

// CreateTestStructure creates a standard test directory structure
func (f *TestFixture) CreateTestStructure() map[string][]string {
	structure := map[string][]string{
		"Beach":  {"bikini.jpg", "sunhat.jpg", "sandals.jpg"},
		"Formal": {"suit.jpg", "dress.jpg", "heels.jpg"},
		"Casual": {"jeans.jpg", "tshirt.jpg"},
	}
	
	for category, files := range structure {
		f.CreateCategory(category, files...)
	}
	
	return structure
}

// SetupConfigEnv sets up a temporary config environment for testing
func (f *TestFixture) SetupConfigEnv() {
	f.T.Setenv("XDG_CONFIG_HOME", f.TempDir)
}