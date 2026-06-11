package usecases

import (
	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/errors"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
)

type GetCategoriesUseCase struct {
	categoryService interfaces.CategoryService
	configManager   ConfigManager
}

func NewGetCategoriesUseCase(categoryService interfaces.CategoryService, configManager ConfigManager) *GetCategoriesUseCase {
	return &GetCategoriesUseCase{categoryService, configManager}
}

func (uc *GetCategoriesUseCase) Execute() ([]entities.CategoryInfo, error) {
	config, err := uc.configManager.LoadOrCreate()
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errors.ErrConfigurationNotFound
	}

	return uc.categoryService.ScanCategories(config.Root, config.ExcludedCategories)
}
