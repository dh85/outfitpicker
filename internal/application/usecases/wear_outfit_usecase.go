package usecases

import (
	"path/filepath"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/errors"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
	"github.com/dh85/outfitpicker/internal/domain/logic"
)

type WearOutfitUseCase struct {
	categoryService interfaces.CategoryService
	configManager   ConfigManager
	cacheManager    CacheManager
}

func NewWearOutfitUseCase(categoryService interfaces.CategoryService, configManager ConfigManager, cacheManager CacheManager) *WearOutfitUseCase {
	return &WearOutfitUseCase{categoryService, configManager, cacheManager}
}

func (uc *WearOutfitUseCase) Execute(outfit entities.OutfitReference) error {
	if err := logic.ValidateOutfit(outfit); err != nil {
		return err
	}

	config, err := uc.configManager.LoadOrCreate()
	if err != nil {
		return err
	}

	cache, err := uc.cacheManager.LoadOrCreate()
	if err != nil {
		return err
	}

	categoryPath := filepath.Join(config.Root, outfit.Category.Name)
	files, err := uc.categoryService.GetOutfits(categoryPath)
	if err != nil {
		return err
	}

	found := false
	for _, f := range files {
		if f.FileName == outfit.FileName {
			found = true
			break
		}
	}
	if !found {
		return errors.ErrNoOutfitsAvailable
	}

	categoryCache, exists := cache.Categories[outfit.Category.Name]
	if !exists {
		categoryCache = entities.NewCategoryCache(len(files))
	}

	if categoryCache.WornOutfits[outfit.FileName] {
		return nil
	}

	categoryCache = categoryCache.Adding(outfit.FileName)
	updatedCache := cache.Updating(outfit.Category.Name, categoryCache)
	if err := uc.cacheManager.Save(&updatedCache); err != nil {
		return err
	}

	if logic.ShouldResetRotation(len(categoryCache.WornOutfits), len(files)) {
		return errors.NewRotationCompletedError(outfit.Category.Name)
	}

	return nil
}
