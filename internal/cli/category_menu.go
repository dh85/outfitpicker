package cli

import (
	"fmt"
	"strings"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

type CategoryMenu struct {
	outfitService OutfitService
	selector      RandomOutfitSelector
	presentation  OutfitPresentation
	renderer      MenuRenderer
	category      entities.CategoryReference
	console       Console
}

func (m CategoryMenu) terminal() Console { return consoleOrDefault(m.console) }

type categoryMenuAction int

const (
	categoryMenuActionInvalid categoryMenuAction = iota
	categoryMenuActionPick
	categoryMenuActionBack
	categoryMenuActionResetAndPick
)

type categoryMenuView struct {
	statusText    string
	message       string
	options       []string
	defaultAction categoryMenuAction
	exhausted     bool
}

func (m CategoryMenu) Show() menuTransition {
	state, err := m.outfitService.GetOutfitState(m.category)
	if err != nil {
		m.terminal().Error(fmt.Sprintf("Error: %v", err))
		return exitMenuTransition()
	}

	view := buildCategoryMenuView(m.category.Name, state)
	HeaderWithConsole(m.console, "Category")
	m.terminal().Printf("📁 %s %s\n", Colorize(sanitizeTerminalText(m.category.Name), uiBold+uiBlue), Dim("("+view.statusText+")"))
	if view.message != "" {
		m.terminal().Warning(view.message)
	}
	for _, option := range view.options {
		m.terminal().Println(option)
	}

	input := strings.ToLower(strings.TrimSpace(m.terminal().Prompt("Choose an option: ")))
	switch resolveCategoryMenuAction(input, view.exhausted) {
	case categoryMenuActionResetAndPick:
		if err := m.outfitService.ResetCategory(m.category.Name); err != nil {
			m.terminal().Error(fmt.Sprintf("Error: %v", err))
			return exitMenuTransition()
		}
		m.terminal().Success(fmt.Sprintf("Reset worn outfits for %s", m.category.Name))
		return m.handleOutfitLoop()
	case categoryMenuActionBack:
		return mainMenuTransition()
	case categoryMenuActionPick:
		return m.handleOutfitLoop()
	default:
		m.terminal().Error("Invalid choice")
		return categoryMenuTransition(m.category)
	}
}

func (m CategoryMenu) handleOutfitLoop() menuTransition {
	for {
		outfit, err := m.selector.ShowNextUniqueRandomOutfitFrom(m.category.Name)
		if err != nil {
			m.terminal().Error(fmt.Sprintf("Error: %v", err))
			return mainMenuTransition()
		}
		if outfit == nil {
			m.terminal().Info(fmt.Sprintf("No outfits available in %s", m.category.Name))
			return mainMenuTransition()
		}

		result := m.presentation.PresentOutfitWithChoice(*outfit)
		switch result {
		case OutfitChoiceWorn:
			m.terminal().Println("Goodbye!")
			return exitMenuTransition()
		case OutfitChoiceSkipped:
			continue
		case OutfitChoiceBack:
			return mainMenuTransition()
		case OutfitChoiceQuit:
			m.terminal().Println("Goodbye!")
			return exitMenuTransition()
		}
	}
}

func buildCategoryMenuView(categoryName string, state entities.CategoryOutfitState) categoryMenuView {
	view := categoryMenuView{
		statusText: fmt.Sprintf("%d of %d outfits worn", state.WornCount(), state.TotalCount()),
	}

	if state.IsRotationComplete() {
		view.message = fmt.Sprintf("All outfits in %s have been worn.", categoryName)
		view.options = []string{
			"  [R] Reset category and pick a random outfit",
			"  [B] Back (default)",
		}
		view.defaultAction = categoryMenuActionBack
		view.exhausted = true
		return view
	}

	view.options = []string{
		"  [P] Pick random outfit (default)",
		"  [B] Back",
	}
	view.defaultAction = categoryMenuActionPick
	return view
}

func resolveCategoryMenuAction(input string, exhausted bool) categoryMenuAction {
	if exhausted {
		switch input {
		case "r":
			return categoryMenuActionResetAndPick
		case "", "b":
			return categoryMenuActionBack
		default:
			return categoryMenuActionInvalid
		}
	}

	switch input {
	case "", "p":
		return categoryMenuActionPick
	case "b":
		return categoryMenuActionBack
	default:
		return categoryMenuActionInvalid
	}
}
