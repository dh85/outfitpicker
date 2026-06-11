package usecases

import (
	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/logic"
)

type ResetCategoryUseCase struct {
	configManager ConfigManager
	cacheManager  CacheManager
}

func NewResetCategoryUseCase(configManager ConfigManager, cacheManager CacheManager) *ResetCategoryUseCase {
	return &ResetCategoryUseCase{configManager, cacheManager}
}

func (uc *ResetCategoryUseCase) Execute(categoryName string) error {
	if err := logic.ValidateCategoryName(categoryName); err != nil {
		return err
	}

	if _, err := uc.configManager.LoadOrCreate(); err != nil {
		return err
	}

	cache, err := uc.cacheManager.LoadOrCreate()
	if err != nil {
		return err
	}

	updatedCache := cache.Removing(categoryName)
	return uc.cacheManager.Save(&updatedCache)
}

func (uc *ResetCategoryUseCase) ExecuteAll() error {
	if _, err := uc.configManager.LoadOrCreate(); err != nil {
		return err
	}

	newCache := entities.NewOutfitCache()
	return uc.cacheManager.Save(&newCache)
}
