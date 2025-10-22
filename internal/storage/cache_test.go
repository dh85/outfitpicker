package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func createManager(t *testing.T, rootPath string) *Manager {
	m, err := NewManager(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func assertNonEmptyPath(t *testing.T, m *Manager) {
	if m.Path() == "" {
		t.Fatal("expected non-empty cache path")
	}
}

func assertEmptyMap(t *testing.T, cm Map, msg string) {
	if len(cm) != 0 {
		t.Fatal(msg)
	}
}

func TestNewManager_ValidRoot(t *testing.T) {
	root := t.TempDir()
	m := createManager(t, root)
	expected := filepath.Join(root, cacheFileName)
	if m.Path() != expected {
		t.Fatalf("expected %s, got %s", expected, m.Path())
	}
}

func TestNewManager_InvalidRoot(t *testing.T) {
	m := createManager(t, "/nonexistent")
	assertNonEmptyPath(t, m)
}

func TestNewManager_EmptyRoot(t *testing.T) {
	m := createManager(t, "")
	assertNonEmptyPath(t, m)
}

func TestNewManager_RootIsFile(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "notadir")
	os.WriteFile(filePath, []byte("test"), 0o644)
	m := createManager(t, filePath)
	if m.Path() == filepath.Join(filePath, cacheFileName) {
		t.Fatal("should not use file path as directory")
	}
}

func TestManager_LoadNonexistentFile(t *testing.T) {
	m := createManager(t, t.TempDir())
	cm := m.Load()
	assertEmptyMap(t, cm, "expected empty map for nonexistent file")
}

func TestManager_LoadInvalidJSON(t *testing.T) {
	m := createManager(t, t.TempDir())
	os.WriteFile(m.Path(), []byte("invalid json"), 0o644)
	cm := m.Load()
	assertEmptyMap(t, cm, "expected empty map for invalid JSON")
}

func TestManager_SaveAndLoad(t *testing.T) {
	m := createManager(t, t.TempDir())
	original := Map{"cat1": []string{"file1", "file2"}}
	m.Save(original)
	loaded := m.Load()
	files := loaded["cat1"]
	if len(files) != 2 || files[0] != "file1" || files[1] != "file2" {
		t.Fatalf("expected %v, got %v", original, loaded)
	}
}

func TestManager_SaveInvalidData(t *testing.T) {
	m := createManager(t, t.TempDir())
	invalidMap := Map{"key": []string{string([]byte{0xff, 0xfe, 0xfd})}}
	m.Save(invalidMap) // Should not panic
}

func TestManager_SaveToReadOnlyDir(t *testing.T) {
	root := t.TempDir()
	readOnlyDir := filepath.Join(root, "readonly")
	os.MkdirAll(readOnlyDir, 0o555)
	defer os.Chmod(readOnlyDir, 0o755)
	m := &Manager{cacheFile: filepath.Join(readOnlyDir, "cache.json")}
	m.Save(Map{"test": []string{"file"}}) // Should not panic
}

func TestManager_AddAndClear(t *testing.T) {
	root := t.TempDir()
	m := createManager(t, root)
	cat := filepath.Join(root, "Beach")
	os.MkdirAll(cat, 0o755)

	m.Add("a.avatar", cat)
	m.Add("a.avatar", cat) // duplicate

	cm := m.Load()
	if len(cm[cat]) != 1 || cm[cat][0] != "a.avatar" {
		t.Fatalf("expected [a.avatar], got %v", cm[cat])
	}

	m.Clear(cat)
	cm = m.Load()
	if _, ok := cm[cat]; ok {
		t.Fatal("expected category to be cleared")
	}
}

func TestManager_AddMultipleFiles(t *testing.T) {
	m := createManager(t, t.TempDir())
	cat := "category"
	m.Add("file1", cat)
	m.Add("file2", cat)
	cm := m.Load()
	if len(cm[cat]) != 2 {
		t.Fatalf("expected 2 files, got %d", len(cm[cat]))
	}
}
