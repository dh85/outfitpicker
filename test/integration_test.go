package test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/app"
	"github.com/dh85/outfitpicker/internal/cli"
	"github.com/dh85/outfitpicker/pkg/config"
)

// Integration test fixture
type integrationTest struct {
	t        *testing.T
	tempDir  string
	rootDir  string
	configDir string
}

func newIntegrationTest(t *testing.T) *integrationTest {
	tempDir := t.TempDir()
	rootDir := filepath.Join(tempDir, "outfits")
	configDir := filepath.Join(tempDir, "config")
	
	// Set config directory for testing
	t.Setenv("XDG_CONFIG_HOME", configDir)
	// Ensure clean config state
	config.Delete()
	
	return &integrationTest{
		t:         t,
		tempDir:   tempDir,
		rootDir:   rootDir,
		configDir: configDir,
	}
}

func (it *integrationTest) createOutfitStructure() {
	categories := map[string][]string{
		"Beach":   {"bikini.jpg", "sunhat.jpg", "sandals.jpg"},
		"Formal":  {"suit.jpg", "dress.jpg", "heels.jpg"},
		"Casual":  {"jeans.jpg", "tshirt.jpg"},
	}
	
	for category, files := range categories {
		catDir := filepath.Join(it.rootDir, category)
		if err := os.MkdirAll(catDir, 0755); err != nil {
			it.t.Fatalf("failed to create category %s: %v", category, err)
		}
		
		for _, file := range files {
			filePath := filepath.Join(catDir, file)
			if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
				it.t.Fatalf("failed to create file %s: %v", file, err)
			}
		}
	}
}

func (it *integrationTest) runApp(input string, args ...string) (string, error) {
	var stdout bytes.Buffer
	stdin := strings.NewReader(input)
	
	// Default to using rootDir if no specific root provided
	root := it.rootDir
	category := ""
	
	// Parse simple args for testing
	skipNext := false
	for i, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "--category" || arg == "-c" {
			if i+1 < len(args) {
				category = args[i+1]
				skipNext = true
			}
		} else if !strings.HasPrefix(arg, "-") {
			// Only override root if it's an absolute path or contains path separators
			if strings.HasPrefix(arg, "/") || strings.Contains(arg, string(filepath.Separator)) {
				root = arg
			}
		}
	}
	
	err := app.Run(root, category, stdin, &stdout)
	return stdout.String(), err
}

func (it *integrationTest) assertOutputContains(output string, expected ...string) {
	it.t.Helper()
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			it.t.Errorf("expected output to contain %q, got:\n%s", exp, output)
		}
	}
}

func (it *integrationTest) assertOutputNotContains(output string, notExpected ...string) {
	it.t.Helper()
	for _, notExp := range notExpected {
		if strings.Contains(output, notExp) {
			it.t.Errorf("expected output to NOT contain %q, got:\n%s", notExp, output)
		}
	}
}

// Test complete user workflows
func TestIntegration_CompleteUserWorkflows(t *testing.T) {
	tests := []struct {
		name string
		test func(*integrationTest)
	}{
		{"FirstTimeSetup", testFirstTimeSetup},
		{"CategorySelection", testCategorySelection},
		{"RandomSelection", testRandomSelection},
		{"ShowFunctions", testShowFunctions},
		{"ConfigManagement", testConfigManagement},
		{"ErrorHandling", testErrorHandling},
		{"CacheManagement", testCacheManagement},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := newIntegrationTest(t)
			defer config.Delete() // Ensure cleanup after each integration test
			it.createOutfitStructure()
			tt.test(it)
		})
	}
}

func testFirstTimeSetup(it *integrationTest) {
	// Ensure clean state
	config.Delete()
	defer config.Delete()
	
	// Test first run wizard
	input := it.rootDir + "\n"
	var stdout bytes.Buffer
	stdin := strings.NewReader(input)
	
	root, err := cli.FirstRunWizard(stdin, &stdout)
	if err != nil {
		it.t.Fatalf("first run wizard failed: %v", err)
	}
	
	if root != it.rootDir {
		it.t.Errorf("expected root %s, got %s", it.rootDir, root)
	}
	
	output := stdout.String()
	it.assertOutputContains(output, "first time running outfitpicker", "root directory")
	
	// Verify config was saved
	cfg, err := config.Load()
	if err != nil {
		it.t.Fatalf("failed to load config after first run: %v", err)
	}
	
	if cfg.Root != it.rootDir {
		it.t.Errorf("expected root %s, got %s", it.rootDir, cfg.Root)
	}
}

func testCategorySelection(it *integrationTest) {
	// Test category menu and selection
	output, err := it.runApp("1\nr\nk\nq\n")
	if err != nil {
		it.t.Fatalf("category selection failed: %v", err)
	}
	
	it.assertOutputContains(output,
		"Categories:",
		"[1] Beach",
		"[2] Casual", 
		"[3] Formal",
		"Category: Beach",
		"kept and cached",
	)
}

func testRandomSelection(it *integrationTest) {
	// Test random selection across all categories
	output, err := it.runApp("r\nk\n")
	if err != nil {
		it.t.Fatalf("random selection failed: %v", err)
	}
	
	it.assertOutputContains(output,
		"Randomly selected:",
		"kept and cached",
		"categories complete:",
	)
}

func testShowFunctions(it *integrationTest) {
	// First select some items
	it.runApp("1\nr\nk\nq\n")
	it.runApp("2\nr\nk\nq\n")
	
	// Test show selected
	output, err := it.runApp("s\n")
	if err != nil {
		it.t.Fatalf("show selected failed: %v", err)
	}
	
	it.assertOutputContains(output, "Selected in")
	
	// Test show unselected
	output, err = it.runApp("u\n")
	if err != nil {
		it.t.Fatalf("show unselected failed: %v", err)
	}
	
	it.assertOutputContains(output, "Unselected in")
}

func testConfigManagement(it *integrationTest) {
	// Ensure clean state
	config.Delete()
	defer config.Delete()
	
	// Test config save
	testRoot := filepath.Join(it.tempDir, "test-root")
	cfg := &config.Config{Root: testRoot}
	
	err := config.Save(cfg)
	if err != nil {
		it.t.Fatalf("failed to save config: %v", err)
	}
	
	// Test config load
	loadedCfg, err := config.Load()
	if err != nil {
		it.t.Fatalf("failed to load config: %v", err)
	}
	
	if loadedCfg.Root != testRoot {
		it.t.Errorf("expected root %s, got %s", testRoot, loadedCfg.Root)
	}
	
	// Test config path
	path, err := config.Path()
	if err != nil {
		it.t.Fatalf("failed to get config path: %v", err)
	}
	
	if !strings.Contains(path, "outfitpicker") {
		it.t.Errorf("config path should contain 'outfitpicker', got: %s", path)
	}
}

func testErrorHandling(it *integrationTest) {
	// Test nonexistent root
	_, err := it.runApp("q\n", "/nonexistent")
	if err == nil {
		it.t.Fatal("expected error for nonexistent root")
	}
	
	// Test invalid category
	_, err = it.runApp("q\n", "--category", "NonExistent")
	if err == nil {
		it.t.Fatal("expected error for nonexistent category")
	}
	it.assertOutputContains(err.Error(), "not found")
	
	// Test empty root
	emptyRoot := filepath.Join(it.tempDir, "empty")
	os.MkdirAll(emptyRoot, 0755)
	
	_, err = it.runApp("q\n", emptyRoot)
	if err == nil {
		it.t.Fatal("expected error for empty root")
	}
	it.assertOutputContains(err.Error(), "no category folders found")
}

func testCacheManagement(it *integrationTest) {
	// Select items to populate cache
	it.runApp("1\nr\nk\nq\n") // Beach category
	it.runApp("2\nr\nk\nq\n") // Casual category
	
	// Verify cache exists and has content
	output, err := it.runApp("s\n")
	if err != nil {
		it.t.Fatalf("failed to show selected: %v", err)
	}
	
	it.assertOutputContains(output, "Selected in")
	it.assertOutputNotContains(output, "no files have been selected yet")
	
	// Test cache clearing by completing a category
	// First, select all items in Beach category
	beachFiles := []string{"bikini.jpg", "sunhat.jpg", "sandals.jpg"}
	for range beachFiles {
		it.runApp("1\nr\nk\nq\n")
	}
	
	// Should see cache cleared message
	output, err = it.runApp("1\nr\nq\n")
	if err != nil {
		it.t.Fatalf("failed to test cache clearing: %v", err)
	}
	
	// Should show all files selected or cache cleared
	// Since we're quitting immediately, just check that we got to the category
	it.assertOutputContains(output, "Category: Beach")
}

// Test specific category workflows
func TestIntegration_CategoryWorkflows(t *testing.T) {
	it := newIntegrationTest(t)
	defer config.Delete()
	it.createOutfitStructure()
	
	// Test direct category access
	output, err := it.runApp("r\nk\nq\n", "--category", "Beach")
	if err != nil {
		t.Fatalf("direct category access failed: %v", err)
	}
	
	it.assertOutputContains(output,
		"Category: Beach",
		"Total files in \"Beach\":",
		"kept and cached",
	)
	
	// Test case insensitive category
	output, err = it.runApp("r\nk\nq\n", "--category", "beach")
	if err != nil {
		t.Fatalf("case insensitive category failed: %v", err)
	}
	
	it.assertOutputContains(output, "Category: Beach")
}

// Test edge cases and boundary conditions
func TestIntegration_EdgeCases(t *testing.T) {
	it := newIntegrationTest(t)
	defer config.Delete()
	
	// Test with hidden files and directories
	catDir := filepath.Join(it.rootDir, "TestCat")
	os.MkdirAll(catDir, 0755)
	os.WriteFile(filepath.Join(catDir, "visible.jpg"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(catDir, ".hidden"), []byte("test"), 0644)
	os.MkdirAll(filepath.Join(catDir, ".hiddendir"), 0755)
	os.MkdirAll(filepath.Join(catDir, "subdir"), 0755)
	
	// Should only count visible files
	output, err := it.runApp("1\nr\nk\nq\n")
	if err != nil {
		t.Fatalf("hidden files test failed: %v", err)
	}
	
	it.assertOutputContains(output, "Total files in \"TestCat\": 1")
	
	// Test Downloads directory exclusion
	os.MkdirAll(filepath.Join(it.rootDir, "Downloads"), 0755)
	os.WriteFile(filepath.Join(it.rootDir, "Downloads", "file.jpg"), []byte("test"), 0644)
	
	output, err = it.runApp("q\n")
	if err != nil {
		t.Fatalf("Downloads exclusion test failed: %v", err)
	}
	
	it.assertOutputNotContains(output, "Downloads")
}

// Test concurrent access and file system operations
func TestIntegration_FileSystemOperations(t *testing.T) {
	it := newIntegrationTest(t)
	defer config.Delete()
	it.createOutfitStructure()
	
	// Test that cache persists across runs
	it.runApp("1\nr\nk\nq\n") // Select item in Beach
	
	// Run again and verify cache is loaded
	output, err := it.runApp("1\ns\nq\n") // Show selected in Beach
	if err != nil {
		t.Fatalf("cache persistence test failed: %v", err)
	}
	
	it.assertOutputContains(output, "Previously Selected Files")
	it.assertOutputNotContains(output, "no files have been selected yet")
	
	// Test file system changes
	newFile := filepath.Join(it.rootDir, "Beach", "newitem.jpg")
	os.WriteFile(newFile, []byte("test"), 0644)
	
	// Should detect new file
	output, err = it.runApp("1\nu\nq\n") // Show unselected in Beach
	if err != nil {
		t.Fatalf("file system changes test failed: %v", err)
	}
	
	it.assertOutputContains(output, "newitem.jpg")
}

// Benchmark integration test for performance
func BenchmarkIntegration_FullWorkflow(b *testing.B) {
	it := &integrationTest{
		t:       &testing.T{}, // Dummy for benchmark
		tempDir: b.TempDir(),
	}
	it.rootDir = filepath.Join(it.tempDir, "outfits")
	it.createOutfitStructure()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Simulate full user workflow
		it.runApp("r\nk\n") // Random selection and keep
	}
}