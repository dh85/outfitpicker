package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	domainerrors "github.com/dh85/outfitpicker/internal/domain/errors"
	"github.com/dh85/outfitpicker/internal/infrastructure/system"
)

func TestIntegration_FirstRunBootstrapToUsableMenu(t *testing.T) {
	configHome := integrationConfigHome(t)
	t.Setenv("XDG_CONFIG_HOME", configHome)
	deps := newProductionStyleRuntimeDependencies()
	root := integrationWardrobeRoot(t, map[string][]string{
		"casual": {"one.avatar"},
	})

	restore := withPromptResponses(t, root, "", "", "", "r", "b", "q")
	defer restore()

	app, ok := BootstrapApplication(deps, nil)
	if !ok {
		t.Fatal("BootstrapApplication() expected success")
	}
	if app == nil {
		t.Fatalf("BootstrapApplication() returned %#v, want app with config", app)
	}

	configPath := integrationConfigPath(t)
	if !integrationFileExists(configPath) {
		t.Fatalf("config file %q was not created", configPath)
	}

	loaded, err := LoadApplicationFromExistingConfig(deps)
	if err != nil {
		t.Fatalf("LoadApplicationFromExistingConfig() error = %v", err)
	}
	loadedConfig, err := loaded.GetConfiguration()
	if err != nil {
		t.Fatalf("loaded GetConfiguration() error = %v", err)
	}
	if loadedConfig.Root != root {
		t.Fatalf("loaded root = %q, want %q", loadedConfig.Root, root)
	}

	newMenuSystemForRuntime(app).ShowMainMenu()
}

func TestIntegration_WearFlowPersistsAcrossRestart(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", integrationConfigHome(t))
	deps := newProductionStyleRuntimeDependencies()
	root := integrationWardrobeRoot(t, map[string][]string{
		"casual": {"one.avatar", "two.avatar"},
	})

	app, err := CreateApplicationFromConfiguration(Configuration{OutfitPath: root, Language: "en"}, deps)
	if err != nil {
		t.Fatalf("CreateApplicationFromConfiguration() error = %v", err)
	}

	restore := withPromptResponses(t, "r", "w")
	defer restore()
	newMenuSystemForRuntime(app).ShowMainMenu()

	reloaded, err := LoadApplicationFromExistingConfig(deps)
	if err != nil {
		t.Fatalf("LoadApplicationFromExistingConfig() error = %v", err)
	}

	state, err := reloaded.GetOutfitState(entities.NewCategoryReference("casual", ""))
	if err != nil {
		t.Fatalf("GetOutfitState() error = %v", err)
	}
	if len(state.WornOutfits) != 1 {
		t.Fatalf("worn outfits = %#v, want exactly one persisted worn outfit", state.WornOutfits)
	}
	if len(state.AvailableOutfits) != 1 {
		t.Fatalf("available outfits = %#v, want exactly one remaining available outfit", state.AvailableOutfits)
	}

	wornName := state.WornOutfits[0].FileName
	availableName := state.AvailableOutfits[0].FileName
	if wornName == availableName {
		t.Fatalf("worn and available outfits should differ, got %q and %q", wornName, availableName)
	}
	if (wornName != "one.avatar" && wornName != "two.avatar") || (availableName != "one.avatar" && availableName != "two.avatar") {
		t.Fatalf("unexpected persisted outfits, worn=%q available=%q", wornName, availableName)
	}
}

func TestIntegration_RootPathChangeResetsPersistedWornState(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", integrationConfigHome(t))
	deps := newProductionStyleRuntimeDependencies()
	rootOne := integrationWardrobeRoot(t, map[string][]string{
		"casual": {"one.avatar", "two.avatar"},
	})
	rootTwo := integrationWardrobeRoot(t, map[string][]string{
		"casual": {"one.avatar"},
	})

	app, err := CreateApplicationFromConfiguration(Configuration{OutfitPath: rootOne, Language: "en"}, deps)
	if err != nil {
		t.Fatalf("CreateApplicationFromConfiguration() error = %v", err)
	}

	if err := app.WearOutfit(entities.NewOutfitReference("one.avatar", entities.NewCategoryReference("casual", filepath.Join(rootOne, "casual")))); err != nil {
		t.Fatalf("WearOutfit() error = %v", err)
	}

	cachePath := integrationCachePath(t)
	if !integrationFileExists(cachePath) {
		t.Fatalf("cache file %q was not created before path change", cachePath)
	}

	menu := AdvancedMenu{outfitService: NewOutfitServiceFromRuntime(app)}
	restore := withPromptResponses(t, rootTwo, "y")
	defer restore()
	menu.handlePathChange()

	if integrationFileExists(cachePath) {
		t.Fatalf("cache file %q still exists after root change", cachePath)
	}

	reloaded, err := LoadApplicationFromExistingConfig(deps)
	if err != nil {
		t.Fatalf("LoadApplicationFromExistingConfig() error = %v", err)
	}
	state, err := reloaded.GetOutfitState(entities.NewCategoryReference("casual", ""))
	if err != nil {
		t.Fatalf("GetOutfitState() error = %v", err)
	}
	if len(state.WornOutfits) != 0 {
		t.Fatalf("worn outfits leaked across root change: %#v", state.WornOutfits)
	}
	if len(state.AvailableOutfits) != 1 || state.AvailableOutfits[0].FileName != "one.avatar" {
		t.Fatalf("available outfits = %#v, want one.avatar available in new root", state.AvailableOutfits)
	}
}

func TestIntegration_ExcludedCategoriesHonoredEndToEnd(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", integrationConfigHome(t))
	deps := newProductionStyleRuntimeDependencies()
	root := integrationWardrobeRoot(t, map[string][]string{
		"casual": {"one.avatar"},
		"formal": {"suit.avatar"},
	})

	restore := withPromptResponses(t, root, "", "", "formal")
	defer restore()

	_, ok := BootstrapApplication(deps, nil)
	if !ok {
		t.Fatal("BootstrapApplication() expected success")
	}

	reloaded, err := LoadApplicationFromExistingConfig(deps)
	if err != nil {
		t.Fatalf("LoadApplicationFromExistingConfig() error = %v", err)
	}
	config, err := reloaded.GetConfiguration()
	if err != nil {
		t.Fatalf("GetConfiguration() error = %v", err)
	}
	if !config.ExcludedCategories["formal"] {
		t.Fatalf("excluded categories = %#v, want formal to be excluded", config.ExcludedCategories)
	}

	outfit, err := reloaded.ShowNextUniqueRandomOutfit()
	if err != nil {
		t.Fatalf("ShowNextUniqueRandomOutfit() error = %v", err)
	}
	if outfit == nil || outfit.Category.Name != "casual" {
		t.Fatalf("ShowNextUniqueRandomOutfit() = %#v, want casual outfit only", outfit)
	}
}

func TestIntegration_FactoryResetDeletesConfigAndCache(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", integrationConfigHome(t))
	deps := newProductionStyleRuntimeDependencies()
	root := integrationWardrobeRoot(t, map[string][]string{
		"casual": {"one.avatar", "two.avatar"},
	})

	app, err := CreateApplicationFromConfiguration(Configuration{OutfitPath: root, Language: "en"}, deps)
	if err != nil {
		t.Fatalf("CreateApplicationFromConfiguration() error = %v", err)
	}
	if err := app.WearOutfit(entities.NewOutfitReference("one.avatar", entities.NewCategoryReference("casual", filepath.Join(root, "casual")))); err != nil {
		t.Fatalf("WearOutfit() error = %v", err)
	}

	configPath := integrationConfigPath(t)
	cachePath := integrationCachePath(t)
	if !integrationFileExists(configPath) || !integrationFileExists(cachePath) {
		t.Fatalf("expected config and cache files to exist before reset, config=%t cache=%t", integrationFileExists(configPath), integrationFileExists(cachePath))
	}

	if err := app.FactoryReset(); err != nil {
		t.Fatalf("FactoryReset() error = %v", err)
	}
	if integrationFileExists(configPath) || integrationFileExists(cachePath) {
		t.Fatalf("files still exist after reset, config=%t cache=%t", integrationFileExists(configPath), integrationFileExists(cachePath))
	}

	restore := withPromptResponses(t, "   ")
	defer restore()
	_, ok := BootstrapApplication(deps, nil)
	if ok {
		t.Fatal("BootstrapApplication() expected first-time setup flow to stop when setup is cancelled")
	}

	_, err = LoadApplicationFromExistingConfig(deps)
	if !errors.Is(err, domainerrors.ErrConfigurationNotFound) {
		t.Fatalf("LoadApplicationFromExistingConfig() error = %v, want %v", err, domainerrors.ErrConfigurationNotFound)
	}
}

func integrationWardrobeRoot(t *testing.T, categories map[string][]string) string {
	t.Helper()
	root := cliTestHomeTempDir(t, "outfitpicker-integration-wardrobe-")

	for category, files := range categories {
		categoryDir := filepath.Join(root, category)
		if err := os.MkdirAll(categoryDir, 0o755); err != nil {
			t.Fatalf("MkdirAll(%q) error = %v", categoryDir, err)
		}
		for _, file := range files {
			path := filepath.Join(categoryDir, file)
			if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
				t.Fatalf("WriteFile(%q) error = %v", path, err)
			}
		}
	}

	return root
}

func integrationConfigHome(t *testing.T) string {
	t.Helper()
	return cliTestHomeTempDir(t, "outfitpicker-integration-config-")
}

func integrationConfigPath(t *testing.T) string {
	t.Helper()
	path, err := system.NewFileService[entities.Config](configFileName).FilePath()
	if err != nil {
		t.Fatalf("config FilePath() error = %v", err)
	}
	return path
}

func integrationCachePath(t *testing.T) string {
	t.Helper()
	path, err := system.NewFileService[entities.OutfitCache](cacheFileName).FilePath()
	if err != nil {
		t.Fatalf("cache FilePath() error = %v", err)
	}
	return path
}

func integrationFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
