package usecases

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	domainerrors "github.com/dh85/outfitpicker/internal/domain/errors"
)

const wardrobeRoot = "/outfitpicker-test/outfits"

func TestWardrobeQueries_GetConfiguration(t *testing.T) {
	t.Run("returns loaded configuration", func(t *testing.T) {
		config := mustWardrobeConfig(t, nil)
		queries := newWardrobeQueries(config, entities.NewOutfitCache(), nil)

		got, err := queries.GetConfiguration()
		if err != nil {
			t.Fatalf("GetConfiguration() error = %v", err)
		}
		if got != config {
			t.Fatalf("GetConfiguration() = %#v, want %#v", got, config)
		}
	})

	t.Run("propagates load error", func(t *testing.T) {
		wantErr := errors.New("load failed")
		queries := NewWardrobeQueries(
			&mockConfigUseCase{loadError: wantErr},
			&mockCacheService{loadResult: newWardrobeCache()},
			&wardrobeCategoryService{},
		)

		got, err := queries.GetConfiguration()
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetConfiguration() error = %v, want %v", err, wantErr)
		}
		if got != nil {
			t.Fatalf("GetConfiguration() = %#v, want nil", got)
		}
	})

	t.Run("nil configuration is not found", func(t *testing.T) {
		queries := NewWardrobeQueries(
			&mockConfigUseCase{},
			&mockCacheService{loadResult: newWardrobeCache()},
			&wardrobeCategoryService{},
		)

		got, err := queries.GetConfiguration()
		if !errors.Is(err, domainerrors.ErrConfigurationNotFound) {
			t.Fatalf("GetConfiguration() error = %v, want %v", err, domainerrors.ErrConfigurationNotFound)
		}
		if got != nil {
			t.Fatalf("GetConfiguration() = %#v, want nil", got)
		}
	})
}

func TestWardrobeQueries_GetCategoryInfoAndCategories(t *testing.T) {
	config := mustWardrobeConfig(t, map[string]bool{"excluded": true})
	service := &wardrobeCategoryService{
		scanResult: []entities.CategoryInfo{
			categoryInfo("casual", entities.CategoryStateHasOutfits, 2),
			categoryInfo("empty", entities.CategoryStateEmpty, 0),
			categoryInfo("docs", entities.CategoryStateNoAvatarFiles, 0),
			categoryInfo("excluded", entities.CategoryStateUserExcluded, 1),
			categoryInfo("formal", entities.CategoryStateHasOutfits, 1),
		},
	}
	queries := newWardrobeQueries(config, entities.NewOutfitCache(), service)

	infos, err := queries.GetCategoryInfo()
	if err != nil {
		t.Fatalf("GetCategoryInfo() error = %v", err)
	}
	if !reflect.DeepEqual(infos, service.scanResult) {
		t.Fatalf("GetCategoryInfo() = %#v, want %#v", infos, service.scanResult)
	}
	if service.lastScanRootPath != wardrobeRoot {
		t.Fatalf("ScanCategories root = %q, want %q", service.lastScanRootPath, wardrobeRoot)
	}
	if !reflect.DeepEqual(service.lastExcludedCategories, map[string]bool{"excluded": true}) {
		t.Fatalf("ScanCategories excluded = %#v", service.lastExcludedCategories)
	}

	categories, err := queries.GetCategories()
	if err != nil {
		t.Fatalf("GetCategories() error = %v", err)
	}
	gotNames := categoryNames(categories)
	wantNames := []string{"casual", "formal"}
	if !reflect.DeepEqual(gotNames, wantNames) {
		t.Fatalf("GetCategories() names = %v, want %v", gotNames, wantNames)
	}
}

func TestWardrobeQueries_GetCategories_PropagatesCategoryInfoError(t *testing.T) {
	wantErr := errors.New("scan failed")
	queries := NewWardrobeQueries(
		&mockConfigUseCase{loadResult: mustWardrobeConfig(t, nil)},
		&mockCacheService{loadResult: newWardrobeCache()},
		&wardrobeCategoryService{scanError: wantErr},
	)

	got, err := queries.GetCategories()
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetCategories() error = %v, want %v", err, wantErr)
	}
	if got != nil {
		t.Fatalf("GetCategories() = %#v, want nil", got)
	}
}

func TestWardrobeQueries_GetRootDirectory(t *testing.T) {
	t.Run("returns configured root", func(t *testing.T) {
		queries := newWardrobeQueries(mustWardrobeConfig(t, nil), entities.NewOutfitCache(), nil)

		root, err := queries.GetRootDirectory()
		if err != nil {
			t.Fatalf("GetRootDirectory() error = %v", err)
		}
		if root != wardrobeRoot {
			t.Fatalf("GetRootDirectory() = %q, want %q", root, wardrobeRoot)
		}
	})

	t.Run("propagates configuration error", func(t *testing.T) {
		wantErr := errors.New("config failed")
		queries := NewWardrobeQueries(
			&mockConfigUseCase{loadError: wantErr},
			&mockCacheService{loadResult: newWardrobeCache()},
			&wardrobeCategoryService{},
		)

		root, err := queries.GetRootDirectory()
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetRootDirectory() error = %v, want %v", err, wantErr)
		}
		if root != "" {
			t.Fatalf("GetRootDirectory() = %q, want empty", root)
		}
	})
}

func TestWardrobeQueries_GetOutfitState(t *testing.T) {
	t.Run("builds state with worn and available outfits", func(t *testing.T) {
		cache := entities.NewOutfitCache()
		cache.Categories["casual"] = entities.CategoryCache{
			WornOutfits:  map[string]bool{"jeans.avatar": true},
			TotalOutfits: 3,
		}
		service := &wardrobeCategoryService{
			outfitsByPath: map[string][]entities.FileEntry{
				wardrobeCategoryPath("casual"): {
					{FileName: "jeans.avatar"},
					{FileName: "shirt.avatar"},
					{FileName: "boots.avatar"},
				},
			},
		}
		queries := newWardrobeQueries(mustWardrobeConfig(t, nil), cache, service)

		state, err := queries.GetOutfitState(categoryRefForWardrobe("casual"))
		if err != nil {
			t.Fatalf("GetOutfitState() error = %v", err)
		}
		if state.Category.Name != "casual" || state.Category.Path != wardrobeCategoryPath("casual") {
			t.Fatalf("state category = %#v", state.Category)
		}
		if state.TotalCount() != 3 || state.AvailableCount() != 2 || state.WornCount() != 1 {
			t.Fatalf("counts = total %d available %d worn %d", state.TotalCount(), state.AvailableCount(), state.WornCount())
		}
		if got := outfitNames(state.WornOutfits); !reflect.DeepEqual(got, []string{"jeans.avatar"}) {
			t.Fatalf("worn outfits = %v", got)
		}
		if got := outfitNames(state.AvailableOutfits); !reflect.DeepEqual(got, []string{"shirt.avatar", "boots.avatar"}) {
			t.Fatalf("available outfits = %v", got)
		}
		if !reflect.DeepEqual(service.outfitPaths, []string{wardrobeCategoryPath("casual")}) {
			t.Fatalf("GetOutfits paths = %v", service.outfitPaths)
		}
	})

	t.Run("uses empty cache when category is missing", func(t *testing.T) {
		service := &wardrobeCategoryService{
			outfitsByPath: map[string][]entities.FileEntry{
				wardrobeCategoryPath("new"): {
					{FileName: "one.avatar"},
				},
			},
		}
		queries := newWardrobeQueries(mustWardrobeConfig(t, nil), entities.NewOutfitCache(), service)

		state, err := queries.GetOutfitState(categoryRefForWardrobe("new"))
		if err != nil {
			t.Fatalf("GetOutfitState() error = %v", err)
		}
		if state.WornCount() != 0 || state.AvailableCount() != 1 {
			t.Fatalf("counts = available %d worn %d", state.AvailableCount(), state.WornCount())
		}
	})

	t.Run("propagates configuration error", func(t *testing.T) {
		wantErr := errors.New("config failed")
		queries := NewWardrobeQueries(
			&mockConfigUseCase{loadError: wantErr},
			&mockCacheService{loadResult: newWardrobeCache()},
			&wardrobeCategoryService{},
		)

		_, err := queries.GetOutfitState(categoryRefForWardrobe("casual"))
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetOutfitState() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("propagates cache error", func(t *testing.T) {
		wantErr := errors.New("cache failed")
		queries := NewWardrobeQueries(
			&mockConfigUseCase{loadResult: mustWardrobeConfig(t, nil)},
			&mockCacheService{loadError: wantErr},
			&wardrobeCategoryService{},
		)

		_, err := queries.GetOutfitState(categoryRefForWardrobe("casual"))
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetOutfitState() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("propagates outfit load error", func(t *testing.T) {
		wantErr := errors.New("outfits failed")
		queries := NewWardrobeQueries(
			&mockConfigUseCase{loadResult: mustWardrobeConfig(t, nil)},
			&mockCacheService{loadResult: newWardrobeCache()},
			&wardrobeCategoryService{outfitsError: wantErr},
		)

		_, err := queries.GetOutfitState(categoryRefForWardrobe("casual"))
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetOutfitState() error = %v, want %v", err, wantErr)
		}
	})
}

func TestWardrobeQueries_GetAllOutfitStates(t *testing.T) {
	t.Run("returns state for each category with outfits", func(t *testing.T) {
		service := &wardrobeCategoryService{
			scanResult: []entities.CategoryInfo{
				categoryInfo("casual", entities.CategoryStateHasOutfits, 1),
				categoryInfo("empty", entities.CategoryStateEmpty, 0),
				categoryInfo("formal", entities.CategoryStateHasOutfits, 1),
			},
			outfitsByPath: map[string][]entities.FileEntry{
				wardrobeCategoryPath("casual"): {{FileName: "one.avatar"}},
				wardrobeCategoryPath("formal"): {{FileName: "suit.avatar"}},
			},
		}
		queries := newWardrobeQueries(mustWardrobeConfig(t, nil), entities.NewOutfitCache(), service)

		states, err := queries.GetAllOutfitStates()
		if err != nil {
			t.Fatalf("GetAllOutfitStates() error = %v", err)
		}
		if len(states) != 2 {
			t.Fatalf("state count = %d, want 2", len(states))
		}
		if states["casual"].TotalCount() != 1 || states["formal"].TotalCount() != 1 {
			t.Fatalf("states = %#v", states)
		}
	})

	t.Run("propagates category list error", func(t *testing.T) {
		wantErr := errors.New("scan failed")
		queries := NewWardrobeQueries(
			&mockConfigUseCase{loadResult: mustWardrobeConfig(t, nil)},
			&mockCacheService{loadResult: newWardrobeCache()},
			&wardrobeCategoryService{scanError: wantErr},
		)

		states, err := queries.GetAllOutfitStates()
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetAllOutfitStates() error = %v, want %v", err, wantErr)
		}
		if states != nil {
			t.Fatalf("GetAllOutfitStates() = %#v, want nil", states)
		}
	})

	t.Run("propagates state error", func(t *testing.T) {
		wantErr := errors.New("formal failed")
		queries := NewWardrobeQueries(
			&mockConfigUseCase{loadResult: mustWardrobeConfig(t, nil)},
			&mockCacheService{loadResult: newWardrobeCache()},
			&wardrobeCategoryService{
				scanResult: []entities.CategoryInfo{
					categoryInfo("casual", entities.CategoryStateHasOutfits, 1),
					categoryInfo("formal", entities.CategoryStateHasOutfits, 1),
				},
				outfitsByPath: map[string][]entities.FileEntry{
					wardrobeCategoryPath("casual"): {{FileName: "one.avatar"}},
				},
				outfitErrorsByPath: map[string]error{
					wardrobeCategoryPath("formal"): wantErr,
				},
			},
		)

		states, err := queries.GetAllOutfitStates()
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetAllOutfitStates() error = %v, want %v", err, wantErr)
		}
		if states != nil {
			t.Fatalf("GetAllOutfitStates() = %#v, want nil", states)
		}
	})
}

func TestWardrobeQueries_GetAvailableOutfits(t *testing.T) {
	t.Run("returns available outfits from state", func(t *testing.T) {
		cache := entities.NewOutfitCache()
		cache.Categories["casual"] = entities.CategoryCache{
			WornOutfits:  map[string]bool{"one.avatar": true},
			TotalOutfits: 2,
		}
		queries := newWardrobeQueries(
			mustWardrobeConfig(t, nil),
			cache,
			&wardrobeCategoryService{
				outfitsByPath: map[string][]entities.FileEntry{
					wardrobeCategoryPath("casual"): {
						{FileName: "one.avatar"},
						{FileName: "two.avatar"},
					},
				},
			},
		)

		outfits, err := queries.GetAvailableOutfits(categoryRefForWardrobe("casual"))
		if err != nil {
			t.Fatalf("GetAvailableOutfits() error = %v", err)
		}
		if got := outfitNames(outfits); !reflect.DeepEqual(got, []string{"two.avatar"}) {
			t.Fatalf("GetAvailableOutfits() = %v, want [two.avatar]", got)
		}
	})

	t.Run("propagates state error", func(t *testing.T) {
		wantErr := errors.New("cache failed")
		queries := NewWardrobeQueries(
			&mockConfigUseCase{loadResult: mustWardrobeConfig(t, nil)},
			&mockCacheService{loadError: wantErr},
			&wardrobeCategoryService{},
		)

		outfits, err := queries.GetAvailableOutfits(categoryRefForWardrobe("casual"))
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetAvailableOutfits() error = %v, want %v", err, wantErr)
		}
		if outfits != nil {
			t.Fatalf("GetAvailableOutfits() = %#v, want nil", outfits)
		}
	})
}

func TestWardrobeQueries_ShowAllOutfits(t *testing.T) {
	t.Run("returns all outfits in category", func(t *testing.T) {
		service := &wardrobeCategoryService{
			outfitsByPath: map[string][]entities.FileEntry{
				wardrobeCategoryPath("casual"): {
					{FileName: "one.avatar"},
					{FileName: "two.avatar"},
				},
			},
		}
		queries := newWardrobeQueries(mustWardrobeConfig(t, nil), entities.NewOutfitCache(), service)

		outfits, err := queries.ShowAllOutfits("casual")
		if err != nil {
			t.Fatalf("ShowAllOutfits() error = %v", err)
		}
		if got := outfitNames(outfits); !reflect.DeepEqual(got, []string{"one.avatar", "two.avatar"}) {
			t.Fatalf("ShowAllOutfits() = %v, want [one.avatar two.avatar]", got)
		}
		if outfits[0].Category.Path != wardrobeCategoryPath("casual") {
			t.Fatalf("outfit category path = %q", outfits[0].Category.Path)
		}
	})

	t.Run("propagates configuration error", func(t *testing.T) {
		wantErr := errors.New("config failed")
		queries := NewWardrobeQueries(
			&mockConfigUseCase{loadError: wantErr},
			&mockCacheService{loadResult: newWardrobeCache()},
			&wardrobeCategoryService{},
		)

		outfits, err := queries.ShowAllOutfits("casual")
		if !errors.Is(err, wantErr) {
			t.Fatalf("ShowAllOutfits() error = %v, want %v", err, wantErr)
		}
		if outfits != nil {
			t.Fatalf("ShowAllOutfits() = %#v, want nil", outfits)
		}
	})

	t.Run("propagates outfit load error", func(t *testing.T) {
		wantErr := errors.New("outfits failed")
		queries := NewWardrobeQueries(
			&mockConfigUseCase{loadResult: mustWardrobeConfig(t, nil)},
			&mockCacheService{loadResult: newWardrobeCache()},
			&wardrobeCategoryService{outfitsError: wantErr},
		)

		outfits, err := queries.ShowAllOutfits("casual")
		if !errors.Is(err, wantErr) {
			t.Fatalf("ShowAllOutfits() error = %v, want %v", err, wantErr)
		}
		if outfits != nil {
			t.Fatalf("ShowAllOutfits() = %#v, want nil", outfits)
		}
	})
}

type wardrobeCategoryService struct {
	scanResult             []entities.CategoryInfo
	scanError              error
	lastScanRootPath       string
	lastExcludedCategories map[string]bool
	outfitsByPath          map[string][]entities.FileEntry
	outfitErrorsByPath     map[string]error
	outfitsError           error
	outfitPaths            []string
}

func (s *wardrobeCategoryService) ScanCategories(rootPath string, excludedCategories map[string]bool) ([]entities.CategoryInfo, error) {
	s.lastScanRootPath = rootPath
	if excludedCategories == nil {
		s.lastExcludedCategories = nil
	} else {
		s.lastExcludedCategories = make(map[string]bool, len(excludedCategories))
		for category, excluded := range excludedCategories {
			s.lastExcludedCategories[category] = excluded
		}
	}
	return s.scanResult, s.scanError
}

func (s *wardrobeCategoryService) GetOutfits(categoryPath string) ([]entities.FileEntry, error) {
	s.outfitPaths = append(s.outfitPaths, categoryPath)
	if err := s.outfitErrorsByPath[categoryPath]; err != nil {
		return nil, err
	}
	if s.outfitsError != nil {
		return nil, s.outfitsError
	}
	return s.outfitsByPath[categoryPath], nil
}

func newWardrobeQueries(config *entities.Config, cache entities.OutfitCache, service *wardrobeCategoryService) *WardrobeQueries {
	if service == nil {
		service = &wardrobeCategoryService{}
	}
	return NewWardrobeQueries(
		&mockConfigUseCase{loadResult: config},
		&mockCacheService{loadResult: &cache},
		service,
	)
}

func mustWardrobeConfig(t *testing.T, excluded map[string]bool) *entities.Config {
	t.Helper()
	config, err := entities.NewConfig(wardrobeRoot, nil, excluded, nil, nil)
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	return config
}

func newWardrobeCache() *entities.OutfitCache {
	cache := entities.NewOutfitCache()
	return &cache
}

func categoryInfo(name string, state entities.CategoryState, count int) entities.CategoryInfo {
	return entities.NewCategoryInfo(categoryRefForWardrobe(name), state, count)
}

func categoryRefForWardrobe(name string) entities.CategoryReference {
	return entities.NewCategoryReference(name, wardrobeCategoryPath(name))
}

func wardrobeCategoryPath(name string) string {
	return filepath.Join(wardrobeRoot, name)
}

func categoryNames(categories []entities.CategoryReference) []string {
	names := make([]string, 0, len(categories))
	for _, category := range categories {
		names = append(names, category.Name)
	}
	return names
}

func outfitNames(outfits []entities.OutfitReference) []string {
	names := make([]string, 0, len(outfits))
	for _, outfit := range outfits {
		names = append(names, outfit.FileName)
	}
	return names
}
