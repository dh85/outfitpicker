package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	domainerrors "github.com/dh85/outfitpicker/internal/domain/errors"
)

func TestOutfitPresentation_PresentOutfitWithCategoryChoice(t *testing.T) {
	t.Run("wear choice", func(t *testing.T) {
		commands := &stubCommandHandler{}
		var output strings.Builder
		presentation := NewOutfitPresentation(commands, TerminalConsole{stdin: strings.NewReader("wear\n"), stdout: &output})

		got := presentation.PresentOutfitWithCategoryChoice(outfitPresentationOutfit("casual", "one.avatar"), "casual")
		if got != OutfitChoiceWorn {
			t.Fatalf("PresentOutfitWithCategoryChoice() = %v, want %v", got, OutfitChoiceWorn)
		}
		if len(commands.wearCalls) != 1 || commands.wearCalls[0].FileName != "one.avatar" {
			t.Fatalf("wearCalls = %#v, want one.avatar worn", commands.wearCalls)
		}
		assertOutputContains(t, output.String(), "👗 one.avatar", "📁 casual", "[W] Mark worn and quit", "[S] Skip", "[B] Back", "[Q] Quit")
	})

	t.Run("invalid then back", func(t *testing.T) {
		presentation := NewOutfitPresentation(&stubCommandHandler{})
		restore := withPromptResponses(t, "x", "b")
		defer restore()

		got := presentation.PresentOutfitWithCategoryChoice(outfitPresentationOutfit("casual", "one.avatar"), "casual")
		if got != OutfitChoiceBack {
			t.Fatalf("PresentOutfitWithCategoryChoice() = %v, want %v", got, OutfitChoiceBack)
		}
	})
}

func TestOutfitPresentation_PresentManualOutfit(t *testing.T) {
	outfit := outfitPresentationOutfit("casual", "one.avatar")

	t.Run("yes delegates to wear choice", func(t *testing.T) {
		commands := &stubCommandHandler{}
		presentation := NewOutfitPresentation(commands)
		restore := withPromptResponses(t, "yes")
		defer restore()

		got := presentation.PresentManualOutfit(outfit, "casual", false)
		if got != OutfitChoiceWorn {
			t.Fatalf("PresentManualOutfit() = %v, want %v", got, OutfitChoiceWorn)
		}
	})

	t.Run("no skips", func(t *testing.T) {
		presentation := NewOutfitPresentation(&stubCommandHandler{})
		restore := withPromptResponses(t, "no")
		defer restore()

		got := presentation.PresentManualOutfit(outfit, "casual", false)
		if got != OutfitChoiceSkipped {
			t.Fatalf("PresentManualOutfit() = %v, want %v", got, OutfitChoiceSkipped)
		}
	})

	t.Run("quit returns quit", func(t *testing.T) {
		presentation := NewOutfitPresentation(&stubCommandHandler{})
		restore := withPromptResponses(t, "exit")
		defer restore()

		got := presentation.PresentManualOutfit(outfit, "casual", false)
		if got != OutfitChoiceQuit {
			t.Fatalf("PresentManualOutfit() = %v, want %v", got, OutfitChoiceQuit)
		}
	})

	t.Run("back returns back", func(t *testing.T) {
		presentation := NewOutfitPresentation(&stubCommandHandler{})
		restore := withPromptResponses(t, "back")
		defer restore()

		got := presentation.PresentManualOutfit(outfit, "casual", false)
		if got != OutfitChoiceBack {
			t.Fatalf("PresentManualOutfit() = %v, want %v", got, OutfitChoiceBack)
		}
	})

	t.Run("invalid then no re-prompts", func(t *testing.T) {
		presentation := NewOutfitPresentation(&stubCommandHandler{})
		restore := withPromptResponses(t, "bad", "n")
		defer restore()

		got := presentation.PresentManualOutfit(outfit, "casual", true)
		if got != OutfitChoiceSkipped {
			t.Fatalf("PresentManualOutfit() = %v, want %v", got, OutfitChoiceSkipped)
		}
	})
}

func TestOutfitPresentation_PresentOutfitWithChoice(t *testing.T) {
	outfit := outfitPresentationOutfit("casual", "one.avatar")

	t.Run("skip returns skipped", func(t *testing.T) {
		presentation := NewOutfitPresentation(&stubCommandHandler{})
		restore := withPromptResponses(t, "skip")
		defer restore()

		got := presentation.PresentOutfitWithChoice(outfit)
		if got != OutfitChoiceSkipped {
			t.Fatalf("PresentOutfitWithChoice() = %v, want %v", got, OutfitChoiceSkipped)
		}
	})

	t.Run("back returns back", func(t *testing.T) {
		presentation := NewOutfitPresentation(&stubCommandHandler{})
		restore := withPromptResponses(t, "back")
		defer restore()

		got := presentation.PresentOutfitWithChoice(outfit)
		if got != OutfitChoiceBack {
			t.Fatalf("PresentOutfitWithChoice() = %v, want %v", got, OutfitChoiceBack)
		}
	})

	t.Run("quit returns quit", func(t *testing.T) {
		presentation := NewOutfitPresentation(&stubCommandHandler{})
		restore := withPromptResponses(t, "quit")
		defer restore()

		got := presentation.PresentOutfitWithChoice(outfit)
		if got != OutfitChoiceQuit {
			t.Fatalf("PresentOutfitWithChoice() = %v, want %v", got, OutfitChoiceQuit)
		}
	})

	t.Run("invalid then skip re-prompts", func(t *testing.T) {
		presentation := NewOutfitPresentation(&stubCommandHandler{})
		restore := withPromptResponses(t, "bad", "n")
		defer restore()

		got := presentation.PresentOutfitWithChoice(outfit)
		if got != OutfitChoiceSkipped {
			t.Fatalf("PresentOutfitWithChoice() = %v, want %v", got, OutfitChoiceSkipped)
		}
	})
}

func TestOutfitPresentation_handleWearChoice(t *testing.T) {
	outfit := outfitPresentationOutfit("casual", "one.avatar")

	t.Run("success", func(t *testing.T) {
		commands := &stubCommandHandler{}
		presentation := NewOutfitPresentation(commands)

		got := presentation.handleWearChoice(outfit)
		if got != OutfitChoiceWorn {
			t.Fatalf("handleWearChoice() = %v, want %v", got, OutfitChoiceWorn)
		}
		if len(commands.wearCalls) != 1 || commands.wearCalls[0].FileName != "one.avatar" {
			t.Fatalf("wearCalls = %#v, want one.avatar worn", commands.wearCalls)
		}
	})

	t.Run("rotation complete still returns worn", func(t *testing.T) {
		commands := &stubCommandHandler{wearErr: domainerrors.NewRotationCompletedError("casual")}
		presentation := NewOutfitPresentation(commands)

		got := presentation.handleWearChoice(outfit)
		if got != OutfitChoiceWorn {
			t.Fatalf("handleWearChoice() = %v, want %v", got, OutfitChoiceWorn)
		}
	})

	t.Run("other error returns quit", func(t *testing.T) {
		commands := &stubCommandHandler{wearErr: errors.New("save failed")}
		presentation := NewOutfitPresentation(commands)

		got := presentation.handleWearChoice(outfit)
		if got != OutfitChoiceQuit {
			t.Fatalf("handleWearChoice() = %v, want %v", got, OutfitChoiceQuit)
		}
	})
}

func outfitPresentationOutfit(categoryName, fileName string) entities.OutfitReference {
	category := entities.NewCategoryReference(categoryName, cliTestCategoryPath(categoryName))
	return entities.NewOutfitReference(fileName, category)
}
