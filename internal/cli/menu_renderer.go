package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

type categoryStateReader interface {
	GetOutfitState(category entities.CategoryReference) (entities.CategoryOutfitState, error)
}

type MenuRenderer struct {
	console Console
}

func NewMenuRenderer(consoles ...Console) MenuRenderer {
	return MenuRenderer{console: optionalConsole(consoles)}
}

func (r MenuRenderer) terminal() Console {
	return consoleOrDefault(r.console)
}

func (r MenuRenderer) ShowTitle() {
	title := "👗 Outfit Picker"
	padding := 0
	if width := 40 - len(title); width > 0 {
		padding = width / 2
	}
	r.terminal().Println()
	r.terminal().Println(Colorize(strings.Repeat(" ", padding)+title, uiBold+uiCyan))
	r.terminal().Println(Colorize(repeatLine("─", 40), uiCyan))
}

func (r MenuRenderer) ShowOutfitDirectory(path string) {
	r.terminal().Printf("📁 %s\n\n", Colorize(sanitizeTerminalText(filepath.Clean(path)), uiCyan))
}

func (r MenuRenderer) ShowAvailableCategories(availableCategories []entities.CategoryInfo, wardrobe categoryStateReader) {
	SectionWithConsole(r.console, "Available Categories", "📂", uiBlue)
	for index, info := range availableCategories {
		state, err := wardrobe.GetOutfitState(info.Category)
		statusText := fmt.Sprintf("%d outfits", info.OutfitCount)
		if err == nil {
			statusText = fmt.Sprintf("%d of %d outfits worn", state.WornCount(), state.TotalCount())
		}
		safeName := sanitizeTerminalText(info.Category.Name)
		padding := strings.Repeat(" ", max(0, 20-len(safeName)))
		r.terminal().Printf("  %s 📁 %s%s %s\n", KeyLabel(fmt.Sprintf("%d", index+1)), safeName, padding, Dim(statusText))
	}
}

func (r MenuRenderer) ShowUnavailableCategories(categoryInfos []entities.CategoryInfo, outfitService OutfitService) {
	var excluded []string
	var noOutfits []string

	for _, info := range categoryInfos {
		switch info.State {
		case entities.CategoryStateUserExcluded:
			count, err := outfitService.GetActualOutfitCount(info.Category)
			if err == nil {
				excluded = append(excluded, fmt.Sprintf("%s (%d outfits)", sanitizeTerminalText(info.Category.Name), count))
			} else {
				excluded = append(excluded, sanitizeTerminalText(info.Category.Name))
			}
		case entities.CategoryStateEmpty, entities.CategoryStateNoAvatarFiles:
			noOutfits = append(noOutfits, fmt.Sprintf("%s (Add .avatar files to %s)", sanitizeTerminalText(info.Category.Name), sanitizeTerminalText(info.Category.Path)))
		}
	}

	if len(excluded) == 0 && len(noOutfits) == 0 {
		return
	}

	r.terminal().Println()
	SectionWithConsole(r.console, "Unavailable Categories", "⚠", uiYellow)
	if len(excluded) > 0 {
		sort.Strings(excluded)
		r.terminal().Printf("  🚫 %s %s\n", Dim("Excluded:"), strings.Join(excluded, ", "))
	}
	if len(noOutfits) > 0 {
		sort.Strings(noOutfits)
		r.terminal().Printf("  📄 %s\n", Dim("No outfits found:"))
		for _, line := range noOutfits {
			r.terminal().Printf("    • %s\n", line)
		}
	}
}

func (r MenuRenderer) ShowMenuOptions() {
	SectionWithConsole(r.console, "Actions", "📋", uiCyan)
	for _, choice := range AllMenuChoices() {
		r.terminal().Printf("  %s %s\n", KeyLabel(strings.ToUpper(string(choice))), choice.Description())
	}
}

func (r MenuRenderer) ShowWornOutfits(wornByCategory map[string][]entities.OutfitReference) {
	r.terminal().Println()
	SectionWithConsole(r.console, "Worn Outfits", "✅", uiGreen)
	for _, categoryName := range sortedCategoryNames(wornByCategory) {
		outfits := wornByCategory[categoryName]
		r.terminal().Printf("\n📁 %s %s\n", Colorize(sanitizeTerminalText(categoryName), uiBold+uiBlue), Dim(fmt.Sprintf("(%d worn)", len(outfits))))
		for _, outfit := range outfits {
			r.terminal().Printf("  • %s\n", displayOutfitName(outfit.FileName))
		}
	}
	r.terminal().Println()
}

func (r MenuRenderer) ShowUnwornOutfits(unwornByCategory map[string][]entities.OutfitReference) {
	r.terminal().Println()
	SectionWithConsole(r.console, "Unworn Outfits", "📄", uiBlue)
	for _, categoryName := range sortedCategoryNames(unwornByCategory) {
		outfits := unwornByCategory[categoryName]
		r.terminal().Printf("\n📁 %s %s\n", Colorize(sanitizeTerminalText(categoryName), uiBold+uiBlue), Dim(fmt.Sprintf("(%d unworn)", len(outfits))))
		for _, outfit := range outfits {
			r.terminal().Printf("  • %s\n", displayOutfitName(outfit.FileName))
		}
	}
	r.terminal().Println()
}

func (r MenuRenderer) ShowManualSelectionCategories(categories []entities.CategoryReference) {
	r.terminal().Println()
	SectionWithConsole(r.console, "Choose Your Outfit", "👕", uiCyan)
	for index, category := range categories {
		r.terminal().Printf("  %s 📁 %s\n", KeyLabel(fmt.Sprintf("%d", index+1)), sanitizeTerminalText(category.Name))
	}
}

func (r MenuRenderer) ShowManualSelectionOutfits(allOutfits []entities.OutfitReference, categoryName string, wornFileNames map[string]bool) {
	r.terminal().Println()
	SectionWithConsole(r.console, fmt.Sprintf("Outfits in %s", sanitizeTerminalText(categoryName)), "👗", uiBlue)
	for index, outfit := range allOutfits {
		wornStatus := ""
		if wornFileNames[outfit.FileName] {
			wornStatus = " " + Dim("(worn)")
		}
		r.terminal().Printf("  %s %s%s\n", KeyLabel(fmt.Sprintf("%d", index+1)), displayOutfitName(outfit.FileName), wornStatus)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
