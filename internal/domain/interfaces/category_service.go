package interfaces

import "github.com/dh85/outfitpicker/internal/domain/entities"

// CategoryService handles category-related operations.
type CategoryService interface {
	ScanCategories(rootPath string, excludedCategories map[string]bool) ([]entities.CategoryInfo, error)
	GetOutfits(categoryPath string) ([]entities.FileEntry, error)
}
