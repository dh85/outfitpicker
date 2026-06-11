package usecases

import (
	"path/filepath"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	domainerrors "github.com/dh85/outfitpicker/internal/domain/errors"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
)

type WardrobeQueries struct {
	configManager ConfigManager
	cacheManager  CacheManager
	categorySvc   interfaces.CategoryService
	categoryInfo  *GetCategoriesUseCase
}

func NewWardrobeQueries(configManager ConfigManager, cacheManager CacheManager, categorySvc interfaces.CategoryService) *WardrobeQueries {
	return &WardrobeQueries{
		configManager: configManager,
		cacheManager:  cacheManager,
		categorySvc:   categorySvc,
		categoryInfo:  NewGetCategoriesUseCase(categorySvc, configManager),
	}
}

func (q *WardrobeQueries) GetConfiguration() (*entities.Config, error) {
	config, err := q.configManager.LoadOrCreate()
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, domainerrors.ErrConfigurationNotFound
	}
	return config, nil
}

func (q *WardrobeQueries) GetCategoryInfo() ([]entities.CategoryInfo, error) {
	return q.categoryInfo.Execute()
}

func (q *WardrobeQueries) GetCategories() ([]entities.CategoryReference, error) {
	infos, err := q.GetCategoryInfo()
	if err != nil {
		return nil, err
	}

	result := make([]entities.CategoryReference, 0, len(infos))
	for _, info := range infos {
		if info.State == entities.CategoryStateHasOutfits {
			result = append(result, info.Category)
		}
	}
	return result, nil
}

func (q *WardrobeQueries) GetRootDirectory() (string, error) {
	config, err := q.GetConfiguration()
	if err != nil {
		return "", err
	}
	return config.Root, nil
}

func (q *WardrobeQueries) GetOutfitState(category entities.CategoryReference) (entities.CategoryOutfitState, error) {
	config, err := q.GetConfiguration()
	if err != nil {
		return entities.CategoryOutfitState{}, err
	}

	cache, err := q.cacheManager.LoadOrCreate()
	if err != nil {
		return entities.CategoryOutfitState{}, err
	}

	categoryPath := filepath.Join(config.Root, category.Name)
	files, err := q.categorySvc.GetOutfits(categoryPath)
	if err != nil {
		return entities.CategoryOutfitState{}, err
	}

	categoryRef := entities.NewCategoryReference(category.Name, categoryPath)
	categoryCache, ok := cache.Categories[category.Name]
	if !ok {
		categoryCache = entities.NewCategoryCache(len(files))
	}

	allOutfits := make([]entities.OutfitReference, 0, len(files))
	wornOutfits := make([]entities.OutfitReference, 0, len(files))
	availableOutfits := make([]entities.OutfitReference, 0, len(files))
	for _, file := range files {
		outfit := entities.NewOutfitReference(file.FileName, categoryRef)
		allOutfits = append(allOutfits, outfit)
		if categoryCache.WornOutfits[file.FileName] {
			wornOutfits = append(wornOutfits, outfit)
			continue
		}
		availableOutfits = append(availableOutfits, outfit)
	}

	return entities.NewCategoryOutfitState(categoryRef, allOutfits, availableOutfits, wornOutfits), nil
}

func (q *WardrobeQueries) GetAllOutfitStates() (map[string]entities.CategoryOutfitState, error) {
	categories, err := q.GetCategories()
	if err != nil {
		return nil, err
	}

	states := make(map[string]entities.CategoryOutfitState, len(categories))
	for _, category := range categories {
		state, err := q.GetOutfitState(category)
		if err != nil {
			return nil, err
		}
		states[category.Name] = state
	}
	return states, nil
}

func (q *WardrobeQueries) GetAvailableOutfits(category entities.CategoryReference) ([]entities.OutfitReference, error) {
	state, err := q.GetOutfitState(category)
	if err != nil {
		return nil, err
	}
	return state.AvailableOutfits, nil
}

func (q *WardrobeQueries) ShowAllOutfits(categoryName string) ([]entities.OutfitReference, error) {
	config, err := q.GetConfiguration()
	if err != nil {
		return nil, err
	}

	categoryPath := filepath.Join(config.Root, categoryName)
	files, err := q.categorySvc.GetOutfits(categoryPath)
	if err != nil {
		return nil, err
	}

	category := entities.NewCategoryReference(categoryName, categoryPath)
	outfits := make([]entities.OutfitReference, 0, len(files))
	for _, file := range files {
		outfits = append(outfits, entities.NewOutfitReference(file.FileName, category))
	}
	return outfits, nil
}
