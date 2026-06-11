package cli

import (
	"fmt"

	"github.com/dh85/outfitpicker/internal/application/usecases"
	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
)

type RuntimeSelectionService struct {
	configManager   usecases.ConfigManager
	categoryInfo    *usecases.GetCategoriesUseCase
	pickOutfit      *usecases.PickOutfitUseCase
	session         *OutfitSession
	randomIndexFunc func(int) int
}

func NewRuntimeSelectionService(
	categoryService interfaces.CategoryService,
	configManager usecases.ConfigManager,
	cacheManager usecases.CacheManager,
	session *OutfitSession,
	randomIndexFunc func(int) int,
) *RuntimeSelectionService {
	return &RuntimeSelectionService{
		configManager:   configManager,
		categoryInfo:    usecases.NewGetCategoriesUseCase(categoryService, configManager),
		pickOutfit:      usecases.NewPickOutfitUseCase(categoryService, configManager, cacheManager),
		session:         session,
		randomIndexFunc: randomIndexFunc,
	}
}

func (s *RuntimeSelectionService) ShowNextUniqueRandomOutfit() (*entities.OutfitReference, error) {
	config, err := s.configManager.LoadOrCreate()
	if err != nil {
		return nil, err
	}

	infos, err := s.categoryInfo.Execute()
	if err != nil {
		return nil, err
	}

	var allAvailable []entities.OutfitReference
	for _, info := range infos {
		if info.State != entities.CategoryStateHasOutfits {
			continue
		}
		if config.ExcludedCategories[info.Category.Name] {
			continue
		}

		available, err := s.pickOutfit.LoadAvailableOutfits(info.Category.Name)
		if err != nil {
			return nil, err
		}
		allAvailable = append(allAvailable, available...)
	}

	if len(allAvailable) == 0 {
		return nil, nil
	}

	available := filterUnseenOutfits(allAvailable, s.session)
	if len(available) == 0 {
		s.session.ResetGlobal()
		available = allAvailable
	}

	selected := available[s.randomIndex(len(available))]
	s.session.MarkGlobalShown(outfitKey(selected))
	return &selected, nil
}

func (s *RuntimeSelectionService) ShowNextUniqueRandomOutfitFrom(categoryName string) (*entities.OutfitReference, error) {
	available, err := s.pickOutfit.LoadAvailableOutfits(categoryName)
	if err != nil {
		return nil, err
	}
	if len(available) == 0 {
		return nil, nil
	}

	unseen := filterCategoryUnseenOutfits(available, categoryName, s.session)
	if len(unseen) == 0 {
		s.session.ResetCategory(categoryName)
		unseen = available
	}

	selected := unseen[s.randomIndex(len(unseen))]
	s.session.MarkCategoryShown(selected.FileName, categoryName)
	return &selected, nil
}

func (s *RuntimeSelectionService) randomIndex(length int) int {
	if length <= 1 {
		return 0
	}
	return s.randomIndexFunc(length)
}

func outfitKey(outfit entities.OutfitReference) string {
	return fmt.Sprintf("%s/%s", outfit.Category.Name, outfit.FileName)
}

func filterUnseenOutfits(outfits []entities.OutfitReference, session *OutfitSession) []entities.OutfitReference {
	result := make([]entities.OutfitReference, 0, len(outfits))
	for _, outfit := range outfits {
		if !session.IsGlobalShown(outfitKey(outfit)) {
			result = append(result, outfit)
		}
	}
	return result
}

func filterCategoryUnseenOutfits(outfits []entities.OutfitReference, category string, session *OutfitSession) []entities.OutfitReference {
	result := make([]entities.OutfitReference, 0, len(outfits))
	for _, outfit := range outfits {
		if !session.IsCategoryShown(outfit.FileName, category) {
			result = append(result, outfit)
		}
	}
	return result
}
