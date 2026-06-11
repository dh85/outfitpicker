package cli

import (
	"errors"
	"sort"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	domainerrors "github.com/dh85/outfitpicker/internal/domain/errors"
	"github.com/dh85/outfitpicker/internal/domain/logic"
)

func (a *Application) GetCategoryInfo() ([]entities.CategoryInfo, error) {
	return a.wardrobe.GetCategoryInfo()
}

func (a *Application) GetCategories() ([]entities.CategoryReference, error) {
	return a.wardrobe.GetCategories()
}

func (a *Application) GetRootDirectory() (string, error) {
	return a.wardrobe.GetRootDirectory()
}

func (a *Application) GetConfiguration() (*entities.Config, error) {
	return a.config.GetConfiguration()
}

func (a *Application) ConfigFilePath() (string, error) {
	if a.pathProvider == nil {
		return "", nil
	}
	return a.pathProvider.ConfigFilePath()
}

func (a *Application) CacheFilePath() (string, error) {
	if a.pathProvider == nil {
		return "", nil
	}
	return a.pathProvider.CacheFilePath()
}

func (a *Application) UpdateConfiguration(config *entities.Config) error {
	return a.config.UpdateConfiguration(config)
}

func (a *Application) FactoryReset() error {
	return a.commands.FactoryReset()
}

func (a *Application) GetOutfitState(category entities.CategoryReference) (entities.CategoryOutfitState, error) {
	return a.wardrobe.GetOutfitState(category)
}

func (a *Application) GetAllOutfitStates() (map[string]entities.CategoryOutfitState, error) {
	return a.wardrobe.GetAllOutfitStates()
}

func (a *Application) GetAvailableOutfits(category entities.CategoryReference) ([]entities.OutfitReference, error) {
	return a.wardrobe.GetAvailableOutfits(category)
}

func (a *Application) ShowAllOutfits(categoryName string) ([]entities.OutfitReference, error) {
	return a.wardrobe.ShowAllOutfits(categoryName)
}

func (a *Application) WearOutfit(outfit entities.OutfitReference) error {
	return a.commands.WearOutfit(outfit)
}

func (a *Application) ResetCategory(categoryName string) error {
	return a.commands.ResetCategory(categoryName)
}

func (a *Application) ResetAllCategories() error {
	return a.commands.ResetAllCategories()
}

func (a *Application) ShowNextUniqueRandomOutfit() (*entities.OutfitReference, error) {
	return a.selection.ShowNextUniqueRandomOutfit()
}

func (a *Application) ShowNextUniqueRandomOutfitFrom(categoryName string) (*entities.OutfitReference, error) {
	return a.selection.ShowNextUniqueRandomOutfitFrom(categoryName)
}

func (a *Application) resetAfterWear(categoryName string) {
	a.session.ResetAll()
	a.session.ResetCategory(categoryName)
}

func (a *Application) resetAllSessionTracking() {
	a.session.ResetAll()
}

func cloneExcludedCategories(in map[string]bool) map[string]bool {
	result := make(map[string]bool, len(in))
	for key, value := range in {
		result[key] = value
	}
	return result
}

func buildUpdatedConfig(current *entities.Config, root, language string, excluded map[string]bool) (*entities.Config, error) {
	return entities.NewConfig(root, &language, excluded, current.KnownCategories, current.KnownCategoryFiles)
}

func sortedCategoryNames(values map[string][]entities.OutfitReference) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func currentCategoryWornFileNames(state entities.CategoryOutfitState) map[string]bool {
	result := make(map[string]bool, len(state.WornOutfits))
	for _, outfit := range state.WornOutfits {
		result[outfit.FileName] = true
	}
	return result
}

func isRotationCompleteError(err error) bool {
	var rotationCompleted *domainerrors.RotationCompletedError
	return errors.As(err, &rotationCompleted)
}

func availableOutfitsFromState(state entities.CategoryOutfitState) []entities.OutfitReference {
	available := make([]entities.OutfitReference, len(state.AvailableOutfits))
	copy(available, state.AvailableOutfits)
	sort.Slice(available, func(i, j int) bool {
		return available[i].FileName < available[j].FileName
	})
	return available
}

func allOutfitsFromFiles(category entities.CategoryReference, files []entities.FileEntry) []entities.OutfitReference {
	outfits := make([]entities.OutfitReference, 0, len(files))
	for _, file := range files {
		outfits = append(outfits, entities.NewOutfitReference(file.FileName, category))
	}
	sort.Slice(outfits, func(i, j int) bool {
		return outfits[i].FileName < outfits[j].FileName
	})
	return outfits
}

func filterAvailableFiles(files []entities.FileEntry, worn map[string]bool) []entities.FileEntry {
	return logic.FilterAvailableOutfits(files, worn)
}
