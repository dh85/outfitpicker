package usecases

import (
	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/errors"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
)

type ConfigManager interface {
	LoadOrCreate() (*entities.Config, error)
	Save(config *entities.Config) error
	Delete() error
}

type CategoryManagementUseCase struct {
	categoryService interfaces.CategoryService
	configManager   ConfigManager
}

func NewCategoryManagementUseCase(categoryService interfaces.CategoryService, configManager ConfigManager) *CategoryManagementUseCase {
	return &CategoryManagementUseCase{categoryService, configManager}
}

func (uc *CategoryManagementUseCase) ScanCategories() ([]entities.CategoryInfo, error) {
	config, err := uc.configManager.LoadOrCreate()
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errors.ErrConfigurationNotFound
	}

	return uc.categoryService.ScanCategories(config.Root, config.ExcludedCategories)
}

func (uc *CategoryManagementUseCase) GetOutfits(categoryPath string) ([]entities.FileEntry, error) {
	return uc.categoryService.GetOutfits(categoryPath)
}
