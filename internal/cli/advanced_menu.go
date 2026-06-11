package cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

type AdvancedMenu struct {
	outfitService OutfitService
	console       Console
}

func (m AdvancedMenu) terminal() Console { return consoleOrDefault(m.console) }

func (m AdvancedMenu) Show() menuTransition {
	HeaderWithConsole(m.console, "Advanced Settings")
	SectionWithConsole(m.console, "Configuration Options", "⚙", uiCyan)
	for _, choice := range AllAdvancedChoices() {
		m.terminal().Printf("  %s %s\n", KeyLabel(strings.ToUpper(string(choice))), choice.Description())
	}

	input := m.terminal().Prompt("\nChoose a letter: ")
	choice, ok := ParseAdvancedChoice(input)
	if !ok {
		m.terminal().Error("Invalid choice")
		return advancedMenuTransition()
	}

	return m.dispatchChoice(choice)
}

func (m AdvancedMenu) dispatchChoice(choice AdvancedChoice) menuTransition {
	switch choice {
	case AdvancedChoiceChangePath:
		return m.handlePathChange()
	case AdvancedChoiceChangeLanguage:
		return m.handleLanguageChange()
	case AdvancedChoiceChangeExcluded:
		return m.handleExcludedChange()
	case AdvancedChoiceResetAll:
		return m.handleResetAll()
	case AdvancedChoiceResetCategory:
		return m.handleResetCategory()
	case AdvancedChoiceBack:
		return mainMenuTransition()
	case AdvancedChoiceQuit:
		m.terminal().Println("Goodbye!")
		return exitMenuTransition()
	case AdvancedChoiceResetSettings:
		return m.handleResetSettings()
	}

	return exitMenuTransition()
}

func (m AdvancedMenu) handleResetAll() menuTransition {
	m.terminal().Warning("This will reset worn status for all categories.")
	m.showResetAllPreview()
	confirm := m.terminal().Prompt("Reset all worn outfits? [y/N]: ")
	if !shouldProceedWithDestructiveAction(confirm) {
		m.terminal().Info("Reset cancelled")
		return advancedMenuTransition()
	}

	if err := m.outfitService.ResetAllCategories(); err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to reset: %v", err))
	} else {
		m.terminal().Success("All worn outfits reset")
	}
	return advancedMenuTransition()
}

func (m AdvancedMenu) handleResetCategory() menuTransition {
	categories, err := m.outfitService.GetCategories()
	if err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to load categories: %v", err))
		return advancedMenuTransition()
	}

	m.terminal().Println("\nSelect category to reset:")
	for index, category := range categories {
		m.terminal().Printf("  [%d] %s\n", index+1, sanitizeTerminalText(category.Name))
	}

	input := m.terminal().Prompt("\nChoose a number: ")
	if isBackOrQuitInput(input) {
		return advancedMenuTransition()
	}
	index, err := strconv.Atoi(normalizeChoiceInput(input))
	if err != nil || index <= 0 || index > len(categories) {
		m.terminal().Error("Invalid choice")
		return advancedMenuTransition()
	}

	category := categories[index-1]
	if err := m.outfitService.ResetCategory(category.Name); err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to reset category: %v", err))
	} else {
		m.terminal().Success(fmt.Sprintf("Reset worn outfits for %s", category.Name))
	}
	return advancedMenuTransition()
}

func (m AdvancedMenu) handlePathChange() menuTransition {
	currentConfig, err := m.outfitService.GetConfiguration()
	if err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to load configuration: %v", err))
		return advancedMenuTransition()
	}

	m.terminal().Printf("\nCurrent outfit path: %s\n", sanitizeTerminalText(currentConfig.Root))
	newPath := strings.TrimSpace(m.terminal().Prompt("Enter new outfit directory path: "))
	if newPath == "" {
		m.terminal().Error("No path provided")
		return advancedMenuTransition()
	}

	updatedConfig, err := buildUpdatedConfig(currentConfig, newPath, currentConfig.Language, cloneExcludedCategories(currentConfig.ExcludedCategories))
	if err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to update path: %v", err))
		return advancedMenuTransition()
	}

	if pathChangeNeedsResetConfirmation(currentConfig.Root, updatedConfig.Root) {
		m.terminal().Warning("Changing wardrobe path will reset worn outfit history.")
		m.terminal().Println()
		m.terminal().Printf("Current: %s\n", sanitizeTerminalText(displayWardrobePath(currentConfig.Root)))
		m.terminal().Printf("New:     %s\n", sanitizeTerminalText(displayWardrobePath(updatedConfig.Root)))
		m.terminal().Println()
		confirm := m.terminal().Prompt("Continue? [y/N]: ")
		if !shouldProceedWithDestructiveAction(confirm) {
			m.terminal().Info("Path change cancelled")
			return advancedMenuTransition()
		}
	}

	if err := m.outfitService.UpdateConfiguration(updatedConfig); err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to update path: %v", err))
	} else {
		m.terminal().Success(fmt.Sprintf("Outfit path updated to: %s", newPath))
	}
	return advancedMenuTransition()
}

func (m AdvancedMenu) handleLanguageChange() menuTransition {
	currentConfig, err := m.outfitService.GetConfiguration()
	if err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to load configuration: %v", err))
		return advancedMenuTransition()
	}

	m.terminal().Printf("\nCurrent language: %s\n", currentConfig.Language)
	m.terminal().Println("Available languages: en, es, fr, de, it, pt, ru, ja, ko, zh")
	newLanguage := strings.TrimSpace(m.terminal().Prompt("Enter new language code: "))
	if newLanguage == "" {
		m.terminal().Error("No language provided")
		return advancedMenuTransition()
	}

	normalized := normalizeLanguage(newLanguage)
	if normalized == "" {
		normalized = currentConfig.Language
	}

	updatedConfig, err := buildUpdatedConfig(currentConfig, currentConfig.Root, normalized, cloneExcludedCategories(currentConfig.ExcludedCategories))
	if err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to update language: %v", err))
		return advancedMenuTransition()
	}

	if err := m.outfitService.UpdateConfiguration(updatedConfig); err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to update language: %v", err))
	} else {
		m.terminal().Success(fmt.Sprintf("Language updated to: %s", normalized))
	}
	return advancedMenuTransition()
}

func (m AdvancedMenu) handleExcludedChange() menuTransition {
	for {
		currentConfig, err := m.outfitService.GetConfiguration()
		if err != nil {
			m.terminal().Error(fmt.Sprintf("Failed to load configuration: %v", err))
			return advancedMenuTransition()
		}
		allCategories, err := m.outfitService.GetCategories()
		if err != nil {
			m.terminal().Error(fmt.Sprintf("Failed to load categories: %v", err))
			return advancedMenuTransition()
		}

		excludedList := make([]string, 0, len(currentConfig.ExcludedCategories))
		for category := range currentConfig.ExcludedCategories {
			excludedList = append(excludedList, category)
		}
		sort.Strings(excludedList)

		nonExcludedList := make([]string, 0, len(allCategories))
		for _, category := range allCategories {
			if !currentConfig.ExcludedCategories[category.Name] {
				nonExcludedList = append(nonExcludedList, category.Name)
			}
		}
		sort.Strings(nonExcludedList)

		m.terminal().Println()
		SectionWithConsole(m.console, "Manage Excluded Categories", "🚫", uiYellow)
		m.terminal().Println("Excluded categories will not appear in random-across-categories selection.")

		if len(excludedList) > 0 {
			m.terminal().Println("\nCurrently Excluded:")
			for _, category := range excludedList {
				m.terminal().Printf("  • %s\n", sanitizeTerminalText(category))
			}
		} else {
			m.terminal().Println("\nNo categories are currently excluded")
		}

		if len(nonExcludedList) > 0 {
			m.terminal().Println("\nCurrently Available:")
			for _, category := range nonExcludedList {
				m.terminal().Printf("  • %s\n", sanitizeTerminalText(category))
			}
		}

		m.terminal().Println("\nOptions:")
		if len(nonExcludedList) > 0 {
			m.terminal().Println("  [A] Add category to exclusion list")
		}
		if len(excludedList) > 0 {
			m.terminal().Println("  [R] Remove category from exclusion list")
			m.terminal().Println("  [C] Clear all exclusions")
		}
		m.terminal().Println("  [B] Back to advanced menu")

		choice := normalizeChoiceInput(m.terminal().Prompt("\nChoose an option: "))
		switch choice {
		case "a", "add":
			m.applyExcludedAdd(currentConfig, nonExcludedList)
		case "r", "remove":
			m.applyExcludedRemove(currentConfig, excludedList)
		case "c", "clear":
			m.applyExcludedClear(currentConfig, excludedList)
		case "b", "back", "q", "quit", "exit":
			return advancedMenuTransition()
		default:
			m.terminal().Error("Invalid option")
		}
	}
}

func (m AdvancedMenu) handleExcludedAdd(currentConfig *entities.Config, nonExcludedList []string) {
	m.applyExcludedAdd(currentConfig, nonExcludedList)
	m.handleExcludedChange()
}

func (m AdvancedMenu) applyExcludedAdd(currentConfig *entities.Config, nonExcludedList []string) {
	if len(nonExcludedList) == 0 {
		m.terminal().Error("All categories are already excluded")
		return
	}

	m.terminal().Println("\nAdd Categories to Exclusion List")
	for index, category := range nonExcludedList {
		m.terminal().Printf("  [%d] %s\n", index+1, sanitizeTerminalText(category))
	}

	input := strings.TrimSpace(m.terminal().Prompt("\nEnter numbers (comma-separated) or category names: "))
	if input == "" {
		return
	}

	categoriesToAdd := parseCategorySelections(input, nonExcludedList)
	if len(categoriesToAdd) == 0 {
		m.terminal().Error("No valid categories selected")
		return
	}

	newExcluded := cloneExcludedCategories(currentConfig.ExcludedCategories)
	for _, category := range categoriesToAdd {
		newExcluded[category] = true
	}

	updatedConfig, err := buildUpdatedConfig(currentConfig, currentConfig.Root, currentConfig.Language, newExcluded)
	if err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to update excluded categories: %v", err))
		return
	}

	if err := m.outfitService.UpdateConfiguration(updatedConfig); err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to update excluded categories: %v", err))
	} else {
		m.terminal().Success(fmt.Sprintf("Added to exclusion list: %s", strings.Join(categoriesToAdd, ", ")))
	}
}

func (m AdvancedMenu) handleExcludedRemove(currentConfig *entities.Config, excludedList []string) {
	m.applyExcludedRemove(currentConfig, excludedList)
	m.handleExcludedChange()
}

func (m AdvancedMenu) applyExcludedRemove(currentConfig *entities.Config, excludedList []string) {
	if len(excludedList) == 0 {
		m.terminal().Error("No categories are excluded")
		return
	}

	m.terminal().Println("\nRemove Categories from Exclusion List")
	for index, category := range excludedList {
		m.terminal().Printf("  [%d] %s\n", index+1, sanitizeTerminalText(category))
	}

	input := strings.TrimSpace(m.terminal().Prompt("\nEnter numbers (comma-separated) or category names: "))
	if input == "" {
		return
	}

	categoriesToRemove := parseCategorySelections(input, excludedList)
	if len(categoriesToRemove) == 0 {
		m.terminal().Error("No valid categories selected")
		return
	}

	newExcluded := cloneExcludedCategories(currentConfig.ExcludedCategories)
	for _, category := range categoriesToRemove {
		delete(newExcluded, category)
	}

	updatedConfig, err := buildUpdatedConfig(currentConfig, currentConfig.Root, currentConfig.Language, newExcluded)
	if err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to update excluded categories: %v", err))
		return
	}

	if err := m.outfitService.UpdateConfiguration(updatedConfig); err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to update excluded categories: %v", err))
	} else {
		m.terminal().Success(fmt.Sprintf("Removed from exclusion list: %s", strings.Join(categoriesToRemove, ", ")))
	}
}

func (m AdvancedMenu) handleExcludedClear(currentConfig *entities.Config, excludedList []string) {
	m.applyExcludedClear(currentConfig, excludedList)
	m.handleExcludedChange()
}

func (m AdvancedMenu) applyExcludedClear(currentConfig *entities.Config, excludedList []string) {
	if len(excludedList) == 0 {
		m.terminal().Error("No categories are excluded")
		return
	}

	updatedConfig, err := buildUpdatedConfig(currentConfig, currentConfig.Root, currentConfig.Language, map[string]bool{})
	if err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to clear exclusions: %v", err))
		return
	}

	if err := m.outfitService.UpdateConfiguration(updatedConfig); err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to clear exclusions: %v", err))
	} else {
		m.terminal().Success("All exclusions cleared")
	}
}

func (m AdvancedMenu) handleResetSettings() menuTransition {
	m.terminal().Println("WARNING: This will delete all configuration and worn outfit data.")
	confirm := m.terminal().Prompt("Reset all settings and worn outfit data? [y/N]: ")
	if !isYesInput(confirm) {
		m.terminal().Info("Reset cancelled")
		return advancedMenuTransition()
	}

	if err := m.outfitService.FactoryReset(); err != nil {
		m.terminal().Error(fmt.Sprintf("Failed to reset settings: %v", err))
		return advancedMenuTransition()
	}

	m.terminal().Success("All settings and data reset successfully")
	m.terminal().Println("Please restart the application to reconfigure")
	return exitMenuTransition()
}

func parseCategorySelections(input string, options []string) []string {
	parts := strings.Split(input, ",")
	selected := map[string]bool{}
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if index, err := strconv.Atoi(trimmed); err == nil {
			if index > 0 && index <= len(options) {
				selected[options[index-1]] = true
			}
			continue
		}
		for _, option := range options {
			if option == trimmed {
				selected[option] = true
				break
			}
		}
	}

	result := make([]string, 0, len(selected))
	for category := range selected {
		result = append(result, category)
	}
	sort.Strings(result)
	return result
}

func (m AdvancedMenu) showResetAllPreview() {
	states, err := m.outfitService.GetAllOutfitStates()
	if err != nil {
		m.terminal().Warning(fmt.Sprintf("Could not load reset preview: %v", err))
		return
	}

	affected := make([]entities.CategoryOutfitState, 0, len(states))
	for _, state := range states {
		if state.WornCount() > 0 {
			affected = append(affected, state)
		}
	}
	sort.Slice(affected, func(i, j int) bool {
		return affected[i].Category.Name < affected[j].Category.Name
	})

	m.terminal().Println()
	m.terminal().Println("Affected:")
	if len(affected) == 0 {
		m.terminal().Println("  none")
		m.terminal().Println()
		return
	}
	for _, state := range affected {
		m.terminal().Printf("  %-10s %d worn\n", sanitizeTerminalText(state.Category.Name), state.WornCount())
	}
	m.terminal().Println()
}

func pathChangeNeedsResetConfirmation(currentRoot, newRoot string) bool {
	return strings.TrimSpace(currentRoot) != "" && strings.TrimSpace(currentRoot) != strings.TrimSpace(newRoot)
}

func shouldProceedWithDestructiveAction(input string) bool {
	return isYesInput(input)
}
