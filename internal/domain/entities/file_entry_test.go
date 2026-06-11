package entities

import (
	"path/filepath"
	"testing"
)

func TestNewFileEntryWithDir(t *testing.T) {
	entry := NewFileEntryWithDir("/outfitpicker-test/outfits/casual/look.avatar", true)
	wantCategoryPath := filepath.Join("/outfitpicker-test/outfits", "casual")

	if entry.filePath != "/outfitpicker-test/outfits/casual/look.avatar" {
		t.Fatalf("filePath = %q, want /outfitpicker-test/outfits/casual/look.avatar", entry.filePath)
	}
	if entry.CategoryPath() != wantCategoryPath {
		t.Fatalf("CategoryPath() = %q, want %q", entry.CategoryPath(), wantCategoryPath)
	}
	if entry.CategoryName() != "casual" {
		t.Fatalf("CategoryName() = %q, want casual", entry.CategoryName())
	}
	if entry.FileName != "look.avatar" {
		t.Fatalf("FileName = %q, want look.avatar", entry.FileName)
	}
	if !entry.IsDirectory {
		t.Fatal("IsDirectory = false, want true")
	}
}

func TestFileEntry_Properties(t *testing.T) {
	tests := []struct {
		name             string
		filePath         string
		wantFileName     string
		wantCategoryPath string
		wantCategoryName string
	}{
		{
			name:             "unix path",
			filePath:         "/Users/user/outfits/casual/jeans-tshirt.avatar",
			wantFileName:     "jeans-tshirt.avatar",
			wantCategoryPath: filepath.Join("/Users/user/outfits", "casual"),
			wantCategoryName: "casual",
		},
		{
			name:             "nested path",
			filePath:         "/Users/user/outfits/work/suit-tie.avatar",
			wantFileName:     "suit-tie.avatar",
			wantCategoryPath: filepath.Join("/Users/user/outfits", "work"),
			wantCategoryName: "work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := NewFileEntry(tt.filePath)

			if got := entry.FileName; got != tt.wantFileName {
				t.Errorf("FileName = %v, want %v", got, tt.wantFileName)
			}

			if got := entry.CategoryPath(); got != tt.wantCategoryPath {
				t.Errorf("CategoryPath() = %v, want %v", got, tt.wantCategoryPath)
			}

			if got := entry.CategoryName(); got != tt.wantCategoryName {
				t.Errorf("CategoryName() = %v, want %v", got, tt.wantCategoryName)
			}
		})
	}
}

func TestFileEntry_Equality(t *testing.T) {
	t.Run("identical paths are equal", func(t *testing.T) {
		entry1 := NewFileEntry("/Users/user/outfits/casual/jeans.avatar")
		entry2 := NewFileEntry("/Users/user/outfits/casual/jeans.avatar")

		if entry1 != entry2 {
			t.Error("identical FileEntries should be equal")
		}
	})

	t.Run("different paths are not equal", func(t *testing.T) {
		entry1 := NewFileEntry("/Users/user/outfits/casual/jeans.avatar")
		entry2 := NewFileEntry("/Users/user/outfits/casual/shirt.avatar")

		if entry1 == entry2 {
			t.Error("different FileEntries should not be equal")
		}
	})
}
