package services

import (
	"path/filepath"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/errors"
)

func TestCategoryScanner_ScanCategories(t *testing.T) {
	t.Run("returns sorted categories", func(t *testing.T) {
		fm := &fakeFileManager{
			dirs: map[string][]string{
				"/test": {"casual", "formal"},
			},
			files: map[string][]string{
				testCategoryPath("casual"): {"outfit1.avatar", "outfit2.avatar"},
				testCategoryPath("formal"): {"suit.avatar"},
			},
		}
		scanner := NewCategoryScanner(fm)

		result, err := scanner.ScanCategories("/test", nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 categories, got %d", len(result))
		}
		if result[0].Category.Name != "casual" {
			t.Errorf("expected first category to be 'casual', got %s", result[0].Category.Name)
		}
		if result[1].Category.Name != "formal" {
			t.Errorf("expected second category to be 'formal', got %s", result[1].Category.Name)
		}
	})

	t.Run("skips non-directory root entries", func(t *testing.T) {
		fm := &fakeFileManager{
			dirs: map[string][]string{
				"/test": {"casual"},
			},
			files: map[string][]string{
				"/test":                    {"notes.txt"},
				testCategoryPath("casual"): {"outfit.avatar"},
			},
		}
		scanner := NewCategoryScanner(fm)

		result, err := scanner.ScanCategories("/test", nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 category after skipping root file, got %d", len(result))
		}
		if result[0].Category.Name != "casual" {
			t.Errorf("expected only category to be 'casual', got %s", result[0].Category.Name)
		}
	})

	t.Run("excludes specified categories", func(t *testing.T) {
		fm := &fakeFileManager{
			dirs: map[string][]string{
				"/test": {"casual", "old"},
			},
			files: map[string][]string{
				testCategoryPath("casual"): {"outfit.avatar"},
				testCategoryPath("old"):    {"outfit.avatar"},
			},
		}
		scanner := NewCategoryScanner(fm)

		result, err := scanner.ScanCategories("/test", map[string]bool{"old": true})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 categories, got %d", len(result))
		}
		excluded := result[1]
		if excluded.Category.Name != "old" {
			t.Errorf("expected excluded category 'old', got %s", excluded.Category.Name)
		}
		if excluded.State != entities.CategoryStateUserExcluded {
			t.Errorf("expected state UserExcluded, got %v", excluded.State)
		}
	})

	t.Run("detects empty categories", func(t *testing.T) {
		fm := &fakeFileManager{
			dirs: map[string][]string{
				"/test": {"empty"},
			},
			files: map[string][]string{
				testCategoryPath("empty"): {},
			},
		}
		scanner := NewCategoryScanner(fm)

		result, err := scanner.ScanCategories("/test", nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result[0].State != entities.CategoryStateEmpty {
			t.Errorf("expected state Empty, got %v", result[0].State)
		}
	})

	t.Run("detects no avatar files", func(t *testing.T) {
		fm := &fakeFileManager{
			dirs: map[string][]string{
				"/test": {"noavatars"},
			},
			files: map[string][]string{
				testCategoryPath("noavatars"): {"readme.txt"},
			},
		}
		scanner := NewCategoryScanner(fm)

		result, err := scanner.ScanCategories("/test", nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result[0].State != entities.CategoryStateNoAvatarFiles {
			t.Errorf("expected state NoAvatarFiles, got %v", result[0].State)
		}
	})

	t.Run("returns error on filesystem failure", func(t *testing.T) {
		fm := &fakeFileManager{err: errors.ErrFileSystem}
		scanner := NewCategoryScanner(fm)

		_, err := scanner.ScanCategories("/test", nil)

		if err != errors.ErrFileSystem {
			t.Errorf("expected ErrFileSystem, got %v", err)
		}
	})

	t.Run("returns error when getting outfits for category fails", func(t *testing.T) {
		fm := &fakeFileManager{
			dirs: map[string][]string{
				"/test": {"casual"},
			},
			readDirErrors: map[string]error{
				testCategoryPath("casual"): errors.ErrFileSystem,
			},
		}
		scanner := NewCategoryScanner(fm)

		_, err := scanner.ScanCategories("/test", nil)

		if err != errors.ErrFileSystem {
			t.Errorf("expected ErrFileSystem from GetOutfits, got %v", err)
		}
	})

	t.Run("returns error when reading all category files fails after outfits load", func(t *testing.T) {
		fm := &fakeFileManager{
			dirs: map[string][]string{
				"/test": {"casual"},
			},
			files: map[string][]string{
				testCategoryPath("casual"): {"outfit.avatar"},
			},
			readDirErrorSequence: map[string][]error{
				testCategoryPath("casual"): {nil, errors.ErrFileSystem},
			},
		}
		scanner := NewCategoryScanner(fm)

		_, err := scanner.ScanCategories("/test", nil)

		if err != errors.ErrFileSystem {
			t.Errorf("expected ErrFileSystem from second ReadDir, got %v", err)
		}
	})
}

func TestCategoryScanner_GetOutfits(t *testing.T) {
	t.Run("returns sorted outfit files", func(t *testing.T) {
		fm := &fakeFileManager{
			files: map[string][]string{
				testCategoryPath("casual"): {"zebra.avatar", "apple.avatar", "readme.txt"},
			},
		}
		scanner := NewCategoryScanner(fm)

		result, err := scanner.GetOutfits(testCategoryPath("casual"))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 outfits, got %d", len(result))
		}
		if result[0].FileName != "apple.avatar" {
			t.Errorf("expected first outfit 'apple.avatar', got %s", result[0].FileName)
		}
		if result[1].FileName != "zebra.avatar" {
			t.Errorf("expected second outfit 'zebra.avatar', got %s", result[1].FileName)
		}
	})

	t.Run("returns error on filesystem failure", func(t *testing.T) {
		fm := &fakeFileManager{err: errors.ErrFileSystem}
		scanner := NewCategoryScanner(fm)

		_, err := scanner.GetOutfits(testCategoryPath("casual"))

		if err != errors.ErrFileSystem {
			t.Errorf("expected ErrFileSystem, got %v", err)
		}
	})
}

type fakeFileManager struct {
	dirs                 map[string][]string
	files                map[string][]string
	readDirErrors        map[string]error
	readDirErrorSequence map[string][]error
	readDirCalls         map[string]int
	err                  error
}

func (f *fakeFileManager) ReadDir(path string) ([]entities.FileEntry, error) {
	if f.err != nil {
		return nil, f.err
	}
	if sequence, ok := f.readDirErrorSequence[path]; ok {
		if f.readDirCalls == nil {
			f.readDirCalls = map[string]int{}
		}
		call := f.readDirCalls[path]
		f.readDirCalls[path] = call + 1
		if call < len(sequence) && sequence[call] != nil {
			return nil, sequence[call]
		}
	}
	if err, ok := f.readDirErrors[path]; ok {
		return nil, err
	}
	var entries []entities.FileEntry
	if dirs, ok := f.dirs[path]; ok {
		for _, dir := range dirs {
			entries = append(entries, entities.NewFileEntryWithDir(filepath.Join(path, dir), true))
		}
	}
	if files, ok := f.files[path]; ok {
		for _, file := range files {
			entries = append(entries, entities.NewFileEntryWithDir(filepath.Join(path, file), false))
		}
	}
	return entries, nil
}

func (f *fakeFileManager) FileExists(path string) bool {
	return true
}

func testCategoryPath(name string) string {
	return filepath.Join("/test", name)
}
