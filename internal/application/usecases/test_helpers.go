package usecases

import (
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

// Test error helper
var assert = struct {
	AnError error
}{
	AnError: &testError{},
}

type testError struct{}

func (e *testError) Error() string {
	return "test error"
}

// Mock repositories
type mockConfigRepo struct {
	loadResult  *entities.Config
	loadError   error
	saveError   error
	deleteError error
}

func (m *mockConfigRepo) Load() (*entities.Config, error) {
	return m.loadResult, m.loadError
}

func (m *mockConfigRepo) Save(config *entities.Config) error {
	return m.saveError
}

func (m *mockConfigRepo) Delete() error {
	return m.deleteError
}

type mockCacheRepo struct {
	loadResult  *entities.OutfitCache
	loadError   error
	saveError   error
	deleteError error
}

func (m *mockCacheRepo) Load() (*entities.OutfitCache, error) {
	return m.loadResult, m.loadError
}

func (m *mockCacheRepo) Save(cache *entities.OutfitCache) error {
	return m.saveError
}

func (m *mockCacheRepo) Delete() error {
	return m.deleteError
}

// Mock use cases
type mockConfigUseCase struct {
	loadResult  *entities.Config
	loadError   error
	saveError   error
	deleteError error
}

func (m *mockConfigUseCase) LoadOrCreate() (*entities.Config, error) {
	return m.loadResult, m.loadError
}

func (m *mockConfigUseCase) Save(config *entities.Config) error {
	return m.saveError
}

func (m *mockConfigUseCase) Delete() error {
	return m.deleteError
}

type mockCacheService struct {
	loadResult  *entities.OutfitCache
	loadError   error
	saveError   error
	saveErrors  []error
	saveCalls   int
	deleteError error
}

func (m *mockCacheService) LoadOrCreate() (*entities.OutfitCache, error) {
	return m.loadResult, m.loadError
}

func (m *mockCacheService) Save(cache *entities.OutfitCache) error {
	if m.saveCalls < len(m.saveErrors) {
		err := m.saveErrors[m.saveCalls]
		m.saveCalls++
		return err
	}
	m.saveCalls++
	return m.saveError
}

func (m *mockCacheService) Delete() error {
	return m.deleteError
}

// Mock services
type mockCategoryService struct {
	scanResult             []entities.CategoryInfo
	scanError              error
	lastScanRootPath       string
	lastExcludedCategories map[string]bool
	outfitsResult          []entities.FileEntry
	outfitsError           error
}

func (m *mockCategoryService) ScanCategories(rootPath string, excludedCategories map[string]bool) ([]entities.CategoryInfo, error) {
	m.lastScanRootPath = rootPath
	if excludedCategories == nil {
		m.lastExcludedCategories = nil
	} else {
		m.lastExcludedCategories = make(map[string]bool, len(excludedCategories))
		for key, value := range excludedCategories {
			m.lastExcludedCategories[key] = value
		}
	}
	return m.scanResult, m.scanError
}

func (m *mockCategoryService) GetOutfits(categoryPath string) ([]entities.FileEntry, error) {
	return m.outfitsResult, m.outfitsError
}

// Test assertion helpers
func assertError(t *testing.T, wantErr bool, err error) {
	t.Helper()
	if wantErr && err == nil {
		t.Error("expected error but got none")
	}
	if !wantErr && err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func assertNil(t *testing.T, wantNil bool, result interface{}) {
	t.Helper()
	if wantNil && result != nil && !isNil(result) {
		t.Error("expected nil result")
	}
	if !wantNil && (result == nil || isNil(result)) {
		t.Error("expected non-nil result")
	}
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	// Use reflection to check if the underlying value is nil
	switch v := i.(type) {
	case *entities.OutfitReference:
		return v == nil
	case *entities.Config:
		return v == nil
	case *entities.OutfitCache:
		return v == nil
	case []entities.CategoryInfo:
		return v == nil
	case []entities.FileEntry:
		return v == nil
	default:
		return false
	}
}
