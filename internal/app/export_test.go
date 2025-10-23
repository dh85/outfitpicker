package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExportManager_ExportImport(t *testing.T) {
	em := NewExportManager()
	tempDir := t.TempDir()
	exportPath := filepath.Join(tempDir, "export.json")
	
	// Test data
	testData := map[string][]string{
		"/path/to/Beach":  {"bikini.jpg", "sunhat.jpg"},
		"/path/to/Formal": {"suit.jpg", "dress.jpg"},
	}
	
	// Test export
	if err := em.Export(testData, exportPath); err != nil {
		t.Errorf("export failed: %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Error("export file was not created")
	}
	
	// Test import
	imported, err := em.Import(exportPath)
	if err != nil {
		t.Errorf("import failed: %v", err)
	}
	
	// Verify data
	if len(imported) != len(testData) {
		t.Errorf("expected %d categories, got %d", len(testData), len(imported))
	}
	
	for category, files := range testData {
		importedFiles, exists := imported[category]
		if !exists {
			t.Errorf("category %s not found in imported data", category)
			continue
		}
		
		if len(importedFiles) != len(files) {
			t.Errorf("category %s: expected %d files, got %d", category, len(files), len(importedFiles))
		}
	}
}

func TestExportManager_Merge(t *testing.T) {
	em := NewExportManager()
	
	existing := map[string][]string{
		"/cat1": {"file1.jpg", "file2.jpg"},
		"/cat2": {"fileA.jpg"},
	}
	
	imported := map[string][]string{
		"/cat1": {"file2.jpg", "file3.jpg"}, // file2.jpg is duplicate
		"/cat3": {"fileX.jpg"},              // new category
	}
	
	merged := em.Merge(existing, imported)
	
	// Check cat1 - should have 3 files (no duplicates)
	if len(merged["/cat1"]) != 3 {
		t.Errorf("cat1: expected 3 files, got %d", len(merged["/cat1"]))
	}
	
	// Check cat2 - should remain unchanged
	if len(merged["/cat2"]) != 1 {
		t.Errorf("cat2: expected 1 file, got %d", len(merged["/cat2"]))
	}
	
	// Check cat3 - should be added
	if len(merged["/cat3"]) != 1 {
		t.Errorf("cat3: expected 1 file, got %d", len(merged["/cat3"]))
	}
}

func TestExportManager_ImportNonExistent(t *testing.T) {
	em := NewExportManager()
	
	_, err := em.Import("/nonexistent/file.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestExportManager_ImportInvalidJSON(t *testing.T) {
	em := NewExportManager()
	tempDir := t.TempDir()
	invalidPath := filepath.Join(tempDir, "invalid.json")
	
	// Write invalid JSON
	os.WriteFile(invalidPath, []byte("invalid json"), 0644)
	
	_, err := em.Import(invalidPath)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}