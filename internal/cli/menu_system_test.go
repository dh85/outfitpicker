package cli

import (
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestNewMenuSystem(t *testing.T) {
	picker := newStubRuntime()
	outfitService := newStubOutfitService(picker)
	presentation := NewOutfitPresentation(picker.commands)
	renderer := NewMenuRenderer()

	system := NewMenuSystem(outfitService, picker.random, presentation, renderer)

	if system.outfitService != outfitService {
		t.Fatal("NewMenuSystem() did not retain the provided outfit service")
	}
}

func TestMenuSystem_ShowMainMenu(t *testing.T) {
	t.Run("quit from main menu", func(t *testing.T) {
		system := newMenuSystemForRuntime(newStubRuntime())
		restore := withPromptResponses(t, "q")
		defer restore()

		system.ShowMainMenu()
	})

	t.Run("routes through category menu", func(t *testing.T) {
		picker := newStubRuntime()
		casual := entities.NewCategoryInfo(entities.NewCategoryReference("casual", cliTestCategoryPath("casual")), entities.CategoryStateHasOutfits, 1)
		picker.wardrobe.categoryInfos = []entities.CategoryInfo{casual}
		picker.wardrobe.outfitStates = map[string]entities.CategoryOutfitState{
			"casual": entities.NewCategoryOutfitState(
				casual.Category,
				[]entities.OutfitReference{entities.NewOutfitReference("one.avatar", casual.Category)},
				[]entities.OutfitReference{entities.NewOutfitReference("one.avatar", casual.Category)},
				nil,
			),
		}

		system := newMenuSystemForRuntime(picker)
		restore := withPromptResponses(t, "1", "b", "q")
		defer restore()

		system.ShowMainMenu()
	})

	t.Run("routes through advanced menu", func(t *testing.T) {
		system := newMenuSystemForRuntime(newStubRuntime())
		restore := withPromptResponses(t, "a", "b", "q")
		defer restore()

		system.ShowMainMenu()
	})
}

func TestMenuSystem_dispatchTransition_DefaultStopsLoop(t *testing.T) {
	picker := newStubRuntime()
	system := newMenuSystemForRuntime(picker)

	_, ok := system.dispatchTransition(menuTransition{destination: menuDestination(999)})
	if ok {
		t.Fatal("dispatchTransition() ok = true, want false for unknown destination")
	}
}
