package app

import (
	"fmt"
	"io"

	"github.com/dh85/outfitpicker/internal/storage"
	"github.com/dh85/outfitpicker/internal/ui"
)

// I18n-aware helper functions
func handleDirectCategoryWithI18n(categoryOpt string, categories []string, cache *storage.Manager, pr *prompter, stdout io.Writer, i18n *I18n) error {
	chosen := findCategory(categoryOpt, categories)
	if chosen == "" {
		avail := baseNames(categories)
		errorMsg := fmt.Sprintf("category %q not found; available: %s", categoryOpt, avail)
		if i18n != nil {
			errorMsg = i18n.T("category_not_found", categoryOpt)
		}
		return fmt.Errorf("%s", errorMsg)
	}
	return runCategoryFlow(chosen, cache, pr, stdout)
}

func handleUncategorizedOnlyMenuWithI18n(uncategorized []string, cache *storage.Manager, pr *prompter, stdout io.Writer, i18n *I18n) error {
	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: false}
	var uiInstance *ui.UI
	if i18n != nil {
		uiInstance = ui.NewUIWithI18n(stdout, theme, i18n)
	} else {
		uiInstance = ui.NewUI(stdout, theme)
	}
	uiInstance.UncategorizedOnlyMenu(len(uncategorized))

	choice, _ := pr.readLineLower()
	switch choice {
	case "r":
		return handleUncategorizedRandom(uncategorized, cache, pr, stdout)
	case "s":
		return showUncategorizedSelected(uncategorized, cache, stdout)
	case "u":
		return showUncategorizedUnselected(uncategorized, cache, stdout)
	case "m":
		return handleManualSelection(nil, uncategorized, cache, pr, stdout)
	case "q":
		exitMsg := "Exiting."
		if i18n != nil {
			exitMsg = i18n.T("exiting")
		}
		fmt.Fprintln(stdout, exitMsg)
		return nil
	default:
		return fmt.Errorf("invalid selection")
	}
}

func handleNumericSelectionWithI18n(n int, categories, uncategorized []string, cache *storage.Manager, pr *prompter, stdout io.Writer, i18n *I18n) error {
	totalOptions := len(categories)
	if len(uncategorized) > 0 {
		totalOptions++ // Add uncategorized option
	}

	if n < 1 || n > totalOptions {
		return fmt.Errorf("invalid selection")
	}

	if n <= len(categories) {
		return runCategoryFlow(categories[n-1], cache, pr, stdout)
	}

	// Handle uncategorized selection
	return handleUncategorizedFlow(uncategorized, cache, pr, stdout)
}

func handleMenuOptionWithI18n(choice string, categories, uncategorized []string, cache *storage.Manager, pr *prompter, stdout io.Writer, i18n *I18n) error {
	switch choice {
	case "r":
		return randomAcrossAll(categories, uncategorized, cache, pr, stdout)
	case "s":
		return showSelectedAcrossAll(categories, uncategorized, cache, stdout)
	case "u":
		return showUnselectedAcrossAll(categories, uncategorized, cache, stdout)
	case "m":
		return handleManualSelection(categories, uncategorized, cache, pr, stdout)
	case "q":
		exitMsg := "Exiting."
		if i18n != nil {
			exitMsg = i18n.T("exiting")
		}
		fmt.Fprintln(stdout, exitMsg)
		return nil
	default:
		return fmt.Errorf("invalid selection")
	}
}
