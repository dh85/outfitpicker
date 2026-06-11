package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	domainerrors "github.com/dh85/outfitpicker/internal/domain/errors"
)

func TestMainMenu_Show(t *testing.T) {
	t.Run("file system error opens advanced menu", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.categoryInfoResults = []stubCategoryInfoResult{{err: domainerrors.ErrDirectoryNotFound}}
		menu := newMainMenuForTest(picker)

		assertMenuTransition(t, menuDestinationAdvanced, menu.Show)
	})

	t.Run("category info error returns", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.categoryInfoResults = []stubCategoryInfoResult{{err: domainerrors.NewInvalidInputError("boom")}}
		menu := newMainMenuForTest(picker)

		assertMenuTransition(t, menuDestinationExit, menu.Show)
	})

	t.Run("available categories error returns", func(t *testing.T) {
		infos := []entities.CategoryInfo{entities.NewCategoryInfo(mainMenuCategory("casual"), entities.CategoryStateHasOutfits, 1)}
		picker := newStubRuntime()
		picker.wardrobe.categoryInfoResults = []stubCategoryInfoResult{{infos: infos}, {err: errors.New("boom")}}
		menu := newMainMenuForTest(picker)

		assertMenuTransition(t, menuDestinationExit, menu.Show)
	})

	t.Run("shows with available categories", func(t *testing.T) {
		infos := []entities.CategoryInfo{entities.NewCategoryInfo(mainMenuCategory("casual"), entities.CategoryStateHasOutfits, 1)}
		picker := newStubRuntime()
		picker.wardrobe.categoryInfos = infos
		picker.wardrobe.outfitStates = map[string]entities.CategoryOutfitState{
			"casual": mainMenuState(mainMenuCategory("casual"), []string{"one.avatar"}, []string{"one.avatar"}, nil),
		}
		menu := newMainMenuForTest(picker)
		assertMenuTransitionWithPrompts(t, menuDestinationExit, menu.Show, "q")
	})

	t.Run("shows with no available categories", func(t *testing.T) {
		infos := []entities.CategoryInfo{entities.NewCategoryInfo(mainMenuCategory("docs"), entities.CategoryStateNoAvatarFiles, 0)}
		picker := newStubRuntime()
		picker.wardrobe.categoryInfos = infos
		menu := newMainMenuForTest(picker)
		assertMenuTransitionWithPrompts(t, menuDestinationExit, menu.Show, "q")
	})
}

func TestMainMenu_handleChoice(t *testing.T) {
	t.Run("random choice", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.categoryInfos = []entities.CategoryInfo{entities.NewCategoryInfo(mainMenuCategory("casual"), entities.CategoryStateHasOutfits, 1)}
		picker.random.globalResults = []stubSelectorResult{{outfit: mainMenuOutfitPtr("casual", "one.avatar")}}
		menu := newMainMenuForTest(picker)
		assertMenuTransitionWithPrompts(t, menuDestinationMain, func() menuTransition {
			return menu.handleChoice("random", nil)
		}, "back")
	})

	t.Run("manual choice", func(t *testing.T) {
		picker := newStubRuntime()
		menu := newMainMenuForTest(picker)
		assertMenuTransitionWithPrompts(t, menuDestinationMain, func() menuTransition {
			return menu.handleChoice("manual", nil)
		}, "exit")
	})

	t.Run("worn choice", func(t *testing.T) {
		picker := newStubRuntime()
		menu := newMainMenuForTest(picker)

		assertMenuTransition(t, menuDestinationMain, func() menuTransition {
			return menu.handleChoice("w", nil)
		})
	})

	t.Run("unworn choice", func(t *testing.T) {
		picker := newStubRuntime()
		menu := newMainMenuForTest(picker)

		assertMenuTransition(t, menuDestinationMain, func() menuTransition {
			return menu.handleChoice("u", nil)
		})
	})

	t.Run("advanced choice", func(t *testing.T) {
		picker := newStubRuntime()
		menu := newMainMenuForTest(picker)

		assertMenuTransition(t, menuDestinationAdvanced, func() menuTransition {
			return menu.handleChoice("a", nil)
		})
	})

	t.Run("quit choice", func(t *testing.T) {
		menu := newMainMenuForTest(newStubRuntime())

		assertMenuTransition(t, menuDestinationExit, func() menuTransition {
			return menu.handleChoice("quit", nil)
		})
	})

	t.Run("numeric category choice", func(t *testing.T) {
		casual := entities.NewCategoryInfo(mainMenuCategory("casual"), entities.CategoryStateHasOutfits, 1)
		picker := newStubRuntime()
		picker.wardrobe.categoryInfos = []entities.CategoryInfo{casual}
		picker.wardrobe.outfitStates = map[string]entities.CategoryOutfitState{
			"casual": mainMenuState(mainMenuCategory("casual"), []string{"one.avatar"}, []string{"one.avatar"}, nil),
		}
		menu := newMainMenuForTest(picker)

		got := assertMenuTransition(t, menuDestinationCategory, func() menuTransition {
			return menu.handleChoice("1", []entities.CategoryInfo{casual})
		})
		if got.category.Name != "casual" {
			t.Fatalf("handleChoice() category = %q, want casual", got.category.Name)
		}
	})

	t.Run("invalid choice reshows with next-step hint", func(t *testing.T) {
		picker := newStubRuntime()
		var output strings.Builder
		menu := newMainMenuForTest(picker)
		menu.console = TerminalConsole{stderr: &output}

		assertMenuTransition(t, menuDestinationMain, func() menuTransition {
			return menu.handleChoice("x", nil)
		})
		assertOutputContains(t, output.String(), "Invalid choice", "Enter a number", "R for random", "M for manual", "A for advanced", "Q to quit")
	})
}

func TestMainMenu_handleRandomOutfit(t *testing.T) {
	t.Run("random error reshows", func(t *testing.T) {
		picker := newStubRuntime()
		picker.random.globalResults = []stubSelectorResult{{err: errors.New("boom")}}
		menu := newMainMenuForTest(picker)

		assertMenuTransition(t, menuDestinationMain, menu.handleRandomOutfit)
	})

	t.Run("nil outfit reshows", func(t *testing.T) {
		picker := newStubRuntime()
		picker.random.globalResults = []stubSelectorResult{{outfit: nil}}
		menu := newMainMenuForTest(picker)

		assertMenuTransition(t, menuDestinationMain, menu.handleRandomOutfit)
	})

	t.Run("worn exits with explicit confirmation", func(t *testing.T) {
		picker := newStubRuntime()
		picker.random.globalResults = []stubSelectorResult{{outfit: mainMenuOutfitPtr("casual", "one.avatar")}}
		var output strings.Builder
		menu := newMainMenuForTest(picker)
		menu.console = TerminalConsole{stdout: &output}

		assertMenuTransitionWithPrompts(t, menuDestinationExit, menu.handleRandomOutfit, "w")
		assertOutputContains(t, output.String(), "Marked as worn. Goodbye!")
	})

	t.Run("back reshows", func(t *testing.T) {
		picker := newStubRuntime()
		picker.random.globalResults = []stubSelectorResult{{outfit: mainMenuOutfitPtr("casual", "one.avatar")}}
		menu := newMainMenuForTest(picker)
		assertMenuTransitionWithPrompts(t, menuDestinationMain, menu.handleRandomOutfit, "b")
	})

	t.Run("quit on wear failure", func(t *testing.T) {
		picker := newStubRuntime()
		picker.random.globalResults = []stubSelectorResult{{outfit: mainMenuOutfitPtr("casual", "one.avatar")}}
		picker.commands.wearErr = errors.New("save failed")
		menu := newMainMenuForTest(picker)
		assertMenuTransitionWithPrompts(t, menuDestinationExit, menu.handleRandomOutfit, "w")
	})

	t.Run("skip continues", func(t *testing.T) {
		picker := newStubRuntime()
		picker.random.globalResults = []stubSelectorResult{
			{outfit: mainMenuOutfitPtr("casual", "one.avatar")},
			{outfit: mainMenuOutfitPtr("casual", "two.avatar")},
		}
		menu := newMainMenuForTest(picker)
		assertMenuTransitionWithPrompts(t, menuDestinationExit, menu.handleRandomOutfit, "s", "w")
	})
}

func TestMainMenu_showWornMenu(t *testing.T) {
	t.Run("error reshows", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.allOutfitStatesErr = errors.New("boom")
		menu := newMainMenuForTest(picker)

		assertMenuDestination(t, menu.showWornMenu(), menuDestinationMain)
	})

	t.Run("empty reshows", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.allOutfitStates = map[string]entities.CategoryOutfitState{}
		menu := newMainMenuForTest(picker)

		assertMenuDestination(t, menu.showWornMenu(), menuDestinationMain)
	})

	t.Run("success prompts and reshows", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.allOutfitStates = map[string]entities.CategoryOutfitState{
			"casual": mainMenuState(mainMenuCategory("casual"), []string{"one.avatar"}, nil, []string{"one.avatar"}),
		}
		menu := newMainMenuForTest(picker)
		restore := withPromptResponses(t, "")
		defer restore()

		assertMenuDestination(t, menu.showWornMenu(), menuDestinationMain)
	})
}

func TestMainMenu_showUnwornMenu(t *testing.T) {
	t.Run("error reshows", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.allOutfitStatesErr = errors.New("boom")
		menu := newMainMenuForTest(picker)

		assertMenuDestination(t, menu.showUnwornMenu(), menuDestinationMain)
	})

	t.Run("empty reshows", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.allOutfitStates = map[string]entities.CategoryOutfitState{}
		menu := newMainMenuForTest(picker)

		assertMenuDestination(t, menu.showUnwornMenu(), menuDestinationMain)
	})

	t.Run("success prompts and reshows", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.allOutfitStates = map[string]entities.CategoryOutfitState{
			"casual": mainMenuState(mainMenuCategory("casual"), []string{"one.avatar"}, []string{"one.avatar"}, nil),
		}
		menu := newMainMenuForTest(picker)
		restore := withPromptResponses(t, "")
		defer restore()

		assertMenuDestination(t, menu.showUnwornMenu(), menuDestinationMain)
	})
}

func TestMainMenu_handleManualSelection(t *testing.T) {
	t.Run("categories error reshows", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.categoriesErr = errors.New("boom")
		menu := newMainMenuForTest(picker)

		assertMenuDestination(t, menu.handleManualSelection(), menuDestinationMain)
	})

	t.Run("category q reshows main menu", func(t *testing.T) {
		picker := newStubRuntime()
		menu := newMainMenuForTest(picker)
		restore := withPromptResponses(t, "back")
		defer restore()

		assertMenuDestination(t, menu.handleManualSelection(), menuDestinationMain)
	})

	t.Run("invalid category choice recurses", func(t *testing.T) {
		picker := newStubRuntime()
		menu := newMainMenuForTest(picker)
		restore := withPromptResponses(t, "x", "q")
		defer restore()

		assertMenuDestination(t, menu.handleManualSelection(), menuDestinationMain)
	})

	t.Run("show all outfits error reshows", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.showAllOutfitsErr = errors.New("boom")
		menu := newMainMenuForTest(picker)
		restore := withPromptResponses(t, "1")
		defer restore()

		assertMenuDestination(t, menu.handleManualSelection(), menuDestinationMain)
	})

	t.Run("no outfits continues with next-step hint", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.allOutfitsByCategory = map[string][]entities.OutfitReference{"casual": nil}
		var output strings.Builder
		menu := newMainMenuForTest(picker)
		menu.console = TerminalConsole{stdin: strings.NewReader("1\nq\n"), stdout: &output}

		assertMenuDestination(t, menu.handleManualSelection(), menuDestinationMain)
		assertOutputContains(t, output.String(), "No outfits found in casual", "Add .avatar files to:", cliTestCategoryPath("casual"))
	})

	t.Run("outfit state error reshows", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.allOutfitsByCategory = map[string][]entities.OutfitReference{"casual": {mainMenuOutfit("casual", "one.avatar")}}
		picker.wardrobe.outfitStateErrors = map[string]error{"casual": errors.New("boom")}
		menu := newMainMenuForTest(picker)
		restore := withPromptResponses(t, "1")
		defer restore()

		assertMenuDestination(t, menu.handleManualSelection(), menuDestinationMain)
	})

	t.Run("outfit q recurses to category choice", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.allOutfitsByCategory = map[string][]entities.OutfitReference{"casual": {mainMenuOutfit("casual", "one.avatar")}}
		picker.wardrobe.outfitStates = map[string]entities.CategoryOutfitState{"casual": mainMenuState(mainMenuCategory("casual"), []string{"one.avatar"}, []string{"one.avatar"}, nil)}
		menu := newMainMenuForTest(picker)
		restore := withPromptResponses(t, "1", "back", "exit")
		defer restore()

		assertMenuDestination(t, menu.handleManualSelection(), menuDestinationMain)
	})

	t.Run("invalid outfit choice recurses", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.allOutfitsByCategory = map[string][]entities.OutfitReference{"casual": {mainMenuOutfit("casual", "one.avatar")}}
		picker.wardrobe.outfitStates = map[string]entities.CategoryOutfitState{"casual": mainMenuState(mainMenuCategory("casual"), []string{"one.avatar"}, []string{"one.avatar"}, nil)}
		menu := newMainMenuForTest(picker)
		restore := withPromptResponses(t, "1", "x", "q")
		defer restore()

		assertMenuDestination(t, menu.handleManualSelection(), menuDestinationMain)
	})

	t.Run("skipped selection recurses", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.allOutfitsByCategory = map[string][]entities.OutfitReference{"casual": {mainMenuOutfit("casual", "one.avatar")}}
		picker.wardrobe.outfitStates = map[string]entities.CategoryOutfitState{"casual": mainMenuState(mainMenuCategory("casual"), []string{"one.avatar"}, []string{"one.avatar"}, nil)}
		menu := newMainMenuForTest(picker)
		restore := withPromptResponses(t, "1", "1", "n", "q")
		defer restore()

		assertMenuDestination(t, menu.handleManualSelection(), menuDestinationMain)
	})

	t.Run("non skipped result exits", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.allOutfitsByCategory = map[string][]entities.OutfitReference{"casual": {mainMenuOutfit("casual", "one.avatar")}}
		picker.wardrobe.outfitStates = map[string]entities.CategoryOutfitState{"casual": mainMenuState(mainMenuCategory("casual"), []string{"one.avatar"}, []string{"one.avatar"}, nil)}
		menu := newMainMenuForTest(picker)
		restore := withPromptResponses(t, "1", "1", "q")
		defer restore()

		assertMenuDestination(t, menu.handleManualSelection(), menuDestinationExit)
	})
}

func newMainMenuForTest(picker *stubRuntime) MainMenu {
	if picker.wardrobe.rootDirectory == "" {
		picker.wardrobe.rootDirectory = cliTestOutfitRoot
	}
	if picker.wardrobe.categories == nil {
		picker.wardrobe.categories = []entities.CategoryReference{mainMenuCategory("casual")}
	}
	return MainMenu{
		outfitService: newStubOutfitService(picker),
		selector:      picker.random,
		presentation:  NewOutfitPresentation(picker.commands),
		renderer:      MenuRenderer{},
	}
}

func mainMenuCategory(name string) entities.CategoryReference {
	return entities.NewCategoryReference(name, cliTestCategoryPath(name))
}

func mainMenuOutfit(categoryName, fileName string) entities.OutfitReference {
	return entities.NewOutfitReference(fileName, mainMenuCategory(categoryName))
}

func mainMenuOutfitPtr(categoryName, fileName string) *entities.OutfitReference {
	outfit := mainMenuOutfit(categoryName, fileName)
	return &outfit
}

func mainMenuState(category entities.CategoryReference, all, available, worn []string) entities.CategoryOutfitState {
	return entities.NewCategoryOutfitState(
		category,
		mainMenuOutfits(category, all),
		mainMenuOutfits(category, available),
		mainMenuOutfits(category, worn),
	)
}

func mainMenuOutfits(category entities.CategoryReference, names []string) []entities.OutfitReference {
	outfits := make([]entities.OutfitReference, 0, len(names))
	for _, name := range names {
		outfits = append(outfits, entities.NewOutfitReference(name, category))
	}
	return outfits
}
