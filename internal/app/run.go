package app

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dh85/outfitpicker/internal/storage"
	"github.com/dh85/outfitpicker/internal/ui"
)

func Run(rootPath, categoryOpt string, stdin io.Reader, stdout io.Writer) error {
	cache, categories, err := initializeApp(rootPath)
	if err != nil {
		return err
	}

	pr := &prompter{r: bufio.NewReader(stdin), w: stdout}

	if categoryOpt != "" {
		return handleDirectCategory(categoryOpt, categories, cache, pr, stdout)
	}

	return handleMainMenu(categories, cache, pr, stdout)
}

func initializeApp(rootPath string) (*storage.Manager, []string, error) {
	cache, err := storage.NewManager(rootPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init cache: %w", err)
	}

	rootAbs, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid root path: %s", rootPath)
	}

	categories, err := listCategories(rootAbs)
	if err != nil {
		return nil, nil, err
	}
	if len(categories) == 0 {
		return nil, nil, fmt.Errorf("no category folders found in %q", filepath.Base(rootAbs))
	}

	return cache, categories, nil
}

func handleDirectCategory(categoryOpt string, categories []string, cache *storage.Manager, pr *prompter, stdout io.Writer) error {
	chosen := findCategory(categoryOpt, categories)
	if chosen == "" {
		avail := baseNames(categories)
		return fmt.Errorf("category %q not found; available: %s", categoryOpt, strings.Join(avail, ", "))
	}
	return runCategoryFlow(chosen, cache, pr, stdout)
}

func findCategory(categoryOpt string, categories []string) string {
	for _, c := range categories {
		if strings.EqualFold(filepath.Base(c), categoryOpt) {
			return c
		}
	}
	return ""
}

func handleMainMenu(categories []string, cache *storage.Manager, pr *prompter, stdout io.Writer) error {
	displayMainMenu(categories, stdout)
	choice, _ := pr.readLineLower()

	if n, err := parseNumericChoice(choice); err == nil {
		return handleNumericSelection(n, categories, cache, pr, stdout)
	}

	return handleMenuOption(choice, categories, cache, pr, stdout)
}

func displayMainMenu(categories []string, stdout io.Writer) {
	// Create enhanced UI for main menu
	theme := ui.Theme{
		UseColors: shouldUseColors(),
		UseEmojis: true,
		Compact:   false,
	}
	uiInstance := ui.NewUI(stdout, theme)
	uiInstance.MainMenu(categories)
}

func parseNumericChoice(choice string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(choice))
}

func handleNumericSelection(n int, categories []string, cache *storage.Manager, pr *prompter, stdout io.Writer) error {
	if n < 1 || n > len(categories) {
		return fmt.Errorf("invalid category selection")
	}
	return runCategoryFlow(categories[n-1], cache, pr, stdout)
}

func handleMenuOption(choice string, categories []string, cache *storage.Manager, pr *prompter, stdout io.Writer) error {
	switch choice {
	case "r":
		return randomAcrossAll(categories, cache, pr, stdout)
	case "s":
		return showSelectedAcrossAll(categories, cache, stdout)
	case "u":
		return showUnselectedAcrossAll(categories, cache, stdout)
	case "q":
		fmt.Fprintln(stdout, "Exiting.")
		return nil
	default:
		return fmt.Errorf("invalid selection")
	}
}
