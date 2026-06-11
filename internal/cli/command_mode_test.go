package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestExecuteCommand_Paths(t *testing.T) {
	t.Run("shows paths", func(t *testing.T) {
		runtime := newStubRuntime()
		runtime.config.currentConfig = mustCommandConfig(t, cliTestOutfitRoot, map[string]bool{"jackets": true})
		runtime.pathProvider = StaticStoragePathProvider{
			ConfigPath: "/state/outfitpicker/config.json",
			CachePath:  "/state/outfitpicker/cache.json",
		}

		var stdout bytes.Buffer
		handled, code := ExecuteCommand([]string{"paths"}, runtime, TerminalConsole{stdout: &stdout})

		if !handled || code != 0 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
		}
		assertOutputContains(t, stdout.String(), "Config file: /state/outfitpicker/config.json", "Cache file:  /state/outfitpicker/cache.json", "Wardrobe:    "+cliTestOutfitRoot)
	})

	t.Run("config path error", func(t *testing.T) {
		runtime := newStubRuntime()
		runtime.pathProvider = FuncStoragePathProvider{
			ConfigPathFunc: func() (string, error) { return "", errors.New("config path failed") },
			CachePathFunc:  func() (string, error) { return "/cache.json", nil },
		}
		var stderr bytes.Buffer
		handled, code := ExecuteCommand([]string{"paths"}, runtime, TerminalConsole{stderr: &stderr})

		if !handled || code != 1 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 1", handled, code)
		}
		assertOutputContains(t, stderr.String(), "Failed to resolve config path", "config path failed")
	})

	t.Run("cache path error", func(t *testing.T) {
		runtime := newStubRuntime()
		runtime.pathProvider = FuncStoragePathProvider{
			ConfigPathFunc: func() (string, error) { return "/config.json", nil },
			CachePathFunc:  func() (string, error) { return "", errors.New("cache path failed") },
		}
		var stderr bytes.Buffer
		handled, code := ExecuteCommand([]string{"paths"}, runtime, TerminalConsole{stderr: &stderr})

		if !handled || code != 1 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 1", handled, code)
		}
		assertOutputContains(t, stderr.String(), "Failed to resolve cache path", "cache path failed")
	})
}

func TestExecuteCommand_Doctor(t *testing.T) {
	t.Run("healthy with warnings", func(t *testing.T) {
		runtime := newStubRuntime()
		stateDir := t.TempDir()
		wardrobeDir := cliTestHomeTempDir(t, "outfitpicker-doctor-wardrobe-*")
		configPath := filepath.Join(stateDir, "config.json")
		cachePath := filepath.Join(stateDir, "cache.json")
		if err := os.WriteFile(configPath, []byte("{}"), 0600); err != nil {
			t.Fatalf("WriteFile(config) error = %v", err)
		}
		runtime.pathProvider = StaticStoragePathProvider{ConfigPath: configPath, CachePath: cachePath}
		runtime.config.currentConfig = mustCommandConfig(t, wardrobeDir, map[string]bool{"jackets": true})
		runtime.wardrobe.categoryInfos = []entities.CategoryInfo{
			entities.NewCategoryInfo(entities.NewCategoryReference("Shoes", filepath.Join(wardrobeDir, "Shoes")), entities.CategoryStateNoAvatarFiles, 0),
			entities.NewCategoryInfo(entities.NewCategoryReference("Jackets", filepath.Join(wardrobeDir, "Jackets")), entities.CategoryStateUserExcluded, 3),
			entities.NewCategoryInfo(entities.NewCategoryReference("Hats", filepath.Join(wardrobeDir, "Hats")), entities.CategoryStateHasOutfits, 2),
		}
		runtime.wardrobe.allOutfitStates = map[string]entities.CategoryOutfitState{}

		var stdout bytes.Buffer
		handled, code := ExecuteCommand([]string{"doctor"}, runtime, TerminalConsole{stdout: &stdout})

		if !handled || code != 1 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 1 for warnings", handled, code)
		}
		assertOutputContains(t, stdout.String(), "Config file exists", "Wardrobe directory exists", "Found 3 categories", "Found 5 .avatar files", "Shoes has no .avatar files", "Jackets is excluded from random selection", "Cache file is valid")
	})

	t.Run("invalid cache fails", func(t *testing.T) {
		runtime := newStubRuntime()
		stateDir := t.TempDir()
		wardrobeDir := cliTestHomeTempDir(t, "outfitpicker-doctor-wardrobe-*")
		configPath := filepath.Join(stateDir, "config.json")
		if err := os.WriteFile(configPath, []byte("{}"), 0600); err != nil {
			t.Fatalf("WriteFile(config) error = %v", err)
		}
		runtime.pathProvider = StaticStoragePathProvider{ConfigPath: configPath, CachePath: filepath.Join(stateDir, "cache.json")}
		runtime.config.currentConfig = mustCommandConfig(t, wardrobeDir, nil)
		runtime.wardrobe.allOutfitStatesErr = errors.New("bad json")

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		handled, code := ExecuteCommand([]string{"doctor"}, runtime, TerminalConsole{stdout: &stdout, stderr: &stderr})

		if !handled || code != 1 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 1", handled, code)
		}
		assertOutputContains(t, stderr.String(), "Cache file is invalid", "bad json")
	})

	t.Run("missing config warns", func(t *testing.T) {
		runtime := newStubRuntime()
		stateDir := t.TempDir()
		runtime.pathProvider = StaticStoragePathProvider{ConfigPath: filepath.Join(stateDir, "missing-config.json"), CachePath: filepath.Join(stateDir, "cache.json")}

		var stdout bytes.Buffer
		handled, code := ExecuteCommand([]string{"doctor"}, runtime, TerminalConsole{stdout: &stdout})

		if !handled || code != 1 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 1", handled, code)
		}
		assertOutputContains(t, stdout.String(), "Config file does not exist", "Wardrobe is not configured")
	})

	t.Run("invalid config fails", func(t *testing.T) {
		runtime := newStubRuntime()
		stateDir := t.TempDir()
		configPath := filepath.Join(stateDir, "config.json")
		if err := os.WriteFile(configPath, []byte("{}"), 0600); err != nil {
			t.Fatalf("WriteFile(config) error = %v", err)
		}
		runtime.pathProvider = StaticStoragePathProvider{ConfigPath: configPath, CachePath: filepath.Join(stateDir, "cache.json")}
		runtime.config.loadErr = errors.New("invalid config")

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		handled, code := ExecuteCommand([]string{"doctor"}, runtime, TerminalConsole{stdout: &stdout, stderr: &stderr})

		if !handled || code != 1 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 1", handled, code)
		}
		assertOutputContains(t, stderr.String(), "Config file is invalid", "invalid config")
	})

	t.Run("missing wardrobe warns", func(t *testing.T) {
		runtime := newStubRuntime()
		stateDir := t.TempDir()
		configPath := filepath.Join(stateDir, "config.json")
		if err := os.WriteFile(configPath, []byte("{}"), 0600); err != nil {
			t.Fatalf("WriteFile(config) error = %v", err)
		}
		runtime.pathProvider = StaticStoragePathProvider{ConfigPath: configPath, CachePath: filepath.Join(stateDir, "cache.json")}
		runtime.config.currentConfig = mustCommandConfig(t, filepath.Join(os.Getenv("HOME"), "outfitpicker-missing-wardrobe"), nil)
		runtime.wardrobe.allOutfitStates = map[string]entities.CategoryOutfitState{}

		var stdout bytes.Buffer
		handled, code := ExecuteCommand([]string{"doctor"}, runtime, TerminalConsole{stdout: &stdout})

		if !handled || code != 1 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 1", handled, code)
		}
		assertOutputContains(t, stdout.String(), "Wardrobe directory does not exist", "Cache file is valid")
	})

	t.Run("category scan failure", func(t *testing.T) {
		runtime := newStubRuntime()
		stateDir := t.TempDir()
		wardrobeDir := cliTestHomeTempDir(t, "outfitpicker-doctor-wardrobe-*")
		configPath := filepath.Join(stateDir, "config.json")
		if err := os.WriteFile(configPath, []byte("{}"), 0600); err != nil {
			t.Fatalf("WriteFile(config) error = %v", err)
		}
		runtime.pathProvider = StaticStoragePathProvider{ConfigPath: configPath, CachePath: filepath.Join(stateDir, "cache.json")}
		runtime.config.currentConfig = mustCommandConfig(t, wardrobeDir, nil)
		runtime.wardrobe.categoryInfoErr = errors.New("scan failed")

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		handled, code := ExecuteCommand([]string{"doctor"}, runtime, TerminalConsole{stdout: &stdout, stderr: &stderr})

		if !handled || code != 1 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 1", handled, code)
		}
		assertOutputContains(t, stderr.String(), "Could not scan wardrobe categories", "scan failed")
	})
}

func TestExecuteCommand_Pick(t *testing.T) {
	runtime := newStubRuntime()
	category := entities.NewCategoryReference("shoes", cliTestCategoryPath("shoes"))
	outfit := entities.NewOutfitReference("boots.avatar", category)
	runtime.random.globalResults = []stubSelectorResult{{outfit: &outfit}}

	var stdout bytes.Buffer
	handled, code := ExecuteCommand([]string{"pick"}, runtime, TerminalConsole{stdin: strings.NewReader("\n"), stdout: &stdout})

	if !handled {
		t.Fatal("ExecuteCommand() handled = false, want true")
	}
	if code != 0 {
		t.Fatalf("ExecuteCommand() code = %d, want 0", code)
	}
	if runtime.random.globalCalls != 1 {
		t.Fatalf("global random calls = %d, want 1", runtime.random.globalCalls)
	}
	if len(runtime.commands.wearCalls) != 1 || runtime.commands.wearCalls[0].FileName != "boots.avatar" {
		t.Fatalf("wear calls = %#v, want boots.avatar", runtime.commands.wearCalls)
	}
	assertOutputContains(t, stdout.String(), "Outfit picked", "Category: shoes", "boots.avatar", "Mark as worn? [Y/n]", "Marked worn")
}

func TestExecuteCommand_PickNoMark(t *testing.T) {
	runtime := newStubRuntime()
	category := entities.NewCategoryReference("shoes", cliTestCategoryPath("shoes"))
	outfit := entities.NewOutfitReference("boots.avatar", category)
	runtime.random.globalResults = []stubSelectorResult{{outfit: &outfit}}

	var stdout bytes.Buffer
	handled, code := ExecuteCommand([]string{"pick", "--no-mark"}, runtime, TerminalConsole{stdout: &stdout})

	if !handled || code != 0 {
		t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
	}
	if len(runtime.commands.wearCalls) != 0 {
		t.Fatalf("wear calls = %#v, want none", runtime.commands.wearCalls)
	}
	assertOutputContains(t, stdout.String(), "Outfit picked", "Category: shoes", "boots.avatar", "Not marked worn")
	assertOutputNotContains(t, stdout.String(), "Mark as worn?")
}

func TestExecuteCommand_PickRejectsConflictingMarkFlags(t *testing.T) {
	var stderr bytes.Buffer
	handled, code := ExecuteCommand([]string{"pick", "--mark-worn", "--no-mark"}, nil, TerminalConsole{stderr: &stderr})

	if !handled || code != 2 {
		t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 2", handled, code)
	}
	assertOutputContains(t, stderr.String(), "Usage:")
}

func TestExecuteCommand_PickMarkWorn(t *testing.T) {
	runtime := newStubRuntime()
	category := entities.NewCategoryReference("shoes", cliTestCategoryPath("shoes"))
	outfit := entities.NewOutfitReference("boots.avatar", category)
	runtime.random.globalResults = []stubSelectorResult{{outfit: &outfit}}

	var stdout bytes.Buffer
	handled, code := ExecuteCommand([]string{"pick", "--mark-worn"}, runtime, TerminalConsole{stdout: &stdout})

	if !handled || code != 0 {
		t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
	}
	if len(runtime.commands.wearCalls) != 1 {
		t.Fatalf("wear calls = %#v, want one", runtime.commands.wearCalls)
	}
	assertOutputContains(t, stdout.String(), "Marked worn")
	assertOutputNotContains(t, stdout.String(), "Mark as worn?")
}

func TestExecuteCommand_PickPromptCanSkipMarkingWorn(t *testing.T) {
	runtime := newStubRuntime()
	category := entities.NewCategoryReference("shoes", cliTestCategoryPath("shoes"))
	outfit := entities.NewOutfitReference("boots.avatar", category)
	runtime.random.globalResults = []stubSelectorResult{{outfit: &outfit}}

	var stdout bytes.Buffer
	handled, code := ExecuteCommand([]string{"pick"}, runtime, TerminalConsole{stdin: strings.NewReader("n\n"), stdout: &stdout})

	if !handled || code != 0 {
		t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
	}
	if len(runtime.commands.wearCalls) != 0 {
		t.Fatalf("wear calls = %#v, want none", runtime.commands.wearCalls)
	}
	assertOutputContains(t, stdout.String(), "Mark as worn? [Y/n]", "Not marked worn")
}

func TestExecuteCommand_PickCategory(t *testing.T) {
	runtime := newStubRuntime()
	category := entities.NewCategoryReference("shoes", cliTestCategoryPath("shoes"))
	outfit := entities.NewOutfitReference("loafers.avatar", category)
	runtime.random.categoryResults = []stubSelectorResult{{outfit: &outfit}}

	var stdout bytes.Buffer
	handled, code := ExecuteCommand([]string{"pick", "--category", "shoes", "--mark-worn"}, runtime, TerminalConsole{stdout: &stdout})

	if !handled || code != 0 {
		t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
	}
	if runtime.random.categoryCalls != 1 {
		t.Fatalf("category random calls = %d, want 1", runtime.random.categoryCalls)
	}
	if len(runtime.commands.wearCalls) != 1 || runtime.commands.wearCalls[0].Category.Name != "shoes" {
		t.Fatalf("wear calls = %#v, want shoes outfit", runtime.commands.wearCalls)
	}
	assertOutputContains(t, stdout.String(), "Category: shoes", "loafers.avatar")
}

func TestExecuteCommand_PickIncludeExcluded(t *testing.T) {
	originalRandomIndex := commandRandomIndex
	commandRandomIndex = func(int) int { return 1 }
	t.Cleanup(func() { commandRandomIndex = originalRandomIndex })

	runtime := newStubRuntime()
	casual := entities.NewCategoryReference("casual", cliTestCategoryPath("casual"))
	formal := entities.NewCategoryReference("formal", cliTestCategoryPath("formal"))
	runtime.wardrobe.categoryInfos = []entities.CategoryInfo{
		entities.NewCategoryInfo(casual, entities.CategoryStateHasOutfits, 1),
		entities.NewCategoryInfo(formal, entities.CategoryStateUserExcluded, 1),
	}
	runtime.wardrobe.availableOutfitsByName = map[string][]entities.OutfitReference{
		"casual": {entities.NewOutfitReference("casual.avatar", casual)},
		"formal": {entities.NewOutfitReference("formal.avatar", formal)},
	}

	var stdout bytes.Buffer
	handled, code := ExecuteCommand([]string{"pick", "--include-excluded", "--mark-worn"}, runtime, TerminalConsole{stdout: &stdout})

	if !handled || code != 0 {
		t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
	}
	if runtime.random.globalCalls != 0 {
		t.Fatalf("global random calls = %d, want 0 for include-excluded path", runtime.random.globalCalls)
	}
	if len(runtime.commands.wearCalls) != 1 || runtime.commands.wearCalls[0].Category.Name != "formal" {
		t.Fatalf("wear calls = %#v, want formal outfit", runtime.commands.wearCalls)
	}
	assertOutputContains(t, stdout.String(), "Category: formal", "formal.avatar", "Marked worn")
}

func TestExecuteCommand_PickInvalidPromptReturnsUsageError(t *testing.T) {
	runtime := newStubRuntime()
	category := entities.NewCategoryReference("shoes", cliTestCategoryPath("shoes"))
	outfit := entities.NewOutfitReference("boots.avatar", category)
	runtime.random.globalResults = []stubSelectorResult{{outfit: &outfit}}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	handled, code := ExecuteCommand([]string{"pick"}, runtime, TerminalConsole{stdin: strings.NewReader("maybe\n"), stdout: &stdout, stderr: &stderr})

	if !handled || code != 2 {
		t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 2", handled, code)
	}
	if len(runtime.commands.wearCalls) != 0 {
		t.Fatalf("wear calls = %#v, want none", runtime.commands.wearCalls)
	}
	assertOutputContains(t, stderr.String(), "Please answer yes or no")
}

func TestExecuteCommand_PickNoOutfits(t *testing.T) {
	runtime := newStubRuntime()
	var stdout bytes.Buffer

	handled, code := ExecuteCommand([]string{"pick", "--mark-worn"}, runtime, TerminalConsole{stdout: &stdout})

	if !handled || code != 0 {
		t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
	}
	assertOutputContains(t, stdout.String(), "No outfits available")
}

func TestExecuteCommand_PickIncludeExcludedNoOutfits(t *testing.T) {
	runtime := newStubRuntime()
	runtime.wardrobe.categoryInfos = []entities.CategoryInfo{
		entities.NewCategoryInfo(entities.NewCategoryReference("empty", cliTestCategoryPath("empty")), entities.CategoryStateEmpty, 0),
	}
	var stdout bytes.Buffer

	handled, code := ExecuteCommand([]string{"pick", "--include-excluded", "--mark-worn"}, runtime, TerminalConsole{stdout: &stdout})

	if !handled || code != 0 {
		t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
	}
	assertOutputContains(t, stdout.String(), "No outfits available")
}

func TestExecuteCommand_PickIncludeExcludedPropagatesAvailableOutfitError(t *testing.T) {
	runtime := newStubRuntime()
	category := entities.NewCategoryReference("formal", cliTestCategoryPath("formal"))
	runtime.wardrobe.categoryInfos = []entities.CategoryInfo{
		entities.NewCategoryInfo(category, entities.CategoryStateUserExcluded, 1),
	}
	runtime.wardrobe.availableOutfitErrors = map[string]error{"formal": errors.New("available outfits failed")}
	var stderr bytes.Buffer

	handled, code := ExecuteCommand([]string{"pick", "--include-excluded", "--mark-worn"}, runtime, TerminalConsole{stderr: &stderr})

	if !handled || code != 1 {
		t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 1", handled, code)
	}
	assertOutputContains(t, stderr.String(), "Failed to pick outfit")
}

func mustCommandConfig(t *testing.T, root string, excluded map[string]bool) *entities.Config {
	t.Helper()
	config, err := entities.NewConfig(root, nil, excluded, nil, nil)
	if err != nil {
		t.Fatalf("NewConfig(%q) error = %v", root, err)
	}
	return config
}

func TestExpandHomePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	tests := []struct {
		path string
		want string
	}{
		{path: "~", want: home},
		{path: "~/Outfits", want: filepath.Join(home, "Outfits")},
		{path: cliTestOutfitRoot, want: cliTestOutfitRoot},
	}

	for _, tt := range tests {
		got, err := expandHomePath(tt.path)
		if err != nil {
			t.Fatalf("expandHomePath(%q) error = %v", tt.path, err)
		}
		if got != tt.want {
			t.Fatalf("expandHomePath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestExecuteCommand_ListCategories(t *testing.T) {
	runtime := newStubRuntime()
	runtime.wardrobe.categoryInfos = []entities.CategoryInfo{
		entities.NewCategoryInfo(entities.NewCategoryReference("shoes", cliTestCategoryPath("shoes")), entities.CategoryStateHasOutfits, 2),
		entities.NewCategoryInfo(entities.NewCategoryReference("hats", cliTestCategoryPath("hats")), entities.CategoryStateEmpty, 0),
	}

	var stdout bytes.Buffer
	handled, code := ExecuteCommand([]string{"list", "categories"}, runtime, TerminalConsole{stdout: &stdout})

	if !handled || code != 0 {
		t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
	}
	assertOutputContains(t, stdout.String(), "shoes", "hasOutfits", "2 outfits", "hats", "empty")
}

func TestExecuteCommand_ListWornAndUnworn(t *testing.T) {
	category := entities.NewCategoryReference("shoes", cliTestCategoryPath("shoes"))
	runtime := newStubRuntime()
	runtime.wardrobe.allOutfitStates = map[string]entities.CategoryOutfitState{
		"shoes": entities.NewCategoryOutfitState(
			category,
			[]entities.OutfitReference{
				entities.NewOutfitReference("boots.avatar", category),
				entities.NewOutfitReference("loafers.avatar", category),
			},
			[]entities.OutfitReference{entities.NewOutfitReference("loafers.avatar", category)},
			[]entities.OutfitReference{entities.NewOutfitReference("boots.avatar", category)},
		),
	}

	t.Run("worn", func(t *testing.T) {
		var stdout bytes.Buffer
		handled, code := ExecuteCommand([]string{"list", "worn"}, runtime, TerminalConsole{stdout: &stdout})
		if !handled || code != 0 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
		}
		assertOutputContains(t, stdout.String(), "shoes", "boots.avatar")
	})

	t.Run("unworn", func(t *testing.T) {
		var stdout bytes.Buffer
		handled, code := ExecuteCommand([]string{"list", "unworn"}, runtime, TerminalConsole{stdout: &stdout})
		if !handled || code != 0 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
		}
		assertOutputContains(t, stdout.String(), "shoes", "loafers.avatar")
	})
}

func TestExecuteCommand_Reset(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		runtime := newStubRuntime()
		var stdout bytes.Buffer

		handled, code := ExecuteCommand([]string{"reset"}, runtime, TerminalConsole{stdout: &stdout})

		if !handled || code != 0 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
		}
		if runtime.commands.resetAllCalls != 1 {
			t.Fatalf("reset all calls = %d, want 1", runtime.commands.resetAllCalls)
		}
		assertOutputContains(t, stdout.String(), "Reset all worn outfits")
	})

	t.Run("category", func(t *testing.T) {
		runtime := newStubRuntime()
		var stdout bytes.Buffer

		handled, code := ExecuteCommand([]string{"reset", "--category", "shoes"}, runtime, TerminalConsole{stdout: &stdout})

		if !handled || code != 0 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
		}
		if len(runtime.commands.resetCategoryCalls) != 1 || runtime.commands.resetCategoryCalls[0] != "shoes" {
			t.Fatalf("reset category calls = %#v, want shoes", runtime.commands.resetCategoryCalls)
		}
		assertOutputContains(t, stdout.String(), "Reset worn outfits for shoes")
	})
}

func TestExecuteCommand_Config(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		runtime := newStubRuntime()
		runtime.config.currentConfig = mustTestConfig(t, cliTestOutfitRoot, map[string]bool{"jackets": true})
		var stdout bytes.Buffer

		handled, code := ExecuteCommand([]string{"config", "get"}, runtime, TerminalConsole{stdout: &stdout})

		if !handled || code != 0 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
		}
		assertOutputContains(t, stdout.String(), "Root:", cliTestOutfitRoot, "Language: en", "Excluded: jackets")
	})

	t.Run("set-root", func(t *testing.T) {
		runtime := newStubRuntime()
		runtime.config.currentConfig = mustTestConfig(t, cliTestOutfitRoot, map[string]bool{"jackets": true})
		var stdout bytes.Buffer

		handled, code := ExecuteCommand([]string{"config", "set-root", cliTestNewOutfitRoot}, runtime, TerminalConsole{stdout: &stdout})

		if !handled || code != 0 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
		}
		if len(runtime.config.updatedConfigs) != 1 {
			t.Fatalf("updated configs = %d, want 1", len(runtime.config.updatedConfigs))
		}
		if runtime.config.updatedConfigs[0].Root != cliTestNewOutfitRoot {
			t.Fatalf("updated root = %q, want %q", runtime.config.updatedConfigs[0].Root, cliTestNewOutfitRoot)
		}
		if !runtime.config.updatedConfigs[0].ExcludedCategories["jackets"] {
			t.Fatal("expected existing excluded category to be preserved")
		}
		assertOutputContains(t, stdout.String(), "Outfit path updated")
	})

	t.Run("exclude", func(t *testing.T) {
		runtime := newStubRuntime()
		runtime.config.currentConfig = mustTestConfig(t, cliTestOutfitRoot, map[string]bool{"jackets": true})
		var stdout bytes.Buffer

		handled, code := ExecuteCommand([]string{"config", "exclude", "shoes", "hats"}, runtime, TerminalConsole{stdout: &stdout})

		if !handled || code != 0 {
			t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 0", handled, code)
		}
		updated := runtime.config.updatedConfigs[0]
		for _, category := range []string{"jackets", "shoes", "hats"} {
			if !updated.ExcludedCategories[category] {
				t.Fatalf("expected %q to be excluded in %#v", category, updated.ExcludedCategories)
			}
		}
		assertOutputContains(t, stdout.String(), "Excluded categories updated", "hats", "shoes")
	})
}

func TestExecuteCommand_Help(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "long flag", args: []string{"--help"}, want: []string{"Usage:", "pick", "list", "reset", "config"}},
		{name: "help command", args: []string{"help"}, want: []string{"Usage:", "pick", "list", "reset", "config"}},
		{name: "pick help", args: []string{"pick", "--help"}, want: []string{"Usage:", "pick", "--category", "--mark-worn", "--no-mark", "--include-excluded"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			handled, code := ExecuteCommand(tt.args, nil, TerminalConsole{stdout: &stdout})

			if !handled || code != 0 {
				t.Fatalf("ExecuteCommand(%#v) = handled %t code %d, want handled true code 0", tt.args, handled, code)
			}
			assertOutputContains(t, stdout.String(), tt.want...)
		})
	}
}

func TestExecuteCommand_UnknownArgReturnsUsageError(t *testing.T) {
	var stderr bytes.Buffer
	handled, code := ExecuteCommand([]string{"--wat"}, nil, TerminalConsole{stderr: &stderr})

	if !handled || code != 2 {
		t.Fatalf("ExecuteCommand() = handled %t code %d, want handled true code 2", handled, code)
	}
	assertOutputContains(t, stderr.String(), "Usage:")
}

func TestExecuteCommand_ReturnsUnhandledForInteractiveMode(t *testing.T) {
	handled, code := ExecuteCommand(nil, newStubRuntime(), TerminalConsole{})
	if handled || code != 0 {
		t.Fatalf("ExecuteCommand(nil) = handled %t code %d, want handled false code 0", handled, code)
	}
}

func mustTestConfig(t *testing.T, root string, excluded map[string]bool) *entities.Config {
	t.Helper()
	config, err := entities.NewConfig(root, stringPtr("en"), excluded, nil, nil)
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	return config
}

func assertOutputContains(t *testing.T, output string, parts ...string) {
	t.Helper()
	for _, part := range parts {
		if !strings.Contains(output, part) {
			t.Fatalf("output = %q, want to contain %q", output, part)
		}
	}
}

func assertOutputNotContains(t *testing.T, output string, parts ...string) {
	t.Helper()
	for _, part := range parts {
		if strings.Contains(output, part) {
			t.Fatalf("output = %q, want not to contain %q", output, part)
		}
	}
}
