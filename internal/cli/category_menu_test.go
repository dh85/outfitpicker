package cli

import (
	"errors"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestBuildCategoryMenuView_NormalCategory(t *testing.T) {
	category := entities.NewCategoryReference("casual", cliTestCategoryPath("casual"))
	state := entities.NewCategoryOutfitState(
		category,
		[]entities.OutfitReference{
			entities.NewOutfitReference("one.avatar", category),
			entities.NewOutfitReference("two.avatar", category),
		},
		[]entities.OutfitReference{
			entities.NewOutfitReference("two.avatar", category),
		},
		[]entities.OutfitReference{
			entities.NewOutfitReference("one.avatar", category),
		},
	)

	view := buildCategoryMenuView(category.Name, state)

	if view.exhausted {
		t.Fatal("expected normal category menu, got exhausted menu")
	}
	if view.defaultAction != categoryMenuActionPick {
		t.Fatalf("defaultAction = %v, want pick", view.defaultAction)
	}
	if len(view.options) != 2 {
		t.Fatalf("options count = %d, want 2", len(view.options))
	}
	if view.options[0] != "  [P] Pick random outfit (default)" {
		t.Fatalf("first option = %q", view.options[0])
	}
	if view.options[1] != "  [B] Back" {
		t.Fatalf("second option = %q", view.options[1])
	}
	if view.message != "" {
		t.Fatalf("message = %q, want empty", view.message)
	}
}

func TestBuildCategoryMenuView_ExhaustedCategory(t *testing.T) {
	category := entities.NewCategoryReference("casual", cliTestCategoryPath("casual"))
	state := entities.NewCategoryOutfitState(
		category,
		[]entities.OutfitReference{
			entities.NewOutfitReference("one.avatar", category),
			entities.NewOutfitReference("two.avatar", category),
		},
		nil,
		[]entities.OutfitReference{
			entities.NewOutfitReference("one.avatar", category),
			entities.NewOutfitReference("two.avatar", category),
		},
	)

	view := buildCategoryMenuView(category.Name, state)

	if !view.exhausted {
		t.Fatal("expected exhausted category menu")
	}
	if view.defaultAction != categoryMenuActionBack {
		t.Fatalf("defaultAction = %v, want back", view.defaultAction)
	}
	if view.message != "All outfits in casual have been worn." {
		t.Fatalf("message = %q", view.message)
	}
	if len(view.options) != 2 {
		t.Fatalf("options count = %d, want 2", len(view.options))
	}
	if view.options[0] != "  [R] Reset category and pick a random outfit" {
		t.Fatalf("first option = %q", view.options[0])
	}
	if view.options[1] != "  [B] Back (default)" {
		t.Fatalf("second option = %q", view.options[1])
	}
}

func TestResolveCategoryMenuAction(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		exhausted bool
		want      categoryMenuAction
	}{
		{name: "normal default pick", input: "", want: categoryMenuActionPick},
		{name: "normal explicit pick", input: "p", want: categoryMenuActionPick},
		{name: "normal back", input: "b", want: categoryMenuActionBack},
		{name: "normal invalid", input: "x", want: categoryMenuActionInvalid},
		{name: "exhausted default back", input: "", exhausted: true, want: categoryMenuActionBack},
		{name: "exhausted explicit back", input: "b", exhausted: true, want: categoryMenuActionBack},
		{name: "exhausted reset and pick", input: "r", exhausted: true, want: categoryMenuActionResetAndPick},
		{name: "exhausted rejects pick", input: "p", exhausted: true, want: categoryMenuActionInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveCategoryMenuAction(tt.input, tt.exhausted)
			if got != tt.want {
				t.Fatalf("resolveCategoryMenuAction(%q, %t) = %v, want %v", tt.input, tt.exhausted, got, tt.want)
			}
		})
	}
}

func TestCategoryMenu_Show(t *testing.T) {
	t.Run("state error", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.outfitStateErr = errors.New("boom")
		menu := newCategoryMenuForTest(picker, "casual", nil)

		menu.Show()
	})

	t.Run("pick default enters outfit loop", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.outfitState = categoryMenuState("casual", []string{"one.avatar"}, []string{"one.avatar"}, nil)
		picker.random.categoryResults = []stubSelectorResult{{outfit: outfitPtr("casual", "one.avatar")}}
		menu := newCategoryMenuForTest(picker, "casual", nil)
		restore := withPromptResponses(t, "", "w")
		defer restore()

		menu.Show()

		assertCategoryRandomRequestCount(t, picker, 1)
		assertWearRequested(t, picker, "one.avatar")
	})

	t.Run("back returns to main menu", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.outfitState = categoryMenuState("casual", []string{"one.avatar"}, []string{"one.avatar"}, nil)
		menu := newCategoryMenuForTest(picker, "casual", nil)
		assertMenuTransitionWithPrompts(t, menuDestinationMain, menu.Show, "b")
	})

	t.Run("exhausted reset error", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.outfitState = categoryMenuState("casual", []string{"one.avatar"}, nil, []string{"one.avatar"})
		picker.commands.resetCategoryErr = errors.New("reset failed")
		menu := newCategoryMenuForTest(picker, "casual", nil)
		restore := withPromptResponses(t, "r")
		defer restore()

		menu.Show()

		assertCategoryResetRequested(t, picker, "casual")
		assertNoCategoryRandomRequested(t, picker)
	})

	t.Run("exhausted reset success enters outfit loop", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.outfitState = categoryMenuState("casual", []string{"one.avatar"}, nil, []string{"one.avatar"})
		picker.random.categoryResults = []stubSelectorResult{{outfit: outfitPtr("casual", "one.avatar")}}
		menu := newCategoryMenuForTest(picker, "casual", nil)
		restore := withPromptResponses(t, "r", "w")
		defer restore()

		menu.Show()

		assertCategoryResetRequested(t, picker, "casual")
		assertWearRequested(t, picker, "one.avatar")
	})

	t.Run("invalid choice re-prompts", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.outfitState = categoryMenuState("casual", []string{"one.avatar"}, []string{"one.avatar"}, nil)
		menu := newCategoryMenuForTest(picker, "casual", nil)
		assertMenuTransitionWithPrompts(t, menuDestinationCategory, menu.Show, "x")
	})
}

func TestCategoryMenu_handleOutfitLoop(t *testing.T) {
	t.Run("random error returns to main menu", func(t *testing.T) {
		picker := newStubRuntime()
		picker.random.categoryResults = []stubSelectorResult{{err: errors.New("boom")}}
		menu := newCategoryMenuForTest(picker, "casual", nil)

		assertMenuTransition(t, menuDestinationMain, menu.handleOutfitLoop)
	})

	t.Run("nil outfit returns to main menu", func(t *testing.T) {
		picker := newStubRuntime()
		picker.random.categoryResults = []stubSelectorResult{{outfit: nil}}
		menu := newCategoryMenuForTest(picker, "casual", nil)

		assertMenuTransition(t, menuDestinationMain, menu.handleOutfitLoop)
	})

	t.Run("skip continues until back", func(t *testing.T) {
		picker := newStubRuntime()
		picker.random.categoryResults = []stubSelectorResult{
			{outfit: outfitPtr("casual", "one.avatar")},
			{outfit: outfitPtr("casual", "two.avatar")},
		}
		menu := newCategoryMenuForTest(picker, "casual", nil)
		got := assertMenuTransitionWithPrompts(t, menuDestinationMain, menu.handleOutfitLoop, "s", "b")

		assertCategoryRandomRequestCount(t, picker, 2)
		_ = got
	})

	t.Run("wear success exits", func(t *testing.T) {
		picker := newStubRuntime()
		picker.random.categoryResults = []stubSelectorResult{{outfit: outfitPtr("casual", "one.avatar")}}
		menu := newCategoryMenuForTest(picker, "casual", nil)
		assertMenuTransitionWithPrompts(t, menuDestinationExit, menu.handleOutfitLoop, "w")

		assertWearRequested(t, picker, "one.avatar")
	})

	t.Run("wear failure quits", func(t *testing.T) {
		picker := newStubRuntime()
		picker.random.categoryResults = []stubSelectorResult{{outfit: outfitPtr("casual", "one.avatar")}}
		picker.commands.wearErr = errors.New("save failed")
		menu := newCategoryMenuForTest(picker, "casual", nil)
		assertMenuTransitionWithPrompts(t, menuDestinationExit, menu.handleOutfitLoop, "w")
	})
}

func newCategoryMenuForTest(picker *stubRuntime, categoryName string, infos []entities.CategoryInfo) CategoryMenu {
	if picker.wardrobe.rootDirectory == "" {
		picker.wardrobe.rootDirectory = cliTestOutfitRoot
	}
	if picker.wardrobe.categoryInfos == nil {
		picker.wardrobe.categoryInfos = infos
	}
	return CategoryMenu{
		outfitService: newStubOutfitService(picker),
		selector:      picker.random,
		presentation:  NewOutfitPresentation(picker.commands),
		category:      entities.NewCategoryReference(categoryName, picker.wardrobe.rootDirectory+"/"+categoryName),
	}
}

func outfitPtr(categoryName, fileName string) *entities.OutfitReference {
	outfit := entities.NewOutfitReference(fileName, entities.NewCategoryReference(categoryName, cliTestCategoryPath(categoryName)))
	return &outfit
}

func categoryMenuState(categoryName string, all, available, worn []string) entities.CategoryOutfitState {
	category := entities.NewCategoryReference(categoryName, cliTestCategoryPath(categoryName))
	return entities.NewCategoryOutfitState(
		category,
		categoryMenuOutfits(category, all),
		categoryMenuOutfits(category, available),
		categoryMenuOutfits(category, worn),
	)
}

func categoryMenuOutfits(category entities.CategoryReference, names []string) []entities.OutfitReference {
	outfits := make([]entities.OutfitReference, 0, len(names))
	for _, name := range names {
		outfits = append(outfits, entities.NewOutfitReference(name, category))
	}
	return outfits
}
