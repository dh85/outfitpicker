package services

import (
	"path/filepath"
	"sort"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
	"github.com/dh85/outfitpicker/internal/domain/logic"
)

// FileManager defines filesystem operations needed by CategoryScanner.
type FileManager interface {
	ReadDir(path string) ([]entities.FileEntry, error)
	FileExists(path string) bool
}

// CategoryScanner scans filesystem for categories and outfits.
type CategoryScanner struct {
	fileManager FileManager
}

// NewCategoryScanner creates a new category scanner.
func NewCategoryScanner(fm FileManager) *CategoryScanner {
	return &CategoryScanner{fileManager: fm}
}

// ScanCategories scans a root path for categories and their outfit counts.
func (s *CategoryScanner) ScanCategories(rootPath string, excludedCategories map[string]bool) ([]entities.CategoryInfo, error) {
	entries, err := s.fileManager.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	var categories []entities.CategoryInfo
	for _, entry := range entries {
		if !entry.IsDirectory {
			continue
		}

		categoryName := entry.FileName
		categoryPath := filepath.Join(rootPath, categoryName)
		categoryRef := entities.NewCategoryReference(categoryName, categoryPath)

		if excludedCategories != nil && excludedCategories[categoryName] {
			categories = append(categories, entities.NewCategoryInfo(
				categoryRef,
				entities.CategoryStateUserExcluded,
				0,
			))
			continue
		}

		outfits, err := s.GetOutfits(categoryPath)
		if err != nil {
			return nil, err
		}

		allFiles, err := s.fileManager.ReadDir(categoryPath)
		if err != nil {
			return nil, err
		}

		var state entities.CategoryState
		if len(outfits) == 0 {
			if len(allFiles) == 0 {
				state = entities.CategoryStateEmpty
			} else {
				state = entities.CategoryStateNoAvatarFiles
			}
		} else {
			state = entities.CategoryStateHasOutfits
		}

		categories = append(categories, entities.NewCategoryInfo(
			categoryRef,
			state,
			len(outfits),
		))
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Category.Name < categories[j].Category.Name
	})

	return categories, nil
}

// GetOutfits returns all outfit files in a category path.
func (s *CategoryScanner) GetOutfits(categoryPath string) ([]entities.FileEntry, error) {
	entries, err := s.fileManager.ReadDir(categoryPath)
	if err != nil {
		return nil, err
	}

	outfits := logic.FilterOutfitFiles(entries)

	sort.Slice(outfits, func(i, j int) bool {
		return outfits[i].FileName < outfits[j].FileName
	})

	return outfits, nil
}

// Ensure CategoryScanner implements the interface
var _ interfaces.CategoryService = (*CategoryScanner)(nil)
