package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	domainerrors "github.com/dh85/outfitpicker/internal/domain/errors"
)

type MainMenu struct {
	outfitService OutfitService
	selector      RandomOutfitSelector
	presentation  OutfitPresentation
	renderer      MenuRenderer
	console       Console
}

func (m MainMenu) terminal() Console { return consoleOrDefault(m.console) }

func (m MainMenu) Show() menuTransition {
	m.renderer.ShowTitle()
	m.showOutfitDirectory()

	categoryInfos, err := m.outfitService.GetCategoryInfo()
	if err != nil {
		if errors.Is(domainerrors.MapError(err), domainerrors.ErrFileSystem) {
			m.terminal().Error("Can't find your outfit folder")
			m.terminal().Println("Use Advanced Settings > Change outfit path to fix this")
			return advancedMenuTransition()
		}
		m.terminal().Error(fmt.Sprintf("Error listing categories: %v", err))
		return exitMenuTransition()
	}

	availableCategories, err := m.outfitService.GetAvailableCategories()
	if err != nil {
		m.terminal().Error(fmt.Sprintf("Error listing available categories: %v", err))
		return exitMenuTransition()
	}

	if len(availableCategories) > 0 {
		m.renderer.ShowAvailableCategories(availableCategories, m.outfitService)
	}
	m.renderer.ShowUnavailableCategories(categoryInfos, m.outfitService)
	m.terminal().Println()
	m.renderer.ShowMenuOptions()

	input := m.terminal().Prompt("Choose a number or letter: ")
	return m.handleChoice(strings.ToLower(strings.TrimSpace(input)), availableCategories)
}

func (m MainMenu) showOutfitDirectory() {
	rootPath, err := m.outfitService.GetRootDirectory()
	if err == nil {
		m.renderer.ShowOutfitDirectory(rootPath)
	}
}

func (m MainMenu) handleChoice(input string, availableCategories []entities.CategoryInfo) menuTransition {
	if choice, ok := ParseMenuChoice(input); ok {
		switch choice {
		case MenuChoiceRandom:
			return m.handleRandomOutfit()
		case MenuChoiceManual:
			return m.handleManualSelection()
		case MenuChoiceWorn:
			return m.showWornMenu()
		case MenuChoiceUnworn:
			return m.showUnwornMenu()
		case MenuChoiceAdvanced:
			return advancedMenuTransition()
		case MenuChoiceQuit:
			m.terminal().Println("Goodbye!")
			return exitMenuTransition()
		}
	}

	index, err := strconv.Atoi(input)
	if err == nil && index > 0 && index <= len(availableCategories) {
		info := availableCategories[index-1]
		return categoryMenuTransition(info.Category)
	}

	m.terminal().Error("Invalid choice")
	return mainMenuTransition()
}

func (m MainMenu) handleRandomOutfit() menuTransition {
	for {
		randomOutfit, err := m.selector.ShowNextUniqueRandomOutfit()
		if err != nil {
			m.terminal().Error(fmt.Sprintf("Error: %v", err))
			return mainMenuTransition()
		}
		if randomOutfit == nil {
			m.terminal().Info("No outfits available")
			return mainMenuTransition()
		}

		result := m.presentation.PresentOutfitWithCategoryChoice(*randomOutfit, randomOutfit.Category.Name)
		switch result {
		case OutfitChoiceWorn:
			m.terminal().Println("Goodbye!")
			return exitMenuTransition()
		case OutfitChoiceBack:
			return mainMenuTransition()
		case OutfitChoiceQuit:
			m.terminal().Println("Goodbye!")
			return exitMenuTransition()
		case OutfitChoiceSkipped:
			continue
		}
	}
}

func (m MainMenu) showWornMenu() menuTransition {
	wornOutfits, err := m.outfitService.GetWornOutfits()
	if err != nil {
		m.terminal().Error(fmt.Sprintf("Error loading worn outfits: %v", err))
		return mainMenuTransition()
	}
	if len(wornOutfits) == 0 {
		m.terminal().Info("No worn outfits found")
		return mainMenuTransition()
	}
	m.renderer.ShowWornOutfits(wornOutfits)
	m.terminal().Prompt("Press Enter to return to main menu: ")
	return mainMenuTransition()
}

func (m MainMenu) showUnwornMenu() menuTransition {
	unwornOutfits, err := m.outfitService.GetUnwornOutfits()
	if err != nil {
		m.terminal().Error(fmt.Sprintf("Error loading unworn outfits: %v", err))
		return mainMenuTransition()
	}
	if len(unwornOutfits) == 0 {
		m.terminal().Info("No unworn outfits found")
		return mainMenuTransition()
	}
	m.renderer.ShowUnwornOutfits(unwornOutfits)
	m.terminal().Prompt("Press Enter to return to main menu: ")
	return mainMenuTransition()
}

func (m MainMenu) handleManualSelection() menuTransition {
	for {
		categories, err := m.outfitService.GetCategories()
		if err != nil {
			m.terminal().Error(fmt.Sprintf("Error: %v", err))
			return mainMenuTransition()
		}

		m.renderer.ShowManualSelectionCategories(categories)
		categoryInput := strings.ToLower(strings.TrimSpace(m.terminal().Prompt(fmt.Sprintf("\nChoose a category (1-%d) or 'q' to go back: ", len(categories)))))
		if categoryInput == "q" {
			return mainMenuTransition()
		}

		categoryIndex, err := strconv.Atoi(categoryInput)
		if err != nil || categoryIndex <= 0 || categoryIndex > len(categories) {
			m.terminal().Error("Invalid category choice")
			continue
		}

		selectedCategory := categories[categoryIndex-1]
		allOutfits, err := m.outfitService.ShowAllOutfits(selectedCategory.Name)
		if err != nil {
			m.terminal().Error(fmt.Sprintf("Error: %v", err))
			return mainMenuTransition()
		}
		if len(allOutfits) == 0 {
			m.terminal().Info(fmt.Sprintf("No outfits found in %s", selectedCategory.Name))
			continue
		}

		state, err := m.outfitService.GetOutfitState(selectedCategory)
		if err != nil {
			m.terminal().Error(fmt.Sprintf("Error: %v", err))
			return mainMenuTransition()
		}
		wornFileNames := currentCategoryWornFileNames(state)
		m.renderer.ShowManualSelectionOutfits(allOutfits, selectedCategory.Name, wornFileNames)

		outfitInput := strings.ToLower(strings.TrimSpace(m.terminal().Prompt(fmt.Sprintf("\nChoose an outfit (1-%d) or 'q' to go back: ", len(allOutfits)))))
		if outfitInput == "q" {
			continue
		}

		outfitIndex, err := strconv.Atoi(outfitInput)
		if err != nil || outfitIndex <= 0 || outfitIndex > len(allOutfits) {
			m.terminal().Error("Invalid outfit choice")
			continue
		}

		selectedOutfit := allOutfits[outfitIndex-1]
		result := m.presentation.PresentManualOutfit(selectedOutfit, selectedCategory.Name, wornFileNames[selectedOutfit.FileName])
		switch result {
		case OutfitChoiceSkipped:
			continue
		default:
			m.terminal().Println("Goodbye!")
			return exitMenuTransition()
		}
	}
}
