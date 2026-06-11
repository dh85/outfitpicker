package cli

import (
	"sort"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

type OutfitService struct {
	wardrobe WardrobeReader
	config   ConfigurationController
	commands OutfitCommandHandler
}

func NewOutfitService(wardrobe WardrobeReader, config ConfigurationController, commands OutfitCommandHandler) OutfitService {
	return OutfitService{
		wardrobe: wardrobe,
		config:   config,
		commands: commands,
	}
}

func NewOutfitServiceFromRuntime(runtime interface {
	WardrobeReader
	ConfigurationController
	OutfitCommandHandler
}) OutfitService {
	return NewOutfitService(runtime, runtime, runtime)
}

func (s OutfitService) GetAvailableOutfits(category entities.CategoryReference) ([]entities.OutfitReference, error) {
	return s.wardrobe.GetAvailableOutfits(category)
}

func (s OutfitService) GetActualOutfitCount(category entities.CategoryReference) (int, error) {
	state, err := s.wardrobe.GetOutfitState(category)
	if err != nil {
		return 0, err
	}
	return state.TotalCount(), nil
}

func (s OutfitService) GetCategoryInfo() ([]entities.CategoryInfo, error) {
	return s.wardrobe.GetCategoryInfo()
}

func (s OutfitService) GetCategories() ([]entities.CategoryReference, error) {
	return s.wardrobe.GetCategories()
}

func (s OutfitService) GetOutfitState(category entities.CategoryReference) (entities.CategoryOutfitState, error) {
	return s.wardrobe.GetOutfitState(category)
}

func (s OutfitService) GetAllOutfitStates() (map[string]entities.CategoryOutfitState, error) {
	return s.wardrobe.GetAllOutfitStates()
}

func (s OutfitService) ShowAllOutfits(categoryName string) ([]entities.OutfitReference, error) {
	return s.wardrobe.ShowAllOutfits(categoryName)
}

func (s OutfitService) GetRootDirectory() (string, error) {
	return s.wardrobe.GetRootDirectory()
}

func (s OutfitService) GetConfiguration() (*entities.Config, error) {
	return s.config.GetConfiguration()
}

func (s OutfitService) UpdateConfiguration(config *entities.Config) error {
	return s.config.UpdateConfiguration(config)
}

func (s OutfitService) WearOutfit(outfit entities.OutfitReference) error {
	return s.commands.WearOutfit(outfit)
}

func (s OutfitService) ResetCategory(categoryName string) error {
	return s.commands.ResetCategory(categoryName)
}

func (s OutfitService) ResetAllCategories() error {
	return s.commands.ResetAllCategories()
}

func (s OutfitService) FactoryReset() error {
	return s.commands.FactoryReset()
}

func (s OutfitService) GetWornOutfits() (map[string][]entities.OutfitReference, error) {
	states, err := s.wardrobe.GetAllOutfitStates()
	if err != nil {
		return nil, err
	}

	result := map[string][]entities.OutfitReference{}
	for category, state := range states {
		if len(state.WornOutfits) == 0 {
			continue
		}
		worn := make([]entities.OutfitReference, len(state.WornOutfits))
		copy(worn, state.WornOutfits)
		sort.Slice(worn, func(i, j int) bool {
			return worn[i].FileName < worn[j].FileName
		})
		result[category] = worn
	}
	return result, nil
}

func (s OutfitService) GetUnwornOutfits() (map[string][]entities.OutfitReference, error) {
	states, err := s.wardrobe.GetAllOutfitStates()
	if err != nil {
		return nil, err
	}

	result := map[string][]entities.OutfitReference{}
	for category, state := range states {
		if len(state.AvailableOutfits) == 0 {
			continue
		}
		result[category] = availableOutfitsFromState(state)
	}
	return result, nil
}

func (s OutfitService) GetAvailableCategories() ([]entities.CategoryInfo, error) {
	infos, err := s.wardrobe.GetCategoryInfo()
	if err != nil {
		return nil, err
	}

	result := make([]entities.CategoryInfo, 0, len(infos))
	for _, info := range infos {
		if info.State == entities.CategoryStateHasOutfits {
			result = append(result, info)
		}
	}
	return result, nil
}
