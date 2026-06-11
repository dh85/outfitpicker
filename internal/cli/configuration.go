package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
	"github.com/dh85/outfitpicker/internal/domain/validation"
)

type Configuration struct {
	OutfitPath         string
	Language           string
	ExcludedCategories []string
}

func PromptConfiguration(categoryService interfaces.CategoryService) *Configuration {
	return PromptConfigurationWithConsole(nil, categoryService)
}

func PromptConfigurationWithConsole(console Console, categoryService interfaces.CategoryService) *Configuration {
	terminal := consoleOrDefault(console)
	showFirstRunWelcome(console)

	path := promptWithConsole(console, "Where are your outfits stored? ")
	if strings.TrimSpace(path) == "" {
		terminal.Error("No directory path provided")
		return nil
	}

	if categoryService != nil {
		categoryInfos, err := categoryService.ScanCategories(strings.TrimSpace(path), nil)
		if err != nil {
			terminal.Error(fmt.Sprintf("Could not scan wardrobe directory: %v", err))
			return nil
		}
		showWardrobePreview(console, categoryInfos)
		if !ConfirmWithConsole(console, "Use this wardrobe? [Y/n]: ", true) {
			terminal.Info("Setup cancelled")
			return nil
		}
	}

	language := promptWithConsole(console, "Set language (en is default): ")

	return &Configuration{
		OutfitPath:         path,
		Language:           language,
		ExcludedCategories: promptExcludedCategoriesWithConsole(console, path, categoryService),
	}
}

func showFirstRunWelcome(console Console) {
	terminal := consoleOrDefault(console)
	terminal.Println("Welcome to OutfitPicker 👗")
	terminal.Println()
	terminal.Println("No wardrobe directory is configured yet.")
	terminal.Println()
}

func showWardrobePreview(console Console, categoryInfos []entities.CategoryInfo) {
	terminal := consoleOrDefault(console)
	terminal.Println()
	if len(categoryInfos) == 0 {
		terminal.Warning("No categories found in this wardrobe directory.")
		return
	}

	terminal.Printf("Found %d %s:\n", len(categoryInfos), pluralize("category", len(categoryInfos)))
	for _, info := range categoryInfos {
		terminal.Printf("  %s %-12s %s\n", setupCategoryPreviewIcon(info), sanitizeTerminalText(info.Category.Name), setupCategoryPreviewStatus(info))
	}
	terminal.Println()
}

func setupCategoryPreviewIcon(info entities.CategoryInfo) string {
	if info.State == entities.CategoryStateHasOutfits {
		return "✓"
	}
	return "⚠"
}

func setupCategoryPreviewStatus(info entities.CategoryInfo) string {
	switch info.State {
	case entities.CategoryStateHasOutfits:
		return fmt.Sprintf("%d %s", info.OutfitCount, pluralize("outfit", info.OutfitCount))
	case entities.CategoryStateEmpty:
		return "empty"
	case entities.CategoryStateNoAvatarFiles:
		return "no .avatar files found"
	case entities.CategoryStateUserExcluded:
		return "excluded"
	default:
		return string(info.State)
	}
}

func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	if word == "category" {
		return "categories"
	}
	return word + "s"
}

func (c Configuration) BuildConfig() (*entities.Config, error) {
	builder := entities.NewConfigBuilder().RootDirectory(strings.TrimSpace(c.OutfitPath))

	if language := normalizeLanguage(c.Language); language != "" {
		builder.Language(language)
	}

	if len(c.ExcludedCategories) > 0 {
		builder.Exclude(c.ExcludedCategories...)
	}

	return builder.Build()
}

func Confirm(message string, defaultValue bool) bool {
	return ConfirmWithConsole(nil, message, defaultValue)
}

func ConfirmWithConsole(console Console, message string, defaultValue bool) bool {
	response := promptWithConsole(console, message)
	if response == "" {
		return defaultValue
	}

	switch strings.ToLower(strings.TrimSpace(response)) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return defaultValue
	}
}

func Info(message string) {
	consoleOrDefault(nil).Info(message)
}

func Error(message string) {
	consoleOrDefault(nil).Error(message)
}

func prompt(message string) string {
	return promptFunc(message)
}

var promptFunc = func(message string) string {
	return readPrompt(os.Stdout, os.Stdin, message)
}

func parseExcludedCategories(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}

	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func promptExcludedCategories(path string, categoryService interfaces.CategoryService) []string {
	return promptExcludedCategoriesWithConsole(nil, path, categoryService)
}

func promptExcludedCategoriesWithConsole(console Console, path string, categoryService interfaces.CategoryService) []string {
	if categoryService == nil {
		excluded := promptWithConsole(console, "Exclude categories (separated by commas, or leave empty): ")
		return parseExcludedCategories(excluded)
	}

	categoryInfos, err := categoryService.ScanCategories(strings.TrimSpace(path), nil)
	if err != nil {
		consoleOrDefault(console).Error(fmt.Sprintf("Could not load categories for exclusion selection: %v", err))
		excluded := promptWithConsole(console, "Exclude categories (separated by commas, or leave empty): ")
		return parseExcludedCategories(excluded)
	}

	if len(categoryInfos) == 0 {
		consoleOrDefault(console).Info("No categories found to exclude")
		return nil
	}

	SectionWithConsole(console, "Exclude Categories", "🚫", uiYellow)
	consoleOrDefault(console).Println("Choose categories to exclude from random selection.")
	for index, info := range categoryInfos {
		consoleOrDefault(console).Printf("  %s 📁 %s%s\n", KeyLabel(fmt.Sprintf("%d", index+1)), sanitizeTerminalText(info.Category.Name), setupCategorySelectionSuffix(info))
	}

	input := promptWithConsole(console, "Exclude categories by number or name (comma-separated), or leave empty: ")
	return excludedCategoriesFromSelection(input, categoryInfos)
}

func excludedCategoriesFromSelection(input string, categoryInfos []entities.CategoryInfo) []string {
	options := make([]string, 0, len(categoryInfos))
	for _, info := range categoryInfos {
		options = append(options, info.Category.Name)
	}
	return parseCategorySelections(input, options)
}

func setupCategorySelectionSuffix(info entities.CategoryInfo) string {
	switch info.State {
	case entities.CategoryStateHasOutfits:
		return fmt.Sprintf(" (%d outfits)", info.OutfitCount)
	case entities.CategoryStateEmpty:
		return " (empty)"
	case entities.CategoryStateNoAvatarFiles:
		return " (no .avatar files)"
	default:
		return ""
	}
}

func normalizeLanguage(language string) string {
	normalized := strings.ToLower(strings.TrimSpace(language))
	if normalized == "" {
		return ""
	}
	if !validation.IsLanguageSupported(normalized) {
		return ""
	}
	return normalized
}
