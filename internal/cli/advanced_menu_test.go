package cli

import (
	"errors"
	"reflect"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestPathChangeNeedsResetConfirmation(t *testing.T) {
	tests := []struct {
		name        string
		currentRoot string
		newRoot     string
		want        bool
	}{
		{name: "different root requires confirmation", currentRoot: "/one", newRoot: "/two", want: true},
		{name: "same root does not require confirmation", currentRoot: "/one", newRoot: "/one", want: false},
		{name: "trimmed same root does not require confirmation", currentRoot: "/one", newRoot: " /one ", want: false},
		{name: "empty current root does not require confirmation", currentRoot: "", newRoot: "/two", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pathChangeNeedsResetConfirmation(tt.currentRoot, tt.newRoot)
			if got != tt.want {
				t.Fatalf("pathChangeNeedsResetConfirmation(%q, %q) = %t, want %t", tt.currentRoot, tt.newRoot, got, tt.want)
			}
		})
	}
}

func TestShouldProceedWithDestructiveAction(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "accepts y", input: "y", want: true},
		{name: "accepts yes", input: "yes", want: true},
		{name: "accepts trimmed uppercase yes", input: " Yes ", want: true},
		{name: "back cancels", input: "b", want: false},
		{name: "empty cancels", input: "", want: false},
		{name: "unknown cancels", input: "no", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldProceedWithDestructiveAction(tt.input)
			if got != tt.want {
				t.Fatalf("shouldProceedWithDestructiveAction(%q) = %t, want %t", tt.input, got, tt.want)
			}
		})
	}
}

func TestAdvancedMenu_Show_DispatchesChoices(t *testing.T) {
	tests := []struct {
		name     string
		picker   *advancedMenuTestPicker
		inputs   []string
		wantDest menuDestination
		assert   func(t *testing.T, picker *advancedMenuTestPicker)
	}{
		{
			name:     "invalid choice loops back to menu",
			picker:   newAdvancedMenuTestPicker(),
			inputs:   []string{"x"},
			wantDest: menuDestinationAdvanced,
		},
		{
			name: "dispatches change path",
			picker: newAdvancedMenuTestPicker(
				withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)),
			),
			inputs:   []string{"p", cliTestNewOutfitRoot, "y"},
			wantDest: menuDestinationAdvanced,
			assert: func(t *testing.T, picker *advancedMenuTestPicker) {
				t.Helper()
				config := picker.config.currentConfig
				if config == nil || config.Root != cliTestNewOutfitRoot {
					t.Fatalf("current config root = %v, want %s", config, cliTestNewOutfitRoot)
				}
			},
		},
		{
			name: "dispatches change language",
			picker: newAdvancedMenuTestPicker(
				withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)),
			),
			inputs:   []string{"l", "fr"},
			wantDest: menuDestinationAdvanced,
			assert: func(t *testing.T, picker *advancedMenuTestPicker) {
				t.Helper()
				config := picker.config.currentConfig
				if config == nil || config.Language != "fr" {
					t.Fatalf("current config language = %v, want fr", config)
				}
			},
		},
		{
			name: "dispatches excluded change",
			picker: newAdvancedMenuTestPicker(
				withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)),
				withCategories(categoryRef("casual")),
			),
			inputs:   []string{"e", "b"},
			wantDest: menuDestinationAdvanced,
		},
		{
			name:     "dispatches reset all",
			picker:   newAdvancedMenuTestPicker(),
			inputs:   []string{"r", "y"},
			wantDest: menuDestinationAdvanced,
			assert: func(t *testing.T, picker *advancedMenuTestPicker) {
				t.Helper()
				assertResetAllRequested(t, picker.stubRuntime)
			},
		},
		{
			name:     "dispatches reset category",
			picker:   newAdvancedMenuTestPicker(withCategories(categoryRef("casual"))),
			inputs:   []string{"c", "1"},
			wantDest: menuDestinationAdvanced,
			assert: func(t *testing.T, picker *advancedMenuTestPicker) {
				t.Helper()
				assertCategoryResetRequested(t, picker.stubRuntime, "casual")
			},
		},
		{
			name: "dispatches back to main menu",
			picker: newAdvancedMenuTestPicker(
				withCategoryInfos([]entities.CategoryInfo{}),
				withRootDirectory(cliTestOutfitRoot),
			),
			inputs:   []string{"b"},
			wantDest: menuDestinationMain,
		},
		{
			name:     "dispatches quit",
			picker:   newAdvancedMenuTestPicker(),
			inputs:   []string{"q"},
			wantDest: menuDestinationExit,
		},
		{
			name:     "dispatches reset settings",
			picker:   newAdvancedMenuTestPicker(),
			inputs:   []string{"s", "b"},
			wantDest: menuDestinationAdvanced,
			assert: func(t *testing.T, picker *advancedMenuTestPicker) {
				t.Helper()
				assertNoFactoryResetRequested(t, picker.stubRuntime)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertMenuTransitionWithPrompts(t, tt.wantDest, func() menuTransition {
				return AdvancedMenu{outfitService: newStubOutfitService(tt.picker.stubRuntime)}.Show()
			}, tt.inputs...)

			if tt.assert != nil {
				tt.assert(t, tt.picker)
			}
		})
	}
}

func TestAdvancedMenu_dispatchChoice_UnhandledChoiceExits(t *testing.T) {
	picker := newAdvancedMenuTestPicker()
	menu := AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}

	got := menu.dispatchChoice(AdvancedChoice("unknown"))
	if got.destination != menuDestinationExit {
		t.Fatalf("dispatchChoice() destination = %v, want exit", got.destination)
	}
}

func TestAdvancedMenu_handleResetAll(t *testing.T) {
	t.Run("cancelled", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker()
		assertMenuTransitionWithPrompts(t, menuDestinationAdvanced, func() menuTransition {
			return AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleResetAll()
		}, "b")

		assertNoResetAllRequested(t, picker.stubRuntime)
	})

	t.Run("reports reset error", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withResetAllError(errors.New("boom")))
		assertMenuTransitionWithPrompts(t, menuDestinationAdvanced, func() menuTransition {
			return AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleResetAll()
		}, "y")

		assertResetAllRequested(t, picker.stubRuntime)
	})
}

func TestAdvancedMenu_handleResetCategory(t *testing.T) {
	t.Run("load categories error", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withCategoriesError(errors.New("boom")))

		got := AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleResetCategory()
		if got.destination != menuDestinationAdvanced {
			t.Fatalf("handleResetCategory() destination = %v, want advanced", got.destination)
		}
	})

	t.Run("invalid selection", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withCategories(categoryRef("casual")))
		restore := withPromptResponses(t, "0")
		defer restore()

		got := AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleResetCategory()
		if got.destination != menuDestinationAdvanced {
			t.Fatalf("handleResetCategory() destination = %v, want advanced", got.destination)
		}
		assertNoCategoryResetRequested(t, picker.stubRuntime)
	})

	t.Run("reset error", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withCategories(categoryRef("casual")), withResetCategoryError(errors.New("boom")))
		restore := withPromptResponses(t, "1")
		defer restore()

		got := AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleResetCategory()
		if got.destination != menuDestinationAdvanced {
			t.Fatalf("handleResetCategory() destination = %v, want advanced", got.destination)
		}
		assertCategoryResetRequested(t, picker.stubRuntime, "casual")
	})
}

func TestAdvancedMenu_handlePathChange(t *testing.T) {
	t.Run("get configuration error", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfigError(errors.New("boom")))

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handlePathChange(), menuDestinationAdvanced)
	})

	t.Run("empty path", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)))
		restore := withPromptResponses(t, "")
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handlePathChange(), menuDestinationAdvanced)
		assertCurrentConfigRoot(t, picker, cliTestOutfitRoot)
	})

	t.Run("invalid existing config fails build", func(t *testing.T) {
		config := &entities.Config{Root: cliTestOutfitRoot, Language: "invalid", ExcludedCategories: map[string]bool{}}
		picker := newAdvancedMenuTestPicker(withConfig(config))
		restore := withPromptResponses(t, cliTestNewOutfitRoot)
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handlePathChange(), menuDestinationAdvanced)
		assertCurrentConfigRoot(t, picker, cliTestOutfitRoot)
	})

	t.Run("cancelled on confirmation", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)))
		restore := withPromptResponses(t, cliTestNewOutfitRoot, "b")
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handlePathChange(), menuDestinationAdvanced)
		assertCurrentConfigRoot(t, picker, cliTestOutfitRoot)
	})

	t.Run("update error", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(
			withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)),
			withUpdateError(errors.New("boom")),
		)
		restore := withPromptResponses(t, cliTestNewOutfitRoot, "y")
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handlePathChange(), menuDestinationAdvanced)
		assertCurrentConfigRoot(t, picker, cliTestOutfitRoot)
	})

	t.Run("same root does not ask confirmation", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)))
		restore := withPromptResponses(t, cliTestOutfitRoot)
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handlePathChange(), menuDestinationAdvanced)
		assertCurrentConfigRoot(t, picker, cliTestOutfitRoot)
	})
}

func TestAdvancedMenu_handleLanguageChange(t *testing.T) {
	t.Run("get configuration error", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfigError(errors.New("boom")))

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleLanguageChange(), menuDestinationAdvanced)
	})

	t.Run("empty language", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)))
		restore := withPromptResponses(t, "")
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleLanguageChange(), menuDestinationAdvanced)
		assertCurrentConfigLanguage(t, picker, "en")
	})

	t.Run("invalid language falls back", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)))
		restore := withPromptResponses(t, "invalid")
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleLanguageChange(), menuDestinationAdvanced)
		assertCurrentConfigLanguage(t, picker, "en")
	})

	t.Run("update error", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(
			withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)),
			withUpdateError(errors.New("boom")),
		)
		restore := withPromptResponses(t, "fr")
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleLanguageChange(), menuDestinationAdvanced)
		assertCurrentConfigLanguage(t, picker, "en")
	})

	t.Run("build config error", func(t *testing.T) {
		config := &entities.Config{Root: "", Language: "en", ExcludedCategories: map[string]bool{}}
		picker := newAdvancedMenuTestPicker(withConfig(config))
		restore := withPromptResponses(t, "fr")
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleLanguageChange(), menuDestinationAdvanced)
		assertCurrentConfigLanguage(t, picker, "en")
	})
}

func TestAdvancedMenu_handleExcludedChange(t *testing.T) {
	t.Run("configuration error", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfigError(errors.New("boom")))

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedChange(), menuDestinationAdvanced)
	})

	t.Run("categories error", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(
			withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)),
			withCategoriesError(errors.New("boom")),
		)

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedChange(), menuDestinationAdvanced)
	})

	t.Run("invalid option loops", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(
			withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)),
			withCategories(categoryRef("casual")),
		)
		restore := withPromptResponses(t, "x", "b")
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedChange(), menuDestinationAdvanced)
	})

	t.Run("dispatches add option", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(
			withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)),
			withCategories(categoryRef("casual")),
		)
		restore := withPromptResponses(t, "a", "1", "b")
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedChange(), menuDestinationAdvanced)
		if !picker.config.currentConfig.ExcludedCategories["casual"] {
			t.Fatalf("current excluded categories = %+v", picker.config.currentConfig.ExcludedCategories)
		}
	})

	t.Run("dispatches remove option", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(
			withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", map[string]bool{"casual": true})),
			withCategories(categoryRef("casual")),
		)
		restore := withPromptResponses(t, "r", "1", "b")
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedChange(), menuDestinationAdvanced)
		if len(picker.config.currentConfig.ExcludedCategories) != 0 {
			t.Fatalf("current excluded categories = %+v", picker.config.currentConfig.ExcludedCategories)
		}
	})

	t.Run("dispatches clear option", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(
			withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", map[string]bool{"casual": true})),
			withCategories(categoryRef("casual")),
		)
		restore := withPromptResponses(t, "c", "b")
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedChange(), menuDestinationAdvanced)
		if len(picker.config.currentConfig.ExcludedCategories) != 0 {
			t.Fatalf("current excluded categories = %+v", picker.config.currentConfig.ExcludedCategories)
		}
	})
}

func TestAdvancedMenu_handleExcludedAdd(t *testing.T) {
	t.Run("all already excluded", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", map[string]bool{"casual": true})), withCategories(categoryRef("casual")))
		restore := withPromptResponses(t, "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedAdd(picker.config.currentConfig, nil)
	})

	t.Run("empty input", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)), withCategories(categoryRef("casual")))
		restore := withPromptResponses(t, "", "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedAdd(picker.config.currentConfig, []string{"casual"})
		assertCurrentExcluded(t, picker, map[string]bool{})
	})

	t.Run("invalid selection", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)), withCategories(categoryRef("casual")))
		restore := withPromptResponses(t, "missing", "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedAdd(picker.config.currentConfig, []string{"casual"})
	})

	t.Run("build config error", func(t *testing.T) {
		config := &entities.Config{Root: "", Language: "en", ExcludedCategories: map[string]bool{}}
		picker := newAdvancedMenuTestPicker(withConfig(config), withCategories(categoryRef("casual")))
		restore := withPromptResponses(t, "1", "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedAdd(config, []string{"casual"})
		assertCurrentExcluded(t, picker, map[string]bool{})
	})

	t.Run("update error", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(
			withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)),
			withCategories(categoryRef("casual")),
			withUpdateError(errors.New("boom")),
		)
		restore := withPromptResponses(t, "1", "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedAdd(picker.config.currentConfig, []string{"casual"})
		assertCurrentExcluded(t, picker, map[string]bool{})
	})

	t.Run("success", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)), withCategories(categoryRef("casual")))
		restore := withPromptResponses(t, "1", "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedAdd(picker.config.currentConfig, []string{"casual"})
		assertCurrentExcluded(t, picker, map[string]bool{"casual": true})
	})
}

func TestAdvancedMenu_handleExcludedRemove(t *testing.T) {
	t.Run("none excluded", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)), withCategories(categoryRef("casual")))
		restore := withPromptResponses(t, "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedRemove(picker.config.currentConfig, nil)
	})

	t.Run("empty input", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", map[string]bool{"casual": true})), withCategories(categoryRef("casual")))
		restore := withPromptResponses(t, "", "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedRemove(picker.config.currentConfig, []string{"casual"})
	})

	t.Run("invalid selection", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", map[string]bool{"casual": true})), withCategories(categoryRef("casual")))
		restore := withPromptResponses(t, "missing", "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedRemove(picker.config.currentConfig, []string{"casual"})
	})

	t.Run("build config error", func(t *testing.T) {
		config := &entities.Config{Root: "", Language: "en", ExcludedCategories: map[string]bool{"casual": true}}
		picker := newAdvancedMenuTestPicker(withConfig(config), withCategories(categoryRef("casual")))
		restore := withPromptResponses(t, "1", "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedRemove(config, []string{"casual"})
	})

	t.Run("update error", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(
			withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", map[string]bool{"casual": true})),
			withCategories(categoryRef("casual")),
			withUpdateError(errors.New("boom")),
		)
		restore := withPromptResponses(t, "1", "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedRemove(picker.config.currentConfig, []string{"casual"})
	})

	t.Run("success", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", map[string]bool{"casual": true})), withCategories(categoryRef("casual")))
		restore := withPromptResponses(t, "1", "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedRemove(picker.config.currentConfig, []string{"casual"})
		assertCurrentExcluded(t, picker, map[string]bool{})
	})
}

func TestAdvancedMenu_handleExcludedClear(t *testing.T) {
	t.Run("none excluded", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", nil)), withCategories(categoryRef("casual")))
		restore := withPromptResponses(t, "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedClear(picker.config.currentConfig, nil)
	})

	t.Run("build config error", func(t *testing.T) {
		config := &entities.Config{Root: "", Language: "en", ExcludedCategories: map[string]bool{"casual": true}}
		picker := newAdvancedMenuTestPicker(withConfig(config), withCategories(categoryRef("casual")))
		restore := withPromptResponses(t, "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedClear(config, []string{"casual"})
	})

	t.Run("update error", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(
			withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", map[string]bool{"casual": true})),
			withCategories(categoryRef("casual")),
			withUpdateError(errors.New("boom")),
		)
		restore := withPromptResponses(t, "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedClear(picker.config.currentConfig, []string{"casual"})
	})

	t.Run("success", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withConfig(mustAdvancedMenuConfig(t, cliTestOutfitRoot, "en", map[string]bool{"casual": true})), withCategories(categoryRef("casual")))
		restore := withPromptResponses(t, "b")
		defer restore()

		AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleExcludedClear(picker.config.currentConfig, []string{"casual"})
		assertCurrentExcluded(t, picker, map[string]bool{})
	})
}

func TestAdvancedMenu_handleResetSettings(t *testing.T) {
	t.Run("cancelled", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker()
		restore := withPromptResponses(t, "b")
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleResetSettings(), menuDestinationAdvanced)
		assertNoFactoryResetRequested(t, picker.stubRuntime)
	})

	t.Run("factory reset error", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker(withFactoryResetError(errors.New("boom")))
		restore := withPromptResponses(t, "y")
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleResetSettings(), menuDestinationAdvanced)
		assertFactoryResetRequested(t, picker.stubRuntime)
	})

	t.Run("success", func(t *testing.T) {
		picker := newAdvancedMenuTestPicker()
		restore := withPromptResponses(t, "y")
		defer restore()

		assertMenuDestination(t, AdvancedMenu{outfitService: newStubOutfitService(picker.stubRuntime)}.handleResetSettings(), menuDestinationExit)
		assertFactoryResetRequested(t, picker.stubRuntime)
	})
}

func TestParseCategorySelections_SkipsEmptyEntries(t *testing.T) {
	got := parseCategorySelections("1, , casual", []string{"casual", "formal"})
	want := []string{"casual"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseCategorySelections() = %v, want %v", got, want)
	}
}

type advancedMenuTestPicker struct {
	*stubRuntime
}

func newAdvancedMenuTestPicker(options ...func(*advancedMenuTestPicker)) *advancedMenuTestPicker {
	picker := &advancedMenuTestPicker{stubRuntime: newStubRuntime()}
	for _, option := range options {
		option(picker)
	}
	return picker
}

func withConfig(config *entities.Config) func(*advancedMenuTestPicker) {
	return func(p *advancedMenuTestPicker) { p.config.currentConfig = config }
}

func withConfigError(err error) func(*advancedMenuTestPicker) {
	return func(p *advancedMenuTestPicker) { p.config.loadErr = err }
}

func withCategories(categories ...entities.CategoryReference) func(*advancedMenuTestPicker) {
	return func(p *advancedMenuTestPicker) { p.wardrobe.categories = categories }
}

func withCategoriesError(err error) func(*advancedMenuTestPicker) {
	return func(p *advancedMenuTestPicker) { p.wardrobe.categoriesErr = err }
}

func withUpdateError(err error) func(*advancedMenuTestPicker) {
	return func(p *advancedMenuTestPicker) { p.config.updateErr = err }
}

func withResetAllError(err error) func(*advancedMenuTestPicker) {
	return func(p *advancedMenuTestPicker) { p.commands.resetAllErr = err }
}

func withResetCategoryError(err error) func(*advancedMenuTestPicker) {
	return func(p *advancedMenuTestPicker) { p.commands.resetCategoryErr = err }
}

func withFactoryResetError(err error) func(*advancedMenuTestPicker) {
	return func(p *advancedMenuTestPicker) { p.commands.factoryResetErr = err }
}

func withCategoryInfos(infos []entities.CategoryInfo) func(*advancedMenuTestPicker) {
	return func(p *advancedMenuTestPicker) { p.wardrobe.categoryInfos = infos }
}

func withRootDirectory(root string) func(*advancedMenuTestPicker) {
	return func(p *advancedMenuTestPicker) { p.wardrobe.rootDirectory = root }
}

func withPromptResponses(t *testing.T, responses ...string) func() {
	t.Helper()
	oldPrompt := promptFunc
	index := 0
	promptFunc = func(string) string {
		if index >= len(responses) {
			t.Fatalf("unexpected prompt after consuming %d responses", index)
		}
		response := responses[index]
		index++
		return response
	}

	return func() {
		promptFunc = oldPrompt
		if index != len(responses) {
			t.Fatalf("consumed %d of %d prompt responses", index, len(responses))
		}
	}
}

func mustAdvancedMenuConfig(t *testing.T, root, language string, excluded map[string]bool) *entities.Config {
	t.Helper()
	config, err := entities.NewConfig(root, &language, excluded, nil, nil)
	if err != nil {
		t.Fatalf("NewConfig(%q) error = %v", root, err)
	}
	return config
}

func assertCurrentConfigRoot(t *testing.T, picker *advancedMenuTestPicker, want string) {
	t.Helper()
	if picker.config.currentConfig == nil {
		t.Fatal("currentConfig = nil")
	}
	if picker.config.currentConfig.Root != want {
		t.Fatalf("currentConfig.Root = %q, want %q", picker.config.currentConfig.Root, want)
	}
}

func assertCurrentConfigLanguage(t *testing.T, picker *advancedMenuTestPicker, want string) {
	t.Helper()
	if picker.config.currentConfig == nil {
		t.Fatal("currentConfig = nil")
	}
	if picker.config.currentConfig.Language != want {
		t.Fatalf("currentConfig.Language = %q, want %q", picker.config.currentConfig.Language, want)
	}
}

func assertCurrentExcluded(t *testing.T, picker *advancedMenuTestPicker, want map[string]bool) {
	t.Helper()
	if picker.config.currentConfig == nil {
		t.Fatal("currentConfig = nil")
	}
	if !reflect.DeepEqual(picker.config.currentConfig.ExcludedCategories, want) {
		t.Fatalf("current excluded categories = %+v, want %+v", picker.config.currentConfig.ExcludedCategories, want)
	}
}

func categoryRef(name string) entities.CategoryReference {
	return entities.NewCategoryReference(name, cliTestCategoryPath(name))
}
