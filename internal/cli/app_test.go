package cli

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/dh85/outfitpicker/internal/application/usecases"
	"github.com/dh85/outfitpicker/internal/domain/entities"
	domainerrors "github.com/dh85/outfitpicker/internal/domain/errors"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
)

func TestApplication_GetCategories_FiltersOnlyCategoriesWithOutfits(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
	categorySvc := &stubCategoryService{
		scanCategoriesResult: []entities.CategoryInfo{
			entities.NewCategoryInfo(entities.NewCategoryReference("casual", cliTestCategoryPath("casual")), entities.CategoryStateHasOutfits, 2),
			entities.NewCategoryInfo(entities.NewCategoryReference("empty", cliTestCategoryPath("empty")), entities.CategoryStateEmpty, 0),
			entities.NewCategoryInfo(entities.NewCategoryReference("excluded", cliTestCategoryPath("excluded")), entities.CategoryStateUserExcluded, 1),
		},
	}
	app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, categorySvc)

	categories, err := app.GetCategories()
	if err != nil {
		t.Fatalf("GetCategories() error = %v", err)
	}
	if len(categories) != 1 || categories[0].Name != "casual" {
		t.Fatalf("GetCategories() = %#v, want only casual", categories)
	}
}

func TestApplication_GetCategories_PropagatesCategoryInfoError(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
	wantErr := errors.New("scan failed")
	app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{scanCategoriesErr: wantErr})

	_, err := app.GetCategories()
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetCategories() error = %v, want %v", err, wantErr)
	}
}

func TestApplication_GetRootDirectory_ReturnsConfiguredRoot(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
	app := newTestApplication(nil, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{})

	root, err := app.GetRootDirectory()
	if err != nil {
		t.Fatalf("GetRootDirectory() error = %v", err)
	}
	if root != cliTestOutfitRoot {
		t.Fatalf("GetRootDirectory() = %q, want %q", root, cliTestOutfitRoot)
	}
}

func TestApplication_GetRootDirectory_PropagatesConfigurationError(t *testing.T) {
	wantErr := errors.New("config load failed")
	app := newTestApplication(nil, &stubConfigManager{err: wantErr}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{})

	_, err := app.GetRootDirectory()
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetRootDirectory() error = %v, want %v", err, wantErr)
	}
}

func TestApplication_GetConfiguration_ReturnsConfigurationNotFoundForNilConfig(t *testing.T) {
	app := newTestApplication(nil, &stubConfigManager{}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{})

	_, err := app.GetConfiguration()
	if !errors.Is(err, domainerrors.ErrConfigurationNotFound) {
		t.Fatalf("GetConfiguration() error = %v, want %v", err, domainerrors.ErrConfigurationNotFound)
	}
}

func TestApplication_GetConfiguration_PropagatesLoadError(t *testing.T) {
	wantErr := errors.New("load failed")
	app := newTestApplication(nil, &stubConfigManager{err: wantErr}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{})

	_, err := app.GetConfiguration()
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetConfiguration() error = %v, want %v", err, wantErr)
	}
}

func TestApplication_ShowNextUniqueRandomOutfit_SkipsExcludedCategories(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), map[string]bool{"formal": true}, nil, nil)
	categorySvc := &stubCategoryService{
		scanCategoriesResult: []entities.CategoryInfo{
			entities.NewCategoryInfo(entities.NewCategoryReference("casual", "/root/outfits/casual"), entities.CategoryStateHasOutfits, 1),
			entities.NewCategoryInfo(entities.NewCategoryReference("formal", "/root/outfits/formal"), entities.CategoryStateHasOutfits, 1),
		},
		outfitsByPath: map[string][]entities.FileEntry{
			cliTestCategoryPath("casual"): {{FileName: "casual.avatar"}},
			cliTestCategoryPath("formal"): {{FileName: "formal.avatar"}},
		},
	}
	app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, categorySvc)
	app.randomInt = func(int) int { return 0 }

	outfit, err := app.ShowNextUniqueRandomOutfit()
	if err != nil {
		t.Fatalf("ShowNextUniqueRandomOutfit() error = %v", err)
	}
	if outfit == nil {
		t.Fatal("ShowNextUniqueRandomOutfit() returned nil outfit")
	}
	if outfit.Category.Name != "casual" {
		t.Fatalf("selected category = %q, want casual", outfit.Category.Name)
	}
}

func TestApplication_ShowNextUniqueRandomOutfitFrom_ResetsShownSession(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
	categorySvc := &stubCategoryService{
		scanCategoriesResult: []entities.CategoryInfo{
			entities.NewCategoryInfo(entities.NewCategoryReference("casual", "/root/outfits/casual"), entities.CategoryStateHasOutfits, 2),
		},
		outfitsByPath: map[string][]entities.FileEntry{
			cliTestCategoryPath("casual"): {
				{FileName: "outfit1.avatar"},
				{FileName: "outfit2.avatar"},
			},
		},
	}
	app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, categorySvc)
	app.randomInt = func(int) int { return 0 }

	first, err := app.ShowNextUniqueRandomOutfitFrom("casual")
	if err != nil {
		t.Fatalf("first ShowNextUniqueRandomOutfitFrom() error = %v", err)
	}
	second, err := app.ShowNextUniqueRandomOutfitFrom("casual")
	if err != nil {
		t.Fatalf("second ShowNextUniqueRandomOutfitFrom() error = %v", err)
	}
	third, err := app.ShowNextUniqueRandomOutfitFrom("casual")
	if err != nil {
		t.Fatalf("third ShowNextUniqueRandomOutfitFrom() error = %v", err)
	}

	if first == nil || second == nil || third == nil {
		t.Fatal("expected non-nil outfits on all calls")
	}
	if first.FileName == second.FileName {
		t.Fatalf("expected second outfit to differ from first, got %q and %q", first.FileName, second.FileName)
	}
	if third.FileName != first.FileName {
		t.Fatalf("expected third outfit to reset shown session and return %q, got %q", first.FileName, third.FileName)
	}
}

func TestApplication_UpdateConfiguration_ResetsCacheWhenRootChanges(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
	updated, _ := entities.NewConfig(cliTestOtherOutfitRoot, stringPtr("en"), nil, nil, nil)
	cacheManager := &stubCacheManager{cache: newOutfitCachePtr()}
	app := newTestApplication(config, &stubConfigManager{config: config}, cacheManager, &stubCategoryService{})
	markGlobalShown(app, "casual/one.avatar")
	markCategoryShown(app, "casual", "one.avatar")

	err := app.UpdateConfiguration(updated)
	if err != nil {
		t.Fatalf("UpdateConfiguration() error = %v", err)
	}
	if cacheManager.deleteCalls != 1 {
		t.Fatalf("cache deleteCalls = %d, want 1", cacheManager.deleteCalls)
	}
	if globalShownCount(app) != 0 {
		t.Fatalf("expected global shown cache to reset, got %d entries", globalShownCount(app))
	}
	if trackedCategoryCount(app) != 0 {
		t.Fatalf("expected category shown cache to reset, got %d entries", trackedCategoryCount(app))
	}
	gotConfig, err := app.GetConfiguration()
	if err != nil {
		t.Fatalf("GetConfiguration() error = %v", err)
	}
	if gotConfig.Root != cliTestOtherOutfitRoot {
		t.Fatalf("config root = %q, want %q", gotConfig.Root, cliTestOtherOutfitRoot)
	}
}

func TestApplication_UpdateConfiguration_DoesNotResetCacheWhenRootUnchanged(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
	updated, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("fr"), nil, nil, nil)
	cacheManager := &stubCacheManager{cache: newOutfitCachePtr()}
	app := newTestApplication(config, &stubConfigManager{config: config}, cacheManager, &stubCategoryService{})

	err := app.UpdateConfiguration(updated)
	if err != nil {
		t.Fatalf("UpdateConfiguration() error = %v", err)
	}
	if cacheManager.deleteCalls != 0 {
		t.Fatalf("cache deleteCalls = %d, want 0", cacheManager.deleteCalls)
	}
}

func TestApplication_UpdateConfiguration_PropagatesSaveAndDeleteErrors(t *testing.T) {
	t.Run("save error", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		updated, _ := entities.NewConfig(cliTestOtherOutfitRoot, stringPtr("en"), nil, nil, nil)
		wantErr := errors.New("save failed")
		configManager := &stubConfigManager{config: config, saveErr: wantErr}
		cacheManager := &stubCacheManager{cache: newOutfitCachePtr()}
		app := newTestApplication(config, configManager, cacheManager, &stubCategoryService{})

		err := app.UpdateConfiguration(updated)
		if !errors.Is(err, wantErr) {
			t.Fatalf("UpdateConfiguration() error = %v, want %v", err, wantErr)
		}
		if cacheManager.deleteCalls != 0 {
			t.Fatalf("cache deleteCalls = %d, want 0", cacheManager.deleteCalls)
		}
	})

	t.Run("delete error on root change", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		updated, _ := entities.NewConfig(cliTestOtherOutfitRoot, stringPtr("en"), nil, nil, nil)
		wantErr := errors.New("delete failed")
		cacheManager := &stubCacheManager{cache: newOutfitCachePtr(), deleteErr: wantErr}
		app := newTestApplication(config, &stubConfigManager{config: config}, cacheManager, &stubCategoryService{})

		err := app.UpdateConfiguration(updated)
		if !errors.Is(err, wantErr) {
			t.Fatalf("UpdateConfiguration() error = %v, want %v", err, wantErr)
		}
		if cacheManager.deleteCalls != 1 {
			t.Fatalf("cache deleteCalls = %d, want 1", cacheManager.deleteCalls)
		}
	})
}

func TestApplication_FactoryReset_DeletesConfigCacheAndResetsState(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
	configManager := &stubConfigManager{config: config}
	cacheManager := &stubCacheManager{cache: newOutfitCachePtr()}
	app := newTestApplication(config, configManager, cacheManager, &stubCategoryService{})
	markGlobalShown(app, "casual/one.avatar")
	markCategoryShown(app, "casual", "one.avatar")

	err := app.FactoryReset()
	if err != nil {
		t.Fatalf("FactoryReset() error = %v", err)
	}
	if configManager.deleteCalls != 1 {
		t.Fatalf("config deleteCalls = %d, want 1", configManager.deleteCalls)
	}
	if cacheManager.deleteCalls != 1 {
		t.Fatalf("cache deleteCalls = %d, want 1", cacheManager.deleteCalls)
	}
	if globalShownCount(app) != 0 || trackedCategoryCount(app) != 0 {
		t.Fatalf("expected session tracking reset, got global=%d category=%d", globalShownCount(app), trackedCategoryCount(app))
	}
}

func TestApplication_FactoryReset_PropagatesDeleteErrors(t *testing.T) {
	t.Run("config delete error", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		wantErr := errors.New("config delete failed")
		configManager := &stubConfigManager{config: config, deleteErr: wantErr}
		cacheManager := &stubCacheManager{cache: newOutfitCachePtr()}
		app := newTestApplication(config, configManager, cacheManager, &stubCategoryService{})

		err := app.FactoryReset()
		if !errors.Is(err, wantErr) {
			t.Fatalf("FactoryReset() error = %v, want %v", err, wantErr)
		}
		if cacheManager.deleteCalls != 0 {
			t.Fatalf("cache deleteCalls = %d, want 0", cacheManager.deleteCalls)
		}
	})

	t.Run("cache delete error", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		wantErr := errors.New("cache delete failed")
		configManager := &stubConfigManager{config: config}
		cacheManager := &stubCacheManager{cache: newOutfitCachePtr(), deleteErr: wantErr}
		app := newTestApplication(config, configManager, cacheManager, &stubCategoryService{})

		err := app.FactoryReset()
		if !errors.Is(err, wantErr) {
			t.Fatalf("FactoryReset() error = %v, want %v", err, wantErr)
		}
	})
}

func TestApplication_GetOutfitState_BuildsStateFromConfigCacheAndFiles(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
	cache := entities.NewOutfitCache()
	cache.Categories["casual"] = entities.CategoryCache{
		WornOutfits:  map[string]bool{"worn.avatar": true},
		TotalOutfits: 2,
	}
	app := newTestApplication(
		config,
		&stubConfigManager{config: config},
		&stubCacheManager{cache: &cache},
		&stubCategoryService{outfitsByPath: map[string][]entities.FileEntry{
			cliTestCategoryPath("casual"): {
				{FileName: "worn.avatar"},
				{FileName: "fresh.avatar"},
			},
		}},
	)

	state, err := app.GetOutfitState(entities.NewCategoryReference("casual", ""))
	if err != nil {
		t.Fatalf("GetOutfitState() error = %v", err)
	}
	if state.Category.Name != "casual" {
		t.Fatalf("state category = %q, want casual", state.Category.Name)
	}
	if len(state.AllOutfits) != 2 || len(state.AvailableOutfits) != 1 || len(state.WornOutfits) != 1 {
		t.Fatalf("unexpected state counts: all=%d available=%d worn=%d", len(state.AllOutfits), len(state.AvailableOutfits), len(state.WornOutfits))
	}
	if state.WornOutfits[0].FileName != "worn.avatar" || state.AvailableOutfits[0].FileName != "fresh.avatar" {
		t.Fatalf("unexpected state contents: %#v", state)
	}
}

func TestApplication_GetOutfitState_PropagatesErrors(t *testing.T) {
	t.Run("configuration error", func(t *testing.T) {
		wantErr := errors.New("config load failed")
		app := newTestApplication(nil, &stubConfigManager{err: wantErr}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{})

		_, err := app.GetOutfitState(entities.NewCategoryReference("casual", ""))
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetOutfitState() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("cache error", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		wantErr := errors.New("cache load failed")
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{err: wantErr}, &stubCategoryService{})

		_, err := app.GetOutfitState(entities.NewCategoryReference("casual", ""))
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetOutfitState() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("category service error", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		wantErr := errors.New("get outfits failed")
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{outfitsErr: wantErr})

		_, err := app.GetOutfitState(entities.NewCategoryReference("casual", ""))
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetOutfitState() error = %v, want %v", err, wantErr)
		}
	})
}

func TestApplication_GetAllOutfitStates_ReturnsStatesForAllCategories(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
	categorySvc := &stubCategoryService{
		scanCategoriesResult: []entities.CategoryInfo{
			entities.NewCategoryInfo(entities.NewCategoryReference("casual", cliTestCategoryPath("casual")), entities.CategoryStateHasOutfits, 2),
			entities.NewCategoryInfo(entities.NewCategoryReference("formal", cliTestCategoryPath("formal")), entities.CategoryStateHasOutfits, 1),
		},
		outfitsByPath: map[string][]entities.FileEntry{
			cliTestCategoryPath("casual"): {{FileName: "one.avatar"}, {FileName: "two.avatar"}},
			cliTestCategoryPath("formal"): {{FileName: "jacket.avatar"}},
		},
	}
	app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, categorySvc)

	states, err := app.GetAllOutfitStates()
	if err != nil {
		t.Fatalf("GetAllOutfitStates() error = %v", err)
	}
	if len(states) != 2 {
		t.Fatalf("len(states) = %d, want 2", len(states))
	}
	if states["casual"].TotalCount() != 2 || states["formal"].TotalCount() != 1 {
		t.Fatalf("unexpected state totals: %#v", states)
	}
}

func TestApplication_GetAllOutfitStates_PropagatesErrors(t *testing.T) {
	t.Run("get categories error", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		wantErr := errors.New("scan failed")
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{scanCategoriesErr: wantErr})

		_, err := app.GetAllOutfitStates()
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetAllOutfitStates() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("state error", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		categorySvc := &stubCategoryService{
			scanCategoriesResult: []entities.CategoryInfo{
				entities.NewCategoryInfo(entities.NewCategoryReference("casual", cliTestCategoryPath("casual")), entities.CategoryStateHasOutfits, 1),
			},
			outfitsErr: errors.New("get outfits failed"),
		}
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, categorySvc)

		_, err := app.GetAllOutfitStates()
		if !errors.Is(err, categorySvc.outfitsErr) {
			t.Fatalf("GetAllOutfitStates() error = %v, want %v", err, categorySvc.outfitsErr)
		}
	})
}

func TestApplication_GetAvailableOutfits_ReturnsAvailableOutfits(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
	cache := entities.NewOutfitCache()
	cache.Categories["casual"] = entities.CategoryCache{WornOutfits: map[string]bool{"worn.avatar": true}, TotalOutfits: 2}
	app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: &cache}, &stubCategoryService{outfitsByPath: map[string][]entities.FileEntry{
		cliTestCategoryPath("casual"): {{FileName: "worn.avatar"}, {FileName: "fresh.avatar"}},
	}})

	outfits, err := app.GetAvailableOutfits(entities.NewCategoryReference("casual", ""))
	if err != nil {
		t.Fatalf("GetAvailableOutfits() error = %v", err)
	}
	if len(outfits) != 1 || outfits[0].FileName != "fresh.avatar" {
		t.Fatalf("GetAvailableOutfits() = %#v, want only fresh.avatar", outfits)
	}
}

func TestApplication_GetAvailableOutfits_PropagatesStateError(t *testing.T) {
	wantErr := errors.New("config load failed")
	app := newTestApplication(nil, &stubConfigManager{err: wantErr}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{})

	_, err := app.GetAvailableOutfits(entities.NewCategoryReference("casual", ""))
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetAvailableOutfits() error = %v, want %v", err, wantErr)
	}
}

func TestApplication_ShowAllOutfits_ReturnsAllOutfitsAndErrors(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{outfitsByPath: map[string][]entities.FileEntry{
			cliTestCategoryPath("casual"): {{FileName: "one.avatar"}, {FileName: "two.avatar"}},
		}})

		outfits, err := app.ShowAllOutfits("casual")
		if err != nil {
			t.Fatalf("ShowAllOutfits() error = %v", err)
		}
		if len(outfits) != 2 || outfits[0].Category.Name != "casual" {
			t.Fatalf("ShowAllOutfits() = %#v, want 2 casual outfits", outfits)
		}
	})

	t.Run("configuration error", func(t *testing.T) {
		wantErr := errors.New("config load failed")
		app := newTestApplication(nil, &stubConfigManager{err: wantErr}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{})

		_, err := app.ShowAllOutfits("casual")
		if !errors.Is(err, wantErr) {
			t.Fatalf("ShowAllOutfits() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("category service error", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		wantErr := errors.New("get outfits failed")
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{outfitsErr: wantErr})

		_, err := app.ShowAllOutfits("casual")
		if !errors.Is(err, wantErr) {
			t.Fatalf("ShowAllOutfits() error = %v, want %v", err, wantErr)
		}
	})
}

func TestApplication_WearOutfit_ResetsSessionOnSuccessAndRotationCompletion(t *testing.T) {
	outfit := entities.NewOutfitReference("one.avatar", entities.NewCategoryReference("casual", cliTestCategoryPath("casual")))

	t.Run("success", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		cache := entities.NewOutfitCache()
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: &cache}, &stubCategoryService{outfitsByPath: map[string][]entities.FileEntry{
			cliTestCategoryPath("casual"): {{FileName: "one.avatar"}, {FileName: "two.avatar"}},
		}})
		markGlobalShown(app, outfitKey(outfit))
		markCategoryShown(app, "casual", "one.avatar", "two.avatar")

		err := app.WearOutfit(outfit)
		if err != nil {
			t.Fatalf("WearOutfit() error = %v", err)
		}
		if globalShownCount(app) != 0 {
			t.Fatalf("expected global shown reset, got %d entries", globalShownCount(app))
		}
		if categoryShownCount(app, "casual") != 0 {
			t.Fatal("expected category shown entry to be removed")
		}
	})

	t.Run("rotation complete", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		cache := entities.NewOutfitCache()
		cache.Categories["casual"] = entities.CategoryCache{WornOutfits: map[string]bool{"two.avatar": true}, TotalOutfits: 2}
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: &cache}, &stubCategoryService{outfitsByPath: map[string][]entities.FileEntry{
			cliTestCategoryPath("casual"): {{FileName: "one.avatar"}, {FileName: "two.avatar"}},
		}})
		markGlobalShown(app, outfitKey(outfit))
		markCategoryShown(app, "casual", "one.avatar")

		err := app.WearOutfit(outfit)
		if !isRotationCompleteError(err) {
			t.Fatalf("WearOutfit() error = %v, want rotation complete", err)
		}
		if globalShownCount(app) != 0 {
			t.Fatalf("expected global shown reset, got %d entries", globalShownCount(app))
		}
		if categoryShownCount(app, "casual") != 0 {
			t.Fatal("expected category shown entry to be removed")
		}
	})
}

func TestApplication_WearOutfit_DoesNotResetSessionForNonRotationError(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
	outfit := entities.NewOutfitReference("one.avatar", entities.NewCategoryReference("casual", cliTestCategoryPath("casual")))
	wantErr := errors.New("get outfits failed")
	app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{outfitsErr: wantErr})
	markGlobalShown(app, outfitKey(outfit))
	markCategoryShown(app, "casual", "one.avatar")

	err := app.WearOutfit(outfit)
	if !errors.Is(err, wantErr) {
		t.Fatalf("WearOutfit() error = %v, want %v", err, wantErr)
	}
	if globalShownCount(app) != 1 {
		t.Fatalf("expected global shown to remain, got %d entries", globalShownCount(app))
	}
	if categoryShownCount(app, "casual") != 1 {
		t.Fatalf("expected category shown to remain, got %d entries", categoryShownCount(app, "casual"))
	}
}

func TestApplication_ResetCategory_DeletesCategorySessionAndErrors(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		cache := entities.NewOutfitCache()
		cache.Categories["casual"] = entities.CategoryCache{WornOutfits: map[string]bool{"one.avatar": true}, TotalOutfits: 1}
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: &cache}, &stubCategoryService{})
		markCategoryShown(app, "casual", "one.avatar")

		err := app.ResetCategory("casual")
		if err != nil {
			t.Fatalf("ResetCategory() error = %v", err)
		}
		if categoryShownCount(app, "casual") != 0 {
			t.Fatal("expected category session tracking to be removed")
		}
	})

	t.Run("use case error", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		wantErr := errors.New("cache load failed")
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{err: wantErr}, &stubCategoryService{})
		markCategoryShown(app, "casual", "one.avatar")

		err := app.ResetCategory("casual")
		if !errors.Is(err, wantErr) {
			t.Fatalf("ResetCategory() error = %v, want %v", err, wantErr)
		}
		if categoryShownCount(app, "casual") == 0 {
			t.Fatal("expected category session tracking to remain on error")
		}
	})
}

func TestApplication_ResetAllCategories_ResetsAllSessionTrackingAndErrors(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		cache := entities.NewOutfitCache()
		cache.Categories["casual"] = entities.CategoryCache{WornOutfits: map[string]bool{"one.avatar": true}, TotalOutfits: 1}
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: &cache}, &stubCategoryService{})
		markGlobalShown(app, "casual/one.avatar")
		markCategoryShown(app, "casual", "one.avatar")

		err := app.ResetAllCategories()
		if err != nil {
			t.Fatalf("ResetAllCategories() error = %v", err)
		}
		if globalShownCount(app) != 0 || trackedCategoryCount(app) != 0 {
			t.Fatalf("expected session tracking reset, got global=%d category=%d", globalShownCount(app), trackedCategoryCount(app))
		}
	})

	t.Run("use case error", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		wantErr := errors.New("save failed")
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{saveErr: wantErr}, &stubCategoryService{})
		markGlobalShown(app, "casual/one.avatar")
		markCategoryShown(app, "casual", "one.avatar")

		err := app.ResetAllCategories()
		if !errors.Is(err, wantErr) {
			t.Fatalf("ResetAllCategories() error = %v, want %v", err, wantErr)
		}
		if globalShownCount(app) != 1 || trackedCategoryCount(app) != 1 {
			t.Fatalf("expected session tracking unchanged on error, got global=%d category=%d", globalShownCount(app), trackedCategoryCount(app))
		}
	})
}

func TestApplication_ShowNextUniqueRandomOutfit_Branches(t *testing.T) {
	t.Run("configuration error", func(t *testing.T) {
		wantErr := errors.New("config load failed")
		app := newTestApplication(nil, &stubConfigManager{err: wantErr}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{})

		_, err := app.ShowNextUniqueRandomOutfit()
		if !errors.Is(err, wantErr) {
			t.Fatalf("ShowNextUniqueRandomOutfit() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("category info error", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		wantErr := errors.New("scan failed")
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{scanCategoriesErr: wantErr})

		_, err := app.ShowNextUniqueRandomOutfit()
		if !errors.Is(err, wantErr) {
			t.Fatalf("ShowNextUniqueRandomOutfit() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("state error", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		categorySvc := &stubCategoryService{
			scanCategoriesResult: []entities.CategoryInfo{
				entities.NewCategoryInfo(entities.NewCategoryReference("casual", cliTestCategoryPath("casual")), entities.CategoryStateHasOutfits, 1),
			},
			outfitsErr: errors.New("get outfits failed"),
		}
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, categorySvc)

		_, err := app.ShowNextUniqueRandomOutfit()
		if !errors.Is(err, categorySvc.outfitsErr) {
			t.Fatalf("ShowNextUniqueRandomOutfit() error = %v, want %v", err, categorySvc.outfitsErr)
		}
	})

	t.Run("no available outfits returns nil", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		categorySvc := &stubCategoryService{
			scanCategoriesResult: []entities.CategoryInfo{
				entities.NewCategoryInfo(entities.NewCategoryReference("empty", cliTestCategoryPath("empty")), entities.CategoryStateEmpty, 0),
			},
		}
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, categorySvc)

		outfit, err := app.ShowNextUniqueRandomOutfit()
		if err != nil {
			t.Fatalf("ShowNextUniqueRandomOutfit() error = %v", err)
		}
		if outfit != nil {
			t.Fatalf("ShowNextUniqueRandomOutfit() = %#v, want nil", outfit)
		}
	})

	t.Run("single available outfit does not call randomInt", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		categorySvc := &stubCategoryService{
			scanCategoriesResult: []entities.CategoryInfo{
				entities.NewCategoryInfo(entities.NewCategoryReference("casual", cliTestCategoryPath("casual")), entities.CategoryStateHasOutfits, 1),
			},
			outfitsByPath: map[string][]entities.FileEntry{
				cliTestCategoryPath("casual"): {{FileName: "one.avatar"}},
			},
		}
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, categorySvc)
		app.randomInt = func(int) int {
			t.Fatal("randomInt should not be called when only one outfit is available")
			return 0
		}

		outfit, err := app.ShowNextUniqueRandomOutfit()
		if err != nil {
			t.Fatalf("ShowNextUniqueRandomOutfit() error = %v", err)
		}
		if outfit == nil || outfit.FileName != "one.avatar" {
			t.Fatalf("ShowNextUniqueRandomOutfit() = %#v, want one.avatar", outfit)
		}
	})

	t.Run("resets global shown when all have been seen", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		categorySvc := &stubCategoryService{
			scanCategoriesResult: []entities.CategoryInfo{
				entities.NewCategoryInfo(entities.NewCategoryReference("casual", cliTestCategoryPath("casual")), entities.CategoryStateHasOutfits, 2),
			},
			outfitsByPath: map[string][]entities.FileEntry{
				cliTestCategoryPath("casual"): {{FileName: "one.avatar"}, {FileName: "two.avatar"}},
			},
		}
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, categorySvc)
		markGlobalShown(app, outfitKey(entities.NewOutfitReference("one.avatar", entities.NewCategoryReference("casual", cliTestCategoryPath("casual")))))
		markGlobalShown(app, outfitKey(entities.NewOutfitReference("two.avatar", entities.NewCategoryReference("casual", cliTestCategoryPath("casual")))))
		app.randomInt = func(int) int { return 1 }

		outfit, err := app.ShowNextUniqueRandomOutfit()
		if err != nil {
			t.Fatalf("ShowNextUniqueRandomOutfit() error = %v", err)
		}
		if outfit == nil || outfit.FileName != "two.avatar" {
			t.Fatalf("ShowNextUniqueRandomOutfit() = %#v, want two.avatar", outfit)
		}
		if !isGlobalShown(app, outfitKey(*outfit)) {
			t.Fatal("expected selected outfit to be marked as shown")
		}
	})
}

func TestApplication_ShowNextUniqueRandomOutfitFrom_ReturnsNilForNoAvailableOutfits(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
	cache := entities.NewOutfitCache()
	cache.Categories["casual"] = entities.CategoryCache{WornOutfits: map[string]bool{"one.avatar": true}, TotalOutfits: 1}
	app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: &cache}, &stubCategoryService{outfitsByPath: map[string][]entities.FileEntry{
		cliTestCategoryPath("casual"): {{FileName: "one.avatar"}},
	}})

	outfit, err := app.ShowNextUniqueRandomOutfitFrom("casual")
	if err != nil {
		t.Fatalf("ShowNextUniqueRandomOutfitFrom() error = %v", err)
	}
	if outfit != nil {
		t.Fatalf("ShowNextUniqueRandomOutfitFrom() = %#v, want nil", outfit)
	}
}

func TestApplication_ShowNextUniqueRandomOutfitFrom_PropagatesStateError(t *testing.T) {
	wantErr := errors.New("config load failed")
	app := newTestApplication(nil, &stubConfigManager{err: wantErr}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{})

	_, err := app.ShowNextUniqueRandomOutfitFrom("casual")
	if !errors.Is(err, wantErr) {
		t.Fatalf("ShowNextUniqueRandomOutfitFrom() error = %v, want %v", err, wantErr)
	}
}

func TestApplication_HelperFunctions(t *testing.T) {
	t.Run("sortedCategoryNames", func(t *testing.T) {
		got := sortedCategoryNames(map[string][]entities.OutfitReference{
			"formal": nil,
			"casual": nil,
			"sport":  nil,
		})
		want := []string{"casual", "formal", "sport"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("sortedCategoryNames() = %#v, want %#v", got, want)
		}
	})

	t.Run("currentCategoryWornFileNames", func(t *testing.T) {
		state := entities.NewCategoryOutfitState(
			entities.NewCategoryReference("casual", cliTestCategoryPath("casual")),
			nil,
			nil,
			[]entities.OutfitReference{
				entities.NewOutfitReference("one.avatar", entities.NewCategoryReference("casual", cliTestCategoryPath("casual"))),
				entities.NewOutfitReference("two.avatar", entities.NewCategoryReference("casual", cliTestCategoryPath("casual"))),
			},
		)

		got := currentCategoryWornFileNames(state)
		want := map[string]bool{"one.avatar": true, "two.avatar": true}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("currentCategoryWornFileNames() = %#v, want %#v", got, want)
		}
	})

	t.Run("isRotationCompleteError", func(t *testing.T) {
		if !isRotationCompleteError(domainerrors.NewRotationCompletedError("casual")) {
			t.Fatal("expected rotation complete error to be detected")
		}
		if isRotationCompleteError(errors.New("other")) {
			t.Fatal("did not expect non-rotation error to be detected")
		}
	})

	t.Run("availableOutfitsFromState sorts copy", func(t *testing.T) {
		state := entities.NewCategoryOutfitState(
			entities.NewCategoryReference("casual", cliTestCategoryPath("casual")),
			nil,
			[]entities.OutfitReference{
				entities.NewOutfitReference("z.avatar", entities.NewCategoryReference("casual", cliTestCategoryPath("casual"))),
				entities.NewOutfitReference("a.avatar", entities.NewCategoryReference("casual", cliTestCategoryPath("casual"))),
			},
			nil,
		)

		got := availableOutfitsFromState(state)
		if got[0].FileName != "a.avatar" || got[1].FileName != "z.avatar" {
			t.Fatalf("availableOutfitsFromState() = %#v, want sorted order", got)
		}
		if state.AvailableOutfits[0].FileName != "z.avatar" {
			t.Fatal("expected original state slice to remain unchanged")
		}
	})

	t.Run("allOutfitsFromFiles sorts outputs", func(t *testing.T) {
		category := entities.NewCategoryReference("casual", cliTestCategoryPath("casual"))
		got := allOutfitsFromFiles(category, []entities.FileEntry{{FileName: "z.avatar"}, {FileName: "a.avatar"}})
		if got[0].FileName != "a.avatar" || got[1].FileName != "z.avatar" {
			t.Fatalf("allOutfitsFromFiles() = %#v, want sorted order", got)
		}
	})

	t.Run("filterAvailableFiles excludes worn", func(t *testing.T) {
		got := filterAvailableFiles([]entities.FileEntry{{FileName: "one.avatar"}, {FileName: "two.avatar"}}, map[string]bool{"one.avatar": true})
		if len(got) != 1 || got[0].FileName != "two.avatar" {
			t.Fatalf("filterAvailableFiles() = %#v, want only two.avatar", got)
		}
	})

	t.Run("resetAfterWear clears session and category entry", func(t *testing.T) {
		config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
		app := newTestApplication(config, &stubConfigManager{config: config}, &stubCacheManager{cache: newOutfitCachePtr()}, &stubCategoryService{})
		markGlobalShown(app, "casual/one.avatar")
		markCategoryShown(app, "casual", "one.avatar")
		markCategoryShown(app, "formal", "jacket.avatar")

		app.resetAfterWear("casual")

		if globalShownCount(app) != 0 {
			t.Fatalf("expected global shown reset, got %d entries", globalShownCount(app))
		}
		if categoryShownCount(app, "casual") != 0 {
			t.Fatal("expected casual session to be removed")
		}
		if trackedCategoryCount(app) != 0 {
			t.Fatalf("expected all category tracking reset, got %d tracked categories", trackedCategoryCount(app))
		}
	})

	t.Run("filter helpers remain sorted as expected", func(t *testing.T) {
		got := []string{"b", "a"}
		sort.Strings(got)
		if !reflect.DeepEqual(got, []string{"a", "b"}) {
			t.Fatalf("sort sanity check failed: %#v", got)
		}
	})
}

type stubConfigManager struct {
	config      *entities.Config
	err         error
	saveErr     error
	deleteErr   error
	deleteCalls int
}

func (s *stubConfigManager) LoadOrCreate() (*entities.Config, error) { return s.config, s.err }
func (s *stubConfigManager) Save(config *entities.Config) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.config = config
	return nil
}
func (s *stubConfigManager) Delete() error {
	s.deleteCalls++
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.config = nil
	return nil
}

type stubCacheManager struct {
	cache       *entities.OutfitCache
	err         error
	saveErr     error
	deleteErr   error
	deleteCalls int
}

func (s *stubCacheManager) LoadOrCreate() (*entities.OutfitCache, error) {
	if s.cache == nil {
		s.cache = newOutfitCachePtr()
	}
	return s.cache, s.err
}
func (s *stubCacheManager) Save(cache *entities.OutfitCache) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.cache = cache
	return nil
}
func (s *stubCacheManager) Delete() error {
	s.deleteCalls++
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.cache = newOutfitCachePtr()
	return nil
}

type stubCategoryService struct {
	scanCategoriesResult []entities.CategoryInfo
	scanCategoriesErr    error
	outfitsByPath        map[string][]entities.FileEntry
	outfitsErr           error
}

var _ interfaces.CategoryService = (*stubCategoryService)(nil)

func (s *stubCategoryService) ScanCategories(rootPath string, excludedCategories map[string]bool) ([]entities.CategoryInfo, error) {
	return s.scanCategoriesResult, s.scanCategoriesErr
}

func (s *stubCategoryService) GetOutfits(categoryPath string) ([]entities.FileEntry, error) {
	if s.outfitsErr != nil {
		return nil, s.outfitsErr
	}
	return s.outfitsByPath[categoryPath], nil
}

func markGlobalShown(app *Application, key string) {
	app.session.MarkGlobalShown(key)
}

func markCategoryShown(app *Application, category string, files ...string) {
	for _, file := range files {
		app.session.MarkCategoryShown(file, category)
	}
}

func globalShownCount(app *Application) int {
	return app.session.GlobalShownCount()
}

func categoryShownCount(app *Application, category string) int {
	return app.session.CategoryShownCount(category)
}

func trackedCategoryCount(app *Application) int {
	return app.session.TrackedCategoryCount()
}

func isGlobalShown(app *Application, key string) bool {
	return app.session.IsGlobalShown(key)
}

func newOutfitCachePtr() *entities.OutfitCache {
	cache := entities.NewOutfitCache()
	return &cache
}

func stringPtr(value string) *string { return &value }

var _ usecases.ConfigManager = (*stubConfigManager)(nil)
