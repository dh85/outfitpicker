package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test helpers
func assertError(t *testing.T, err error, expectedMsg string) {
	if err == nil {
		t.Fatal("expected error but got none")
	}
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("expected error containing '%s', got: %v", expectedMsg, err)
	}
}

func assertNoError(t *testing.T, err error, msg string) {
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func assertOutputContains(t *testing.T, output, expected, msg string) {
	if !strings.Contains(output, expected) {
		t.Fatalf("%s: expected to contain '%s', got: %s", msg, expected, output)
	}
}

func runTest(root, category, input string) (string, error) {
	var stdout bytes.Buffer
	stdin := strings.NewReader(input)
	err := Run(root, category, stdin, &stdout)
	return stdout.String(), err
}

func createTestFile(t *testing.T, dir, filename string) {
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte("test content"), 0o644); err != nil {
		t.Fatalf("failed to create test file %s: %v", path, err)
	}
}

func createTestStructure(t *testing.T) string {
	root := t.TempDir()
	categories := []string{"Category1", "Category2"}
	files := map[string][]string{
		"Category1": {"file1.txt", "file2.txt"},
		"Category2": {"file3.txt"},
	}

	for _, cat := range categories {
		catPath := filepath.Join(root, cat)
		if err := os.MkdirAll(catPath, 0o755); err != nil {
			t.Fatalf("failed to create category dir: %v", err)
		}
		for _, file := range files[cat] {
			createTestFile(t, catPath, file)
		}
	}

	return root
}

func TestRun_CacheInitError(t *testing.T) {
	_, err := runTest("/dev/null", "", "")
	assertError(t, err, "failed to")
}

func TestRun_InvalidRootPath(t *testing.T) {
	_, err := runTest("", "", "")
	assertError(t, err, "no category folders found")
}

func TestRun_ListCategoriesError(t *testing.T) {
	nonExistent := filepath.Join(t.TempDir(), "nonexistent")
	_, err := runTest(nonExistent, "", "")
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
}

func TestRun_NoCategoriesFound(t *testing.T) {
	_, err := runTest(t.TempDir(), "", "")
	assertError(t, err, "no category folders found")
}

func TestRun_CategoryOptionFound(t *testing.T) {
	root := createTestStructure(t)
	output, err := runTest(root, "Category1", "r\nq\n")
	assertNoError(t, err, "category option found")
	assertOutputContains(t, output, "Category1", "category flow")
}

func TestRun_CategoryOptionNotFound(t *testing.T) {
	root := createTestStructure(t)
	_, err := runTest(root, "NonExistent", "")
	assertError(t, err, "category \"NonExistent\" not found")
	assertError(t, err, "available: Category1, Category2")
}

func TestRun_NumericSelection_Valid(t *testing.T) {
	root := createTestStructure(t)
	output, err := runTest(root, "", "1\nr\nq\n")
	assertNoError(t, err, "numeric selection")
	assertOutputContains(t, output, "Categories", "menu display")
	assertOutputContains(t, output, "Category1", "category flow")
}

func TestRun_NumericSelection_Invalid(t *testing.T) {
	root := createTestStructure(t)
	_, err := runTest(root, "", "99\n")
	assertError(t, err, "invalid category selection")
}

func TestRun_RandomAcrossAll(t *testing.T) {
	root := createTestStructure(t)
	output, err := runTest(root, "", "r\nq\n")
	assertNoError(t, err, "random across all")
	assertOutputContains(t, output, "🎲 Randomly selected:", "random selection")
}

func TestRun_ShowSelectedAcrossAll(t *testing.T) {
	root := createTestStructure(t)
	output, err := runTest(root, "", "s\n")
	assertNoError(t, err, "show selected across all")
	assertOutputContains(t, output, "No files have been selected yet", "no selected files")
}

func TestRun_ShowUnselectedAcrossAll(t *testing.T) {
	root := createTestStructure(t)
	output, err := runTest(root, "", "u\n")
	assertNoError(t, err, "show unselected across all")
	assertOutputContains(t, output, "Category1", "unselected files")
}

func TestRun_Quit(t *testing.T) {
	root := createTestStructure(t)
	output, err := runTest(root, "", "q\n")
	assertNoError(t, err, "quit")
	assertOutputContains(t, output, "Exiting.", "exit message")
}

func TestRun_InvalidSelection(t *testing.T) {
	root := createTestStructure(t)
	_, err := runTest(root, "", "x\n")
	assertError(t, err, "invalid selection")
}

func TestRun_MenuDisplay(t *testing.T) {
	root := createTestStructure(t)
	output, err := runTest(root, "", "q\n")
	assertNoError(t, err, "menu display")

	expectedItems := []string{
		"Categories", "1", "Category1", "2", "Category2",
		"All-categories options", "r", "Select a random file",
		"s", "Show previously selected", "u", "Show unselected",
		"q", "Quit", "Enter a category number or option",
	}

	for _, item := range expectedItems {
		assertOutputContains(t, output, item, "menu item: "+item)
	}
}

func TestRun_CategoryCaseInsensitive(t *testing.T) {
	root := createTestStructure(t)
	output, err := runTest(root, "category1", "r\nq\n")
	assertNoError(t, err, "case insensitive category")
	assertOutputContains(t, output, "Category1", "category flow")
}

func TestRun_NumericSelectionEdgeCases(t *testing.T) {
	root := createTestStructure(t)

	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"zero", "0\n", false},
		{"negative", "-1\n", false},
		{"too high", "3\n", false},
		{"valid min", "1\nr\nq\n", true},
		{"valid max", "2\nr\nq\n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := runTest(root, "", tt.input)
			if tt.valid {
				assertNoError(t, err, "valid input: "+tt.name)
			} else {
				if err == nil {
					t.Fatal("expected error for invalid input")
				}
			}
		})
	}
}

func TestRun_AbsolutePathError(t *testing.T) {
	longPath := strings.Repeat("a", 1000)
	_, err := runTest(longPath, "", "")
	if err == nil {
		t.Fatal("expected error for problematic path")
	}
}

func TestRun_WithSpacesInInput(t *testing.T) {
	root := createTestStructure(t)
	output, err := runTest(root, "", "  1  \nr\nq\n")
	assertNoError(t, err, "spaces in input")
	assertOutputContains(t, output, "Category1", "category flow")
}

func TestRun_NonNumericInput(t *testing.T) {
	root := createTestStructure(t)
	_, err := runTest(root, "", "abc\n")
	assertError(t, err, "invalid selection")
}

func TestRun_EmptyInput(t *testing.T) {
	root := createTestStructure(t)
	_, err := runTest(root, "", "\n")
	assertError(t, err, "invalid selection")
}
