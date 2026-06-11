package cli

import "github.com/dh85/outfitpicker/internal/domain/entities"

type WardrobeReader interface {
	GetCategoryInfo() ([]entities.CategoryInfo, error)
	GetCategories() ([]entities.CategoryReference, error)
	GetOutfitState(category entities.CategoryReference) (entities.CategoryOutfitState, error)
	GetAllOutfitStates() (map[string]entities.CategoryOutfitState, error)
	GetAvailableOutfits(category entities.CategoryReference) ([]entities.OutfitReference, error)
	ShowAllOutfits(categoryName string) ([]entities.OutfitReference, error)
	GetRootDirectory() (string, error)
}

type ConfigurationController interface {
	GetConfiguration() (*entities.Config, error)
	UpdateConfiguration(config *entities.Config) error
}

type OutfitCommandHandler interface {
	WearOutfit(outfit entities.OutfitReference) error
	ResetCategory(categoryName string) error
	ResetAllCategories() error
	FactoryReset() error
}

type RandomOutfitSelector interface {
	ShowNextUniqueRandomOutfit() (*entities.OutfitReference, error)
	ShowNextUniqueRandomOutfitFrom(categoryName string) (*entities.OutfitReference, error)
}
