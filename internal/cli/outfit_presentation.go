package cli

import (
	"errors"
	"fmt"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	domainerrors "github.com/dh85/outfitpicker/internal/domain/errors"
)

type OutfitPresentation struct {
	commands OutfitCommandHandler
	console  Console
}

func NewOutfitPresentation(commands OutfitCommandHandler, consoles ...Console) OutfitPresentation {
	return OutfitPresentation{commands: commands, console: optionalConsole(consoles)}
}

func (p OutfitPresentation) terminal() Console {
	return consoleOrDefault(p.console)
}

func (p OutfitPresentation) PresentOutfitWithCategoryChoice(outfit entities.OutfitReference, category string) OutfitChoice {
	return p.presentOutfit(outfit, category)
}

func (p OutfitPresentation) PresentOutfitWithChoice(outfit entities.OutfitReference) OutfitChoice {
	return p.presentOutfit(outfit, "")
}

func (p OutfitPresentation) PresentManualOutfit(outfit entities.OutfitReference, category string, isWorn bool) OutfitChoice {
	cleanName := displayOutfitName(outfit.FileName)
	safeCategory := sanitizeTerminalText(category)
	wornText := ""
	if isWorn {
		wornText = " (already worn)"
	}

	p.terminal().Printf("\nYou selected: %s from %s%s\n", cleanName, safeCategory, wornText)
	promptText := "Wear this outfit? (y)es, (n)o, or (q)uit? "
	if isWorn {
		promptText = "Wear outfit again? (y)es, (n)o, or (q)uit? "
	}

	input := p.terminal().Prompt(promptText)
	switch {
	case isYesInput(input):
		return p.handleWearChoice(outfit)
	case isNoInput(input):
		return OutfitChoiceSkipped
	case isBackInput(input):
		return OutfitChoiceBack
	case isQuitInput(input):
		return OutfitChoiceQuit
	default:
		p.terminal().Error("Please enter 'y' or 'yes' for yes, 'n' or 'no' for no, 'b' or 'back' to go back, or 'q', 'quit', or 'exit' to quit.")
		return p.PresentManualOutfit(outfit, category, isWorn)
	}
}

func (p OutfitPresentation) presentOutfit(outfit entities.OutfitReference, categoryContext string) OutfitChoice {
	categoryName := categoryContext
	if categoryName == "" {
		categoryName = outfit.Category.Name
	}

	p.terminal().Printf("\n👗 %s\n", sanitizeTerminalText(outfit.FileName))
	p.terminal().Printf("📁 %s\n\n", sanitizeTerminalText(categoryName))
	p.terminal().Println("[W] Mark worn and quit")
	p.terminal().Println("[S] Skip")
	p.terminal().Println("[B] Back")
	p.terminal().Println("[Q] Quit")

	input := normalizeChoiceInput(p.terminal().Prompt("Choose an option: "))
	switch input {
	case "w", "wear", "worn", "mark worn", "mark", "y", "yes":
		return p.handleWearChoice(outfit)
	case "s", "skip", "n", "no":
		p.terminal().Printf("Skipped: %s\n", sanitizeTerminalText(outfit.FileName))
		return OutfitChoiceSkipped
	case "b", "back":
		return OutfitChoiceBack
	case "q", "quit", "exit":
		return OutfitChoiceQuit
	default:
		p.terminal().Error("Please enter 'w' to mark worn and quit, 's' to skip, 'b' to go back, or 'q' to quit.")
		return p.presentOutfit(outfit, categoryContext)
	}
}

func (p OutfitPresentation) handleWearChoice(outfit entities.OutfitReference) OutfitChoice {
	err := p.commands.WearOutfit(outfit)
	if err == nil {
		return OutfitChoiceWorn
	}

	var rotationCompleted *domainerrors.RotationCompletedError
	if errors.As(err, &rotationCompleted) {
		p.terminal().Success(fmt.Sprintf("You have now worn all outfits in %s.", rotationCompleted.Category))
		p.terminal().Info(fmt.Sprintf("Use reset category or reset all to make %s available again.", rotationCompleted.Category))
		return OutfitChoiceWorn
	}

	p.terminal().Error("Could not save this outfit. Please try again.")
	return OutfitChoiceQuit
}
