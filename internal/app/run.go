package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/dh85/outfitpicker/internal/storage"
	"github.com/dh85/outfitpicker/internal/ui"
)

// FileGroup represents a group of files (category or uncategorized)
type FileGroup struct {
	Name  string
	Files []GroupedFile
}

// GroupedFile represents a file with selection status
type GroupedFile struct {
	Name     string
	Path     string
	Selected bool
}

func Run(rootPath, categoryOpt string, stdin io.Reader, stdout io.Writer) error {
	return RunWithI18n(rootPath, categoryOpt, stdin, stdout, nil)
}

func RunWithI18n(rootPath, categoryOpt string, stdin io.Reader, stdout io.Writer, i18n *I18n) error {
	cache, categories, uncategorized, err := initializeApp(rootPath)
	if err != nil {
		return err
	}

	pr := &prompter{r: bufio.NewReader(stdin), w: stdout}

	if categoryOpt != "" {
		return handleDirectCategoryWithI18n(categoryOpt, categories, cache, pr, stdout, i18n)
	}

	return handleMainMenuWithI18n(categories, uncategorized, cache, pr, stdout, i18n)
}

func initializeApp(rootPath string) (*storage.Manager, []string, []string, error) {
	cache, err := storage.NewManager(rootPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to init cache: %w", err)
	}

	rootAbs, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid root path: %s", rootPath)
	}

	categories, err := listCategories(rootAbs)
	if err != nil {
		return nil, nil, nil, err
	}

	uncategorized, err := listUncategorizedFiles(rootAbs)
	if err != nil {
		return nil, nil, nil, err
	}

	// Handle different scenarios
	if len(categories) == 0 && len(uncategorized) == 0 {
		return nil, nil, nil, fmt.Errorf("no outfit files found in %q", filepath.Base(rootAbs))
	}

	return cache, categories, uncategorized, nil
}

func findCategory(categoryOpt string, categories []string) string {
	for _, c := range categories {
		if strings.EqualFold(filepath.Base(c), categoryOpt) {
			return c
		}
	}
	return ""
}

func handleMainMenuWithI18n(categories, uncategorized []string, cache *storage.Manager, pr *prompter, stdout io.Writer, i18n *I18n) error {
	// Scenario 2: No categories, only uncategorized files
	if len(categories) == 0 && len(uncategorized) > 0 {
		return handleUncategorizedOnlyMenuWithI18n(uncategorized, cache, pr, stdout, i18n)
	}

	// Scenario 4: Categories exist but are all empty, show uncategorized
	if len(categories) > 0 && areAllCategoriesEmpty(categories) {
		if len(uncategorized) == 0 {
			errorMsg := "all categories are empty and no uncategorized files found"
			if i18n != nil {
				errorMsg = i18n.T("no_outfits_available")
			}
			return fmt.Errorf("%s", errorMsg)
		}
		theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: false}
		var uiInstance *ui.UI
		if i18n != nil {
			uiInstance = ui.NewUIWithI18n(stdout, theme, i18n)
		} else {
			uiInstance = ui.NewUI(stdout, theme)
		}
		uiInstance.Warning("All categories are empty")
		return handleUncategorizedOnlyMenuWithI18n(uncategorized, cache, pr, stdout, i18n)
	}

	// Standard menu with categories and optional uncategorized
	displayMainMenuWithI18n(categories, uncategorized, stdout, i18n)
	choice, _ := pr.readLineLower()

	if n, err := parseNumericChoice(choice); err == nil {
		return handleNumericSelectionWithI18n(n, categories, uncategorized, cache, pr, stdout, i18n)
	}

	return handleMenuOptionWithI18n(choice, categories, uncategorized, cache, pr, stdout, i18n)
}

func displayMainMenuWithI18n(categories, uncategorized []string, stdout io.Writer, i18n *I18n) {
	// Create enhanced UI for main menu
	theme := ui.Theme{
		UseColors: shouldUseColors(),
		UseEmojis: true,
		Compact:   false,
	}

	var uiInstance *ui.UI
	if i18n != nil {
		uiInstance = ui.NewUIWithI18n(stdout, theme, i18n)
	} else {
		uiInstance = ui.NewUI(stdout, theme)
	}
	uiInstance.MainMenu(categories, uncategorized)
}

func listUncategorizedFiles(rootAbs string) ([]string, error) {
	ents, err := os.ReadDir(rootAbs)
	if err != nil {
		return nil, fmt.Errorf("failed to read root %q: %w", rootAbs, err)
	}

	var files []string
	for _, e := range ents {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") &&
			!strings.EqualFold(e.Name(), "Downloads") &&
			!strings.Contains(e.Name(), "Cache") {
			files = append(files, filepath.Join(rootAbs, e.Name()))
		}
	}

	return files, nil
}

func areAllCategoriesEmpty(categories []string) bool {
	for _, cat := range categories {
		count, err := categoryFileCount(cat)
		if err == nil && count > 0 {
			return false
		}
	}
	return true
}

func parseNumericChoice(choice string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(choice))
}

func handleUncategorizedFlow(uncategorized []string, cache *storage.Manager, pr *prompter, stdout io.Writer) error {
	const uncategorizedKey = "UNCATEGORIZED"

	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: false}
	uiInstance := ui.NewUI(stdout, theme)

	selected := cache.Load()[uncategorizedKey]
	uiInstance.UncategorizedInfo(len(uncategorized), len(selected))
	uiInstance.Menu()

	choice, _ := pr.readLineLower()
	switch choice {
	case "r":
		return handleUncategorizedRandom(uncategorized, cache, pr, stdout)
	case "s":
		return showUncategorizedSelected(uncategorized, cache, stdout)
	case "u":
		return showUncategorizedUnselected(uncategorized, cache, stdout)
	case "q":
		fmt.Fprintln(stdout, "Exiting.")
		return nil
	default:
		return fmt.Errorf("invalid selection")
	}
}

func handleUncategorizedRandom(uncategorized []string, cache *storage.Manager, pr *prompter, stdout io.Writer) error {
	const uncategorizedKey = "UNCATEGORIZED"

	selected := toSet(cache.Load()[uncategorizedKey])
	var available []string

	for _, f := range uncategorized {
		if !selected[filepath.Base(f)] {
			available = append(available, f)
		}
	}

	for {
		if len(available) == 0 {
			fmt.Fprintf(stdout, "\n🎉 all uncategorized files have been selected\n")
			cache.Clear(uncategorizedKey)
			fmt.Fprintf(stdout, "cache cleared for uncategorized files — next random will restart the cycle\n")
			return nil
		}

		idx := rand.Intn(len(available))
		randomFile := available[idx]
		fileName := filepath.Base(randomFile)

		theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: false}
		uiInstance := ui.NewUI(stdout, theme)
		uiInstance.RandomSelection(fileName)

		action, err := pr.readLineLowerDefault("k")
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Fprintln(stdout, "invalid action. please try again.")
			continue
		}

		switch action {
		case "k":
			theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: true}
			uiInstance := ui.NewUI(stdout, theme)
			uiInstance.KeepAction(fileName)
			cache.Add(fileName, uncategorizedKey)
			return nil
		case "s":
			theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: true}
			uiInstance := ui.NewUI(stdout, theme)
			uiInstance.SkipAction(fileName)
			available = append(available[:idx], available[idx+1:]...)
		case "d":
			return handleDeleteFile(randomFile, pr, stdout)
		case "q":
			fmt.Fprintln(stdout, "Exiting.")
			return nil
		default:
			fmt.Fprintln(stdout, "invalid action. please try again.")
		}
	}
}

func showUncategorizedSelected(uncategorized []string, cache *storage.Manager, stdout io.Writer) error {
	const uncategorizedKey = "UNCATEGORIZED"

	selected := cache.Load()[uncategorizedKey]
	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: false}
	uiInstance := ui.NewUI(stdout, theme)
	uiInstance.SelectedFiles("Uncategorized", selected)
	return nil
}

func showUncategorizedUnselected(uncategorized []string, cache *storage.Manager, stdout io.Writer) error {
	const uncategorizedKey = "UNCATEGORIZED"

	selected := toSet(cache.Load()[uncategorizedKey])
	var unselected []string

	for _, f := range uncategorized {
		if !selected[filepath.Base(f)] {
			unselected = append(unselected, filepath.Base(f))
		}
	}

	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: false}
	uiInstance := ui.NewUI(stdout, theme)
	uiInstance.UnselectedFiles(unselected)
	return nil
}

func handleManualSelection(categories, uncategorized []string, cache *storage.Manager, pr *prompter, stdout io.Writer) error {
	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: false}
	uiInstance := ui.NewUI(stdout, theme)

	// Build file list with indexing
	allFiles, groupInfo := buildIndexedFileList(categories, uncategorized, cache)

	totalFiles := len(allFiles)
	groupCount := len(categories)
	if len(uncategorized) > 0 {
		groupCount++
	}

	uiInstance.ManualSelectionMenu(groupCount, totalFiles)

	// Display all groups
	fileIndex := 1
	m := cache.Load()

	for _, cat := range categories {
		ents, err := os.ReadDir(cat)
		if err != nil {
			continue
		}

		var files []string
		selected := toSet(m[cat])

		for _, e := range ents {
			if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				files = append(files, e.Name())
			}
		}

		if len(files) > 0 {
			sort.Strings(files)
			fileIndex = uiInstance.DisplayFileGroup(filepath.Base(cat), files, selected, fileIndex)
		}
	}

	// Display uncategorized files
	if len(uncategorized) > 0 {
		const uncategorizedKey = "UNCATEGORIZED"
		var files []string
		selected := toSet(m[uncategorizedKey])

		for _, f := range uncategorized {
			files = append(files, filepath.Base(f))
		}

		sort.Strings(files)
		uiInstance.DisplayFileGroup("Uncategorized", files, selected, fileIndex)
	}

	choice, _ := pr.readLineLower()
	if n, err := parseNumericChoice(choice); err == nil {
		return handleManualFileSelection(n, allFiles, groupInfo, cache, pr, stdout)
	}

	if choice == "q" {
		fmt.Fprintln(stdout, "Exiting.")
		return nil
	}

	return fmt.Errorf("invalid selection")
}

func buildIndexedFileList(categories, uncategorized []string, cache *storage.Manager) ([]GroupedFile, map[int]string) {
	var allFiles []GroupedFile
	groupInfo := make(map[int]string) // maps file index to cache key
	m := cache.Load()

	// Add categorized files
	for _, cat := range categories {
		ents, err := os.ReadDir(cat)
		if err != nil {
			continue
		}

		selected := toSet(m[cat])
		var files []string

		for _, e := range ents {
			if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				files = append(files, e.Name())
			}
		}

		sort.Strings(files)
		for _, name := range files {
			allFiles = append(allFiles, GroupedFile{
				Name:     name,
				Path:     filepath.Join(cat, name),
				Selected: selected[name],
			})
			groupInfo[len(allFiles)] = cat // 1-based indexing
		}
	}

	// Add uncategorized files
	if len(uncategorized) > 0 {
		const uncategorizedKey = "UNCATEGORIZED"
		selected := toSet(m[uncategorizedKey])
		var files []string

		for _, f := range uncategorized {
			files = append(files, filepath.Base(f))
		}

		sort.Strings(files)
		for _, name := range files {
			// Find full path
			var fullPath string
			for _, f := range uncategorized {
				if filepath.Base(f) == name {
					fullPath = f
					break
				}
			}

			allFiles = append(allFiles, GroupedFile{
				Name:     name,
				Path:     fullPath,
				Selected: selected[name],
			})
			groupInfo[len(allFiles)] = uncategorizedKey // 1-based indexing
		}
	}

	return allFiles, groupInfo
}

func handleManualFileSelection(n int, allFiles []GroupedFile, groupInfo map[int]string, cache *storage.Manager, pr *prompter, stdout io.Writer) error {
	if n < 1 || n > len(allFiles) {
		return fmt.Errorf("invalid file selection")
	}

	selectedFile := allFiles[n-1]
	cacheKey := groupInfo[n]

	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: false}
	uiInstance := ui.NewUI(stdout, theme)

	if selectedFile.Selected {
		uiInstance.Warning(fmt.Sprintf("You've already picked '%s' before!", selectedFile.Name))
		return nil
	}

	groupName := "Uncategorized"
	if cacheKey != "UNCATEGORIZED" {
		groupName = filepath.Base(cacheKey)
	}

	uiInstance.Success(fmt.Sprintf("Great choice! I've saved '%s' from %s", selectedFile.Name, groupName))
	cache.Add(selectedFile.Name, cacheKey)
	return nil
}

func handleDeleteFile(filePath string, pr *prompter, stdout io.Writer) error {
	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: false}
	uiInstance := ui.NewUI(stdout, theme)

	fileName := filepath.Base(filePath)
	uiInstance.Warning(fmt.Sprintf("Are you sure you want to permanently delete '%s'? You can't get it back!", fileName))
	fmt.Fprint(stdout, "Type 'yes' to delete it forever: ")

	confirmation, _ := pr.readLine()
	if strings.ToLower(strings.TrimSpace(confirmation)) != "yes" {
		uiInstance.Info("Okay, I won't delete it")
		return nil
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("couldn't delete the file: %w", err)
	}

	uiInstance.Success(fmt.Sprintf("Deleted '%s' - it's gone forever", fileName))
	return nil
}
