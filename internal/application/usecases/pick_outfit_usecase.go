package usecases

import (
	"path/filepath"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
	"github.com/dh85/outfitpicker/internal/domain/logic"
)

type PickOutfitUseCase struct {
	categoryService interfaces.CategoryService
	configManager   ConfigManager
	cacheManager    CacheManager
}

func NewPickOutfitUseCase(categoryService interfaces.CategoryService, configManager ConfigManager, cacheManager CacheManager) *PickOutfitUseCase {
	return &PickOutfitUseCase{categoryService, configManager, cacheManager}
}

func (uc *PickOutfitUseCase) LoadAvailableOutfits(categoryName string) ([]entities.OutfitReference, error) {
	if err := logic.ValidateCategoryName(categoryName); err != nil {
		return nil, err
	}

	config, err := uc.configManager.LoadOrCreate()
	if err != nil {
		return nil, err
	}

	cache, err := uc.cacheManager.LoadOrCreate()
	if err != nil {
		return nil, err
	}

	categoryPath := filepath.Join(config.Root, categoryName)
	files, err := uc.categoryService.GetOutfits(categoryPath)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, nil
	}

	categoryCache, exists := cache.Categories[categoryName]
	if !exists {
		categoryCache = entities.NewCategoryCache(len(files))
	}

	if logic.ShouldResetRotation(len(categoryCache.WornOutfits), len(files)) {
		return nil, nil
	}

	pool := logic.FilterAvailableOutfits(files, categoryCache.WornOutfits)

	if len(pool) == 0 {
		pool = files
	}

	outfits := make([]entities.OutfitReference, 0, len(pool))
	category := entities.NewCategoryReference(categoryName, categoryPath)
	for _, file := range pool {
		outfits = append(outfits, entities.NewOutfitReference(file.FileName, category))
	}

	return outfits, nil
}
