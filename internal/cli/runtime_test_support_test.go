package cli

import (
	"path/filepath"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

const (
	cliTestOutfitRoot      = "/outfitpicker-test/outfits"
	cliTestNewOutfitRoot   = "/outfitpicker-test/new-outfits"
	cliTestOtherOutfitRoot = "/outfitpicker-test/other-outfits"
)

func cliTestCategoryPath(name string) string {
	return filepath.Join(cliTestOutfitRoot, name)
}

type stubCategoryInfoResult struct {
	infos []entities.CategoryInfo
	err   error
}

type stubSelectorResult struct {
	outfit *entities.OutfitReference
	err    error
}

type stubWardrobeReader struct {
	categoryInfos          []entities.CategoryInfo
	categoryInfoResults    []stubCategoryInfoResult
	categoryInfoCalls      int
	categoryInfoErr        error
	categories             []entities.CategoryReference
	categoriesErr          error
	outfitState            entities.CategoryOutfitState
	outfitStateErr         error
	outfitStates           map[string]entities.CategoryOutfitState
	outfitStateErrors      map[string]error
	allOutfitStates        map[string]entities.CategoryOutfitState
	allOutfitStatesErr     error
	availableOutfits       []entities.OutfitReference
	availableOutfitsErr    error
	availableOutfitsByName map[string][]entities.OutfitReference
	availableOutfitErrors  map[string]error
	allOutfitsByCategory   map[string][]entities.OutfitReference
	showAllOutfitsErr      error
	rootDirectory          string
	rootErr                error
}

func newStubWardrobeReader() *stubWardrobeReader {
	return &stubWardrobeReader{
		outfitStates:           map[string]entities.CategoryOutfitState{},
		outfitStateErrors:      map[string]error{},
		availableOutfitsByName: map[string][]entities.OutfitReference{},
		availableOutfitErrors:  map[string]error{},
		allOutfitsByCategory:   map[string][]entities.OutfitReference{},
	}
}

func (s *stubWardrobeReader) GetCategoryInfo() ([]entities.CategoryInfo, error) {
	if s.categoryInfoCalls < len(s.categoryInfoResults) {
		result := s.categoryInfoResults[s.categoryInfoCalls]
		s.categoryInfoCalls++
		return result.infos, result.err
	}
	s.categoryInfoCalls++
	return s.categoryInfos, s.categoryInfoErr
}

func (s *stubWardrobeReader) GetCategories() ([]entities.CategoryReference, error) {
	return s.categories, s.categoriesErr
}

func (s *stubWardrobeReader) GetOutfitState(category entities.CategoryReference) (entities.CategoryOutfitState, error) {
	if err := s.outfitStateErrors[category.Name]; err != nil {
		return entities.CategoryOutfitState{}, err
	}
	if s.outfitStateErr != nil {
		return entities.CategoryOutfitState{}, s.outfitStateErr
	}
	if state, ok := s.outfitStates[category.Name]; ok {
		return state, nil
	}
	return s.outfitState, nil
}

func (s *stubWardrobeReader) GetAllOutfitStates() (map[string]entities.CategoryOutfitState, error) {
	return s.allOutfitStates, s.allOutfitStatesErr
}

func (s *stubWardrobeReader) GetAvailableOutfits(category entities.CategoryReference) ([]entities.OutfitReference, error) {
	if err := s.availableOutfitErrors[category.Name]; err != nil {
		return nil, err
	}
	if outfits, ok := s.availableOutfitsByName[category.Name]; ok {
		return outfits, s.availableOutfitsErr
	}
	return s.availableOutfits, s.availableOutfitsErr
}

func (s *stubWardrobeReader) ShowAllOutfits(categoryName string) ([]entities.OutfitReference, error) {
	if s.showAllOutfitsErr != nil {
		return nil, s.showAllOutfitsErr
	}
	return s.allOutfitsByCategory[categoryName], nil
}

func (s *stubWardrobeReader) GetRootDirectory() (string, error) {
	return s.rootDirectory, s.rootErr
}

type stubConfigurationController struct {
	currentConfig  *entities.Config
	loadErr        error
	updateErr      error
	updatedConfigs []*entities.Config
}

func (s *stubConfigurationController) GetConfiguration() (*entities.Config, error) {
	return s.currentConfig, s.loadErr
}

func (s *stubConfigurationController) UpdateConfiguration(config *entities.Config) error {
	s.updatedConfigs = append(s.updatedConfigs, config)
	if s.updateErr == nil {
		s.currentConfig = config
	}
	return s.updateErr
}

type stubCommandHandler struct {
	wearErr            error
	wearCalls          []entities.OutfitReference
	resetCategoryErr   error
	resetCategoryCalls []string
	resetAllErr        error
	resetAllCalls      int
	factoryResetErr    error
	factoryResetCalls  int
}

func (s *stubCommandHandler) WearOutfit(outfit entities.OutfitReference) error {
	s.wearCalls = append(s.wearCalls, outfit)
	return s.wearErr
}

func (s *stubCommandHandler) ResetCategory(categoryName string) error {
	s.resetCategoryCalls = append(s.resetCategoryCalls, categoryName)
	return s.resetCategoryErr
}

func (s *stubCommandHandler) ResetAllCategories() error {
	s.resetAllCalls++
	return s.resetAllErr
}

func (s *stubCommandHandler) FactoryReset() error {
	s.factoryResetCalls++
	return s.factoryResetErr
}

type stubRandomOutfitSelector struct {
	globalResults   []stubSelectorResult
	globalCalls     int
	categoryResults []stubSelectorResult
	categoryCalls   int
}

func (s *stubRandomOutfitSelector) ShowNextUniqueRandomOutfit() (*entities.OutfitReference, error) {
	if s.globalCalls >= len(s.globalResults) {
		return nil, nil
	}
	result := s.globalResults[s.globalCalls]
	s.globalCalls++
	return result.outfit, result.err
}

func (s *stubRandomOutfitSelector) ShowNextUniqueRandomOutfitFrom(categoryName string) (*entities.OutfitReference, error) {
	if s.categoryCalls >= len(s.categoryResults) {
		return nil, nil
	}
	result := s.categoryResults[s.categoryCalls]
	s.categoryCalls++
	return result.outfit, result.err
}

type stubRuntime struct {
	wardrobe *stubWardrobeReader
	config   *stubConfigurationController
	commands *stubCommandHandler
	random   *stubRandomOutfitSelector
}

func newStubRuntime() *stubRuntime {
	return &stubRuntime{
		wardrobe: newStubWardrobeReader(),
		config:   &stubConfigurationController{},
		commands: &stubCommandHandler{},
		random:   &stubRandomOutfitSelector{},
	}
}

func newStubOutfitService(picker *stubRuntime) OutfitService {
	return NewOutfitService(picker.wardrobe, picker.config, picker.commands)
}

func (s *stubRuntime) GetCategoryInfo() ([]entities.CategoryInfo, error) {
	return s.wardrobe.GetCategoryInfo()
}

func (s *stubRuntime) GetCategories() ([]entities.CategoryReference, error) {
	return s.wardrobe.GetCategories()
}

func (s *stubRuntime) GetOutfitState(category entities.CategoryReference) (entities.CategoryOutfitState, error) {
	return s.wardrobe.GetOutfitState(category)
}

func (s *stubRuntime) GetAllOutfitStates() (map[string]entities.CategoryOutfitState, error) {
	return s.wardrobe.GetAllOutfitStates()
}

func (s *stubRuntime) GetAvailableOutfits(category entities.CategoryReference) ([]entities.OutfitReference, error) {
	return s.wardrobe.GetAvailableOutfits(category)
}

func (s *stubRuntime) ShowAllOutfits(categoryName string) ([]entities.OutfitReference, error) {
	return s.wardrobe.ShowAllOutfits(categoryName)
}

func (s *stubRuntime) GetRootDirectory() (string, error) {
	return s.wardrobe.GetRootDirectory()
}

func (s *stubRuntime) GetConfiguration() (*entities.Config, error) {
	return s.config.GetConfiguration()
}

func (s *stubRuntime) UpdateConfiguration(config *entities.Config) error {
	return s.config.UpdateConfiguration(config)
}

func (s *stubRuntime) WearOutfit(outfit entities.OutfitReference) error {
	return s.commands.WearOutfit(outfit)
}

func (s *stubRuntime) ResetCategory(categoryName string) error {
	return s.commands.ResetCategory(categoryName)
}

func (s *stubRuntime) ResetAllCategories() error {
	return s.commands.ResetAllCategories()
}

func (s *stubRuntime) FactoryReset() error {
	return s.commands.FactoryReset()
}

func (s *stubRuntime) ShowNextUniqueRandomOutfit() (*entities.OutfitReference, error) {
	return s.random.ShowNextUniqueRandomOutfit()
}

func (s *stubRuntime) ShowNextUniqueRandomOutfitFrom(categoryName string) (*entities.OutfitReference, error) {
	return s.random.ShowNextUniqueRandomOutfitFrom(categoryName)
}
