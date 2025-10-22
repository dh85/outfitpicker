package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/pkg/config"
)

func assertError(t *testing.T, err error, expectError bool, msg string) {
	if expectError && err == nil {
		t.Fatal("expected error but got none")
	}
	if !expectError && err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func assertStringEquals(t *testing.T, expected, actual, msg string) {
	if expected != actual {
		t.Fatalf("%s: expected %s, got %s", msg, expected, actual)
	}
}

func assertStringContains(t *testing.T, haystack, needle, msg string) {
	if !strings.Contains(haystack, needle) {
		t.Fatalf("%s: expected to contain '%s', got: %s", msg, needle, haystack)
	}
}

func runWizardTest(input string) (string, string, error) {
	stdin := strings.NewReader(input)
	var stdout bytes.Buffer
	defer config.Delete()
	result, err := FirstRunWizard(stdin, &stdout)
	return result, stdout.String(), err
}

func TestExpandUserHome(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{"empty path", "", "", false},
		{"no tilde", "/some/path", "/some/path", false},
		{"windows with tilde", "~/path", "~/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandUserHome(tt.input)
			assertError(t, err, tt.hasError, "ExpandUserHome")
			if tt.name == "windows with tilde" && runtime.GOOS == "windows" {
				assertStringEquals(t, tt.expected, result, "windows tilde expansion")
			} else if tt.name != "windows with tilde" {
				assertStringEquals(t, tt.expected, result, "path expansion")
			}
		})
	}
}

func TestExpandUserHome_UnixTilde(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping unix tilde test on windows")
	}

	result, err := ExpandUserHome("~/test")
	assertError(t, err, false, "unix tilde expansion")
	assertStringContains(t, result, "test", "tilde expansion result")
	if strings.HasPrefix(result, "~") {
		t.Fatalf("expected tilde to be expanded, got %s", result)
	}
}

func TestReadLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{"normal input", "hello\n", "hello", false},
		{"input with spaces", "  hello world  \n", "hello world", false},
		{"empty line", "\n", "", false},
		{"eof with content", "hello", "hello", false},
		{"eof without content", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			result, err := readLine(reader)
			assertError(t, err, tt.hasError, "readLine")
			if !tt.hasError {
				assertStringEquals(t, tt.expected, result, "readLine result")
			}
		})
	}
}

func TestEnsureCacheAtRoot(t *testing.T) {
	root := t.TempDir()
	var buf bytes.Buffer

	err := EnsureCacheAtRoot(root, &buf)
	assertError(t, err, false, "EnsureCacheAtRoot")
	assertStringContains(t, buf.String(), "created cache at", "cache creation message")
}

func TestEnsureCacheAtRoot_ExistingCache(t *testing.T) {
	root := t.TempDir()
	cacheFile := filepath.Join(root, "OutfitSelectorCache.json")
	os.WriteFile(cacheFile, []byte("{}"), 0o644)

	var buf bytes.Buffer
	err := EnsureCacheAtRoot(root, &buf)
	assertError(t, err, false, "EnsureCacheAtRoot with existing cache")
	if strings.Contains(buf.String(), "created cache at") {
		t.Fatalf("should not create cache when it exists, got: %s", buf.String())
	}
}

func TestFirstRunWizard_ReadError(t *testing.T) {
	_, _, err := runWizardTest("")
	assertError(t, err, true, "empty input should cause error")
	assertStringContains(t, err.Error(), "no input provided", "error message")
}

func TestFirstRunWizard_EmptyPath(t *testing.T) {
	validPath := t.TempDir()
	result, output, err := runWizardTest(fmt.Sprintf("\n%s\n", validPath))
	assertError(t, err, false, "empty path then valid path")
	assertStringEquals(t, validPath, result, "result path")
	assertStringContains(t, output, "please enter a non-empty path", "empty path message")
}

func TestFirstRunWizard_ValidExistingDirectory(t *testing.T) {
	validPath := t.TempDir()
	result, _, err := runWizardTest(validPath + "\n")
	assertError(t, err, false, "valid existing directory")
	assertStringEquals(t, validPath, result, "result path")
}

func TestFirstRunWizard_PathIsFile(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "notadir")
	os.WriteFile(filePath, []byte("test"), 0o644)

	validPath := t.TempDir()
	result, output, err := runWizardTest(fmt.Sprintf("%s\n%s\n", filePath, validPath))
	assertError(t, err, false, "file path then valid path")
	assertStringEquals(t, validPath, result, "result path")
	assertStringContains(t, output, "path exists but is not a directory", "file error message")
}

func TestFirstRunWizard_CreateNewDirectory_Yes(t *testing.T) {
	root := t.TempDir()
	newPath := filepath.Join(root, "newdir")
	result, _, err := runWizardTest(fmt.Sprintf("%s\ny\n", newPath))
	assertError(t, err, false, "create new directory")
	assertStringEquals(t, newPath, result, "result path")

	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Fatal("expected directory to be created")
	}
}

func TestFirstRunWizard_CreateNewDirectory_No(t *testing.T) {
	root := t.TempDir()
	newPath := filepath.Join(root, "newdir")
	validPath := t.TempDir()
	result, _, err := runWizardTest(fmt.Sprintf("%s\nn\n%s\n", newPath, validPath))
	assertError(t, err, false, "decline create then valid path")
	assertStringEquals(t, validPath, result, "result path")
}

func TestFirstRunWizard_CreateNewDirectory_YesVariations(t *testing.T) {
	yesResponses := []string{"YES", "Yes", "yes", "Y", "y"}

	for _, response := range yesResponses {
		t.Run(fmt.Sprintf("response_%s", response), func(t *testing.T) {
			root := t.TempDir()
			newPath := filepath.Join(root, "newdir_"+response)
			result, _, err := runWizardTest(fmt.Sprintf("%s\n%s\n", newPath, response))
			assertError(t, err, false, "create directory with "+response)
			assertStringEquals(t, newPath, result, "result path")

			if _, err := os.Stat(newPath); os.IsNotExist(err) {
				t.Fatal("expected directory to be created")
			}
		})
	}
}

func TestFirstRunWizard_CreateDirectoryError(t *testing.T) {
	root := t.TempDir()
	readOnlyDir := filepath.Join(root, "readonly")
	os.MkdirAll(readOnlyDir, 0o555)
	defer os.Chmod(readOnlyDir, 0o755)

	newPath := filepath.Join(readOnlyDir, "newdir")
	validPath := t.TempDir()
	result, output, err := runWizardTest(fmt.Sprintf("%s\ny\n%s\n", newPath, validPath))
	assertError(t, err, false, "create directory error then valid path")
	assertStringEquals(t, validPath, result, "result path")
	assertStringContains(t, output, "failed to create directory", "create directory error")
}

func TestFirstRunWizard_StatError(t *testing.T) {
	root := t.TempDir()
	restrictedDir := filepath.Join(root, "restricted")
	os.MkdirAll(restrictedDir, 0o000)
	defer os.Chmod(restrictedDir, 0o755)

	testPath := filepath.Join(restrictedDir, "test")
	validPath := t.TempDir()
	result, output, err := runWizardTest(fmt.Sprintf("%s\n%s\n", testPath, validPath))
	assertError(t, err, false, "stat error then valid path")
	assertStringEquals(t, validPath, result, "result path")
	assertStringContains(t, output, "failed to access path", "access path error")
}

func TestExpandUserHome_HomeError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping unix tilde test on windows")
	}

	result, err := ExpandUserHome("~/test")
	if err != nil {
		assertStringEquals(t, "", result, "empty result on error")
	} else {
		assertStringContains(t, result, "test", "tilde expansion result")
	}
}

func TestEnsureCacheAtRoot_ManagerError(t *testing.T) {
	var buf bytes.Buffer
	err := EnsureCacheAtRoot("/dev/null", &buf)
	if err != nil {
		assertStringContains(t, err.Error(), "failed to init cache manager", "manager error format")
	}
}
