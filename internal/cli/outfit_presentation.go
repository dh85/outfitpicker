package cli

import (
	"errors"
	"fmt"
	"strings"

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

	input := strings.ToLower(strings.TrimSpace(p.terminal().Prompt(promptText)))
	switch input {
	case "y":
		return p.handleWearChoice(outfit)
	case "n":
		return OutfitChoiceSkipped
	case "q":
		return OutfitChoiceQuit
	default:
		p.terminal().Error("Please enter 'y' for yes, 'n' for no, or 'q' to quit.")
		return p.PresentManualOutfit(outfit, category, isWorn)
	}
}

func (p OutfitPresentation) presentOutfit(outfit entities.OutfitReference, categoryContext string) OutfitChoice {
	cleanName := displayOutfitName(outfit.FileName)
	if categoryContext != "" {
		p.terminal().Printf("\n✨ %s %s %s\n", Accent("I picked this outfit for you:"), cleanName, Dim("(from "+sanitizeTerminalText(categoryContext)+")"))
	} else {
		p.terminal().Printf("\n✨ %s %s\n", Accent("I picked this outfit for you:"), cleanName)
	}

	input := strings.ToLower(strings.TrimSpace(p.terminal().Prompt("Do you want to (w)ear it, (s)kip it, or go (b)ack? ")))
	switch input {
	case "w":
		return p.handleWearChoice(outfit)
	case "s":
		p.terminal().Printf("Skipped: %s\n", cleanName)
		return OutfitChoiceSkipped
	case "b":
		return OutfitChoiceBack
	default:
		p.terminal().Error("Please enter 'w' to wear, 's' to skip, or 'b' to go back.")
		return p.presentOutfit(outfit, categoryContext)
	}
}

func (p OutfitPresentation) handleWearChoice(outfit entities.OutfitReference) OutfitChoice {
	err := p.commands.WearOutfit(outfit)
	if err == nil {
		p.terminal().Success(fmt.Sprintf("Saved %s to worn outfits.", strings.TrimSuffix(outfit.FileName, ".avatar")))
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
