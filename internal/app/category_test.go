package app

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/storage"
)

// Test helpers and fixtures
type testFixture struct {
	t     *testing.T
	root  string
	cache *storage.Manager
}

func newTestFixture(t *testing.T) *testFixture {
	cache, _ := storage.NewManager(t.TempDir())
	return &testFixture{
		t:     t,
		root:  t.TempDir(),
		cache: cache,
	}
}

func (f *testFixture) createCategory(name string, files ...string) string {
	catPath := filepath.Join(f.root, name)
	if err := os.MkdirAll(catPath, 0755); err != nil {
		f.t.Fatalf("failed to create category %s: %v", name, err)
	}

	for _, file := range files {
		if err := os.WriteFile(filepath.Join(catPath, file), []byte("test"), 0644); err != nil {
			f.t.Fatalf("failed to create file %s: %v", file, err)
		}
	}

	return catPath
}

func (f *testFixture) createFile(path, name string) {
	if err := os.WriteFile(filepath.Join(path, name), []byte("test"), 0644); err != nil {
		f.t.Fatalf("failed to create file %s: %v", name, err)
	}
}

func (f *testFixture) runCategoryFlow(catPath, input string) (string, error) {
	pr := &prompter{r: bufio.NewReader(strings.NewReader(input))}
	var stdout bytes.Buffer
	err := runCategoryFlow(catPath, f.cache, pr, &stdout)
	return stdout.String(), err
}

func (f *testFixture) runRandomAcrossAll(categories []string, input string) (string, error) {
	pr := &prompter{r: bufio.NewReader(strings.NewReader(input))}
	var stdout bytes.Buffer
	err := randomAcrossAll(categories, f.cache, pr, &stdout)
	return stdout.String(), err
}

func (f *testFixture) assertError(err error, msg string) {
	f.t.Helper()
	if err == nil {
		f.t.Fatal("expected error but got none")
	}
}

func (f *testFixture) assertNoError(err error) {
	f.t.Helper()
	if err != nil {
		f.t.Fatalf("unexpected error: %v", err)
	}
}

func (f *testFixture) assertOutputContains(output, expected string) {
	f.t.Helper()
	if !strings.Contains(output, expected) {
		f.t.Errorf("expected output to contain %q, got: %s", expected, output)
	}
}

// Utility function tests
func TestListCategories(t *testing.T) {
	f := newTestFixture(t)

	// Create test directories
	os.MkdirAll(filepath.Join(f.root, "Category1"), 0755)
	os.MkdirAll(filepath.Join(f.root, "category2"), 0755)
	os.MkdirAll(filepath.Join(f.root, ".hidden"), 0755)
	os.MkdirAll(filepath.Join(f.root, "Downloads"), 0755)
	f.createFile(f.root, "file.txt")

	cats, err := listCategories(f.root)
	f.assertNoError(err)

	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cats))
	}

	expected := []string{
		filepath.Join(f.root, "Category1"),
		filepath.Join(f.root, "category2"),
	}
	for i, cat := range cats {
		if cat != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], cat)
		}
	}
}

func TestListCategories_Error(t *testing.T) {
	_, err := listCategories("/nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestUtilityFunctions(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{"BaseNames", testBaseNames},
		{"ToSet", testToSet},
		{"MapKeys", testMapKeys},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func testBaseNames(t *testing.T) {
	paths := []string{"/path/to/file1.txt", "/another/file2.txt"}
	names := baseNames(paths)
	expected := []string{"file1.txt", "file2.txt"}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], name)
		}
	}
}

func testToSet(t *testing.T) {
	list := []string{"a", "b", "c"}
	set := toSet(list)

	if len(set) != 3 {
		t.Fatalf("expected 3 items, got %d", len(set))
	}

	for _, item := range list {
		if !set[item] {
			t.Errorf("expected %s to be in set", item)
		}
	}
}

func testMapKeys(t *testing.T) {
	m := map[string]bool{"a": true, "b": false, "c": true}
	keys := mapKeys(m)

	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}

	keySet := toSet(keys)
	for k := range m {
		if !keySet[k] {
			t.Errorf("expected key %s to be present", k)
		}
	}
}

// Category file operations tests
func TestCategoryFileOperations(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{"FileCount", testCategoryFileCount},
		{"FileCount_Error", testCategoryFileCountError},
		{"Complete", testCategoryComplete},
		{"Complete_EmptyCategory", testCategoryCompleteEmptyCategory},
		{"Complete_Error", testCategoryCompleteError},
		{"CompletionSummary", testCategoriesCompletionSummary},
		{"CompletionSummary_ErrorAndEmpty", testCategoriesCompletionSummaryErrorAndEmpty},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func testCategoryFileCount(t *testing.T) {
	f := newTestFixture(t)
	catPath := f.createCategory("test", "file1.txt", "file2.txt")
	f.createFile(catPath, ".hidden")
	os.MkdirAll(filepath.Join(catPath, "subdir"), 0755)

	count, err := categoryFileCount(catPath)
	f.assertNoError(err)

	if count != 2 {
		t.Errorf("expected 2 files, got %d", count)
	}
}

func testCategoryFileCountError(t *testing.T) {
	_, err := categoryFileCount("/nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func testCategoryComplete(t *testing.T) {
	f := newTestFixture(t)
	catPath := f.createCategory("test", "file1.txt", "file2.txt")

	// Not complete initially
	complete, err := categoryComplete(catPath, f.cache)
	f.assertNoError(err)
	if complete {
		t.Error("expected category to not be complete")
	}

	// Add files to cache
	f.cache.Add("file1.txt", catPath)
	f.cache.Add("file2.txt", catPath)

	// Should be complete now
	complete, err = categoryComplete(catPath, f.cache)
	f.assertNoError(err)
	if !complete {
		t.Error("expected category to be complete")
	}
}

func testCategoryCompleteEmptyCategory(t *testing.T) {
	f := newTestFixture(t)
	catPath := f.createCategory("empty")

	complete, err := categoryComplete(catPath, f.cache)
	f.assertNoError(err)
	if complete {
		t.Error("expected empty category to not be complete")
	}
}

func testCategoryCompleteError(t *testing.T) {
	f := newTestFixture(t)
	complete, err := categoryComplete("/nonexistent", f.cache)
	f.assertError(err, "expected error for nonexistent directory")
	if complete {
		t.Error("expected category to not be complete on error")
	}
}

func testCategoriesCompletionSummary(t *testing.T) {
	f := newTestFixture(t)
	cat1 := f.createCategory("cat1", "file1.txt")
	cat2 := f.createCategory("cat2", "file2.txt")
	f.cache.Add("file1.txt", cat1)

	completed, total, names := categoriesCompletionSummary([]string{cat1, cat2}, f.cache)

	if completed != 1 {
		t.Errorf("expected 1 completed, got %d", completed)
	}
	if total != 2 {
		t.Errorf("expected 2 total, got %d", total)
	}
	if len(names) != 1 || names[0] != "cat1" {
		t.Errorf("expected [cat1], got %v", names)
	}
}

func testCategoriesCompletionSummaryErrorAndEmpty(t *testing.T) {
	f := newTestFixture(t)
	cat1 := f.createCategory("cat1", "file1.txt")
	cat2 := "/nonexistent"
	cat3 := f.createCategory("empty") // empty category
	f.cache.Add("file1.txt", cat1)

	completed, total, names := categoriesCompletionSummary([]string{cat1, cat2, cat3}, f.cache)

	if completed != 1 {
		t.Errorf("expected 1 completed, got %d", completed)
	}
	if total != 3 {
		t.Errorf("expected 3 total, got %d", total)
	}
	if len(names) != 1 || names[0] != "cat1" {
		t.Errorf("expected [cat1], got %v", names)
	}
}

// Category flow tests
func TestRunCategoryFlow(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testFixture) string
		input    string
		wantErr  bool
		contains []string
	}{
		{
			name:    "ReadDirError",
			setup:   func(f *testFixture) string { return "/nonexistent" },
			input:   "",
			wantErr: true,
		},
		{
			name:    "NoFiles",
			setup:   func(f *testFixture) string { return f.createCategory("empty") },
			input:   "",
			wantErr: true,
		},
		{
			name:     "ShowSelected_Empty",
			setup:    func(f *testFixture) string { return f.createCategory("test", "file1.txt") },
			input:    "s\n",
			contains: []string{"no files have been selected yet"},
		},
		{
			name: "ShowSelected_WithFiles",
			setup: func(f *testFixture) string {
				catPath := f.createCategory("test", "file1.txt")
				f.cache.Add("file1.txt", catPath)
				return catPath
			},
			input:    "s\n",
			contains: []string{"Previously Selected Files", "file1.txt"},
		},
		{
			name: "ShowUnselected_AllSelected",
			setup: func(f *testFixture) string {
				catPath := f.createCategory("test", "file1.txt")
				f.cache.Add("file1.txt", catPath)
				return catPath
			},
			input:    "u\n",
			contains: []string{"all files in this category have been selected"},
		},
		{
			name: "ShowUnselected_WithFiles",
			setup: func(f *testFixture) string {
				catPath := f.createCategory("test", "file1.txt", "file2.txt")
				f.cache.Add("file1.txt", catPath)
				return catPath
			},
			input:    "u\n",
			contains: []string{"Unselected Files", "file2.txt"},
		},
		{
			name: "Random_AllSelected",
			setup: func(f *testFixture) string {
				catPath := f.createCategory("test", "file1.txt")
				f.cache.Add("file1.txt", catPath)
				return catPath
			},
			input:    "r\n",
			contains: []string{"all files in", "have been selected"},
		},
		{
			name:     "Random_Keep",
			setup:    func(f *testFixture) string { return f.createCategory("test", "file1.txt") },
			input:    "r\nk\n",
			contains: []string{"kept and cached"},
		},
		{
			name:     "Random_Skip",
			setup:    func(f *testFixture) string { return f.createCategory("test", "file1.txt", "file2.txt") },
			input:    "r\ns\nq\n",
			contains: []string{"skipped"},
		},
		{
			name:     "Random_Quit",
			setup:    func(f *testFixture) string { return f.createCategory("test", "file1.txt") },
			input:    "r\nq\n",
			contains: []string{"Exiting"},
		},
		{
			name:     "Random_InvalidAction",
			setup:    func(f *testFixture) string { return f.createCategory("test", "file1.txt") },
			input:    "r\nx\nq\n",
			contains: []string{"invalid action"},
		},
		{
			name:     "Quit",
			setup:    func(f *testFixture) string { return f.createCategory("test", "file1.txt") },
			input:    "q\n",
			contains: []string{"Exiting"},
		},
		{
			name:    "InvalidSelection",
			setup:   func(f *testFixture) string { return f.createCategory("test", "file1.txt") },
			input:   "x\n",
			wantErr: true,
		},
		{
			name:    "Random_ReadError",
			setup:   func(f *testFixture) string { return f.createCategory("test", "file1.txt") },
			input:   "r\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTestFixture(t)
			catPath := tt.setup(f)

			var output string
			var err error

			if tt.name == "Random_ReadError" {
				pr := &prompter{r: bufio.NewReader(&errorReader{})}
				var stdout bytes.Buffer
				err = runCategoryFlow(catPath, f.cache, pr, &stdout)
				output = stdout.String()
			} else {
				output, err = f.runCategoryFlow(catPath, tt.input)
			}

			if tt.wantErr {
				f.assertError(err, "expected error")
				return
			}

			f.assertNoError(err)
			for _, expected := range tt.contains {
				f.assertOutputContains(output, expected)
			}
		})
	}
}

// Random across all tests
func TestRandomAcrossAll(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testFixture) []string
		input    string
		contains []string
	}{
		{
			name: "AllSelected",
			setup: func(f *testFixture) []string {
				cat1 := f.createCategory("cat1", "file1.txt")
				f.cache.Add("file1.txt", cat1)
				return []string{cat1}
			},
			input:    "",
			contains: []string{"all files in all categories have been selected"},
		},
		{
			name: "Keep",
			setup: func(f *testFixture) []string {
				return []string{f.createCategory("cat1", "file1.txt")}
			},
			input:    "k\n",
			contains: []string{"kept and cached"},
		},
		{
			name: "Skip",
			setup: func(f *testFixture) []string {
				return []string{f.createCategory("cat1", "file1.txt")}
			},
			input:    "s\n",
			contains: []string{"skipped"},
		},
		{
			name: "Quit",
			setup: func(f *testFixture) []string {
				return []string{f.createCategory("cat1", "file1.txt")}
			},
			input:    "q\n",
			contains: []string{"Exiting"},
		},
		{
			name: "InvalidAction",
			setup: func(f *testFixture) []string {
				return []string{f.createCategory("cat1", "file1.txt")}
			},
			input:    "x\n",
			contains: []string{"invalid action"},
		},
		{
			name: "ReadError",
			setup: func(f *testFixture) []string {
				return []string{f.createCategory("cat1", "file1.txt")}
			},
			input:    "",
			contains: []string{"invalid action"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTestFixture(t)
			categories := tt.setup(f)

			var output string
			var err error

			if tt.name == "ReadError" {
				pr := &prompter{r: bufio.NewReader(&errorReader{})}
				var stdout bytes.Buffer
				err = randomAcrossAll(categories, f.cache, pr, &stdout)
				output = stdout.String()
			} else {
				output, err = f.runRandomAcrossAll(categories, tt.input)
			}

			f.assertNoError(err)
			for _, expected := range tt.contains {
				f.assertOutputContains(output, expected)
			}
		})
	}
}

// Show functions tests
func TestShowFunctions(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{"ShowSelectedAcrossAll_Empty", testShowSelectedAcrossAllEmpty},
		{"ShowSelectedAcrossAll_WithFiles", testShowSelectedAcrossAllWithFiles},
		{"ShowUnselectedAcrossAll_AllSelected", testShowUnselectedAcrossAllAllSelected},
		{"ShowUnselectedAcrossAll_WithFiles", testShowUnselectedAcrossAllWithFiles},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func testShowSelectedAcrossAllEmpty(t *testing.T) {
	f := newTestFixture(t)
	cat1 := f.createCategory("cat1")
	var stdout bytes.Buffer

	err := showSelectedAcrossAll([]string{cat1}, f.cache, &stdout)
	f.assertNoError(err)
	f.assertOutputContains(stdout.String(), "no files have been selected yet")
}

func testShowSelectedAcrossAllWithFiles(t *testing.T) {
	f := newTestFixture(t)
	cat1 := f.createCategory("cat1")
	f.cache.Add("file1.txt", cat1)
	var stdout bytes.Buffer

	err := showSelectedAcrossAll([]string{cat1}, f.cache, &stdout)
	f.assertNoError(err)
	output := stdout.String()
	f.assertOutputContains(output, "Selected in cat1")
	f.assertOutputContains(output, "file1.txt")
}

func testShowUnselectedAcrossAllAllSelected(t *testing.T) {
	f := newTestFixture(t)
	cat1 := f.createCategory("cat1", "file1.txt")
	f.cache.Add("file1.txt", cat1)
	var stdout bytes.Buffer

	err := showUnselectedAcrossAll([]string{cat1}, f.cache, &stdout)
	f.assertNoError(err)
	f.assertOutputContains(stdout.String(), "all files in all categories have been selected")
}

func testShowUnselectedAcrossAllWithFiles(t *testing.T) {
	f := newTestFixture(t)
	cat1 := f.createCategory("cat1", "file1.txt", "file2.txt")
	f.cache.Add("file1.txt", cat1)
	var stdout bytes.Buffer

	err := showUnselectedAcrossAll([]string{cat1}, f.cache, &stdout)
	f.assertNoError(err)
	output := stdout.String()
	f.assertOutputContains(output, "Unselected in cat1")
	f.assertOutputContains(output, "file2.txt")
}
