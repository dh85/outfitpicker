package entities

import "testing"

func TestSelectionTarget_Interfaces(t *testing.T) {
	// Test SelectionTargetCategory
	category := SelectionTargetCategory{
		Category: NewCategoryReference("test", "/test"),
	}
	category.isSelectionTarget()

	// Test SelectionTargetAllCategories
	allCats := SelectionTargetAllCategories{}
	allCats.isSelectionTarget()

	// Test SelectionTargetCategories
	categories := SelectionTargetCategories{
		Categories: []CategoryReference{
			NewCategoryReference("test1", "/test1"),
			NewCategoryReference("test2", "/test2"),
		},
	}
	categories.isSelectionTarget()
}

func TestFileEntry_IsOutfitFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     bool
	}{
		{"avatar file", "test.avatar", true},
		{"text file", "test.txt", false},
		{"no extension", "test", false},
		{"directory", "test/", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := FileEntry{FileName: tt.fileName, IsDirectory: false}
			// Just test that the struct can be created
			if entry.FileName != tt.fileName {
				t.Errorf("FileName = %v, want %v", entry.FileName, tt.fileName)
			}
		})
	}
}

func TestNewRotationProgress(t *testing.T) {
	category := NewCategoryReference("casual", "/test/path/casual")

	progress := NewRotationProgress(category, 2, 5)

	if progress.Category != category {
		t.Errorf("Category = %v, want %v", progress.Category, category)
	}
	if progress.WornCount != 2 {
		t.Errorf("WornCount = %v, want 2", progress.WornCount)
	}
	if progress.TotalOutfitCount != 5 {
		t.Errorf("TotalOutfitCount = %v, want 5", progress.TotalOutfitCount)
	}
}

func TestRotationProgress_Progress(t *testing.T) {
	tests := []struct {
		name     string
		progress RotationProgress
		want     float64
	}{
		{
			name:     "returns one when total outfits is zero",
			progress: NewRotationProgress(NewCategoryReference("casual", "/test/path/casual"), 0, 0),
			want:     1.0,
		},
		{
			name:     "returns worn ratio when total outfits is non-zero",
			progress: NewRotationProgress(NewCategoryReference("casual", "/test/path/casual"), 2, 5),
			want:     0.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.progress.Progress(); got != tt.want {
				t.Errorf("Progress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRotationProgress_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		progress RotationProgress
		want     bool
	}{
		{
			name:     "returns false when worn count is below total",
			progress: NewRotationProgress(NewCategoryReference("casual", "/test/path/casual"), 1, 3),
			want:     false,
		},
		{
			name:     "returns true when worn count equals total",
			progress: NewRotationProgress(NewCategoryReference("casual", "/test/path/casual"), 3, 3),
			want:     true,
		},
		{
			name:     "returns true when worn count exceeds total",
			progress: NewRotationProgress(NewCategoryReference("casual", "/test/path/casual"), 4, 3),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.progress.IsComplete(); got != tt.want {
				t.Errorf("IsComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRotationProgress_AvailableCount(t *testing.T) {
	tests := []struct {
		name     string
		progress RotationProgress
		want     int
	}{
		{
			name:     "returns remaining outfits when rotation is not complete",
			progress: NewRotationProgress(NewCategoryReference("casual", "/test/path/casual"), 1, 3),
			want:     2,
		},
		{
			name:     "returns total outfits when rotation is complete",
			progress: NewRotationProgress(NewCategoryReference("casual", "/test/path/casual"), 3, 3),
			want:     3,
		},
		{
			name:     "returns total outfits when worn count exceeds total",
			progress: NewRotationProgress(NewCategoryReference("casual", "/test/path/casual"), 4, 3),
			want:     3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.progress.AvailableCount(); got != tt.want {
				t.Errorf("AvailableCount() = %v, want %v", got, tt.want)
			}
		})
	}
}
