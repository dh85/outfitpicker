// Package app provides the core application logic for the outfit picker CLI.
package app

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dh85/outfitpicker/internal/storage"
	"github.com/dh85/outfitpicker/internal/ui"
)

// Category represents a category with its files and metadata
type Category struct {
	Path  string
	Files []string
}

// FileEntry represents a file with its category context
type FileEntry struct {
	CategoryPath string
	FilePath     string
	FileName     string
}

// CategoryManager handles category operations
type CategoryManager struct {
	cache     *storage.Manager
	stdout    io.Writer
	filePool  []FileEntry
	poolDirty bool
	metrics   *Metrics
}

// newUIInstance creates a UI instance with consistent theme
func (cm *CategoryManager) newUIInstance(compact bool) *ui.UI {
	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: compact}
	return ui.NewUI(cm.stdout, theme)
}

func NewCategoryManager(cache *storage.Manager, stdout io.Writer) *CategoryManager {
	return &CategoryManager{
		cache:     cache,
		stdout:    stdout,
		poolDirty: true,
		metrics:   NewMetrics(),
	}
}

func listCategories(rootAbs string) ([]string, error) {
	ents, err := os.ReadDir(rootAbs)
	if err != nil {
		return nil, fmt.Errorf("failed to read root %q: %w", rootAbs, err)
	}

	var cats []string
	for _, e := range ents {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || strings.EqualFold(e.Name(), "Downloads") {
			continue
		}
		cats = append(cats, filepath.Join(rootAbs, e.Name()))
	}

	sort.Slice(cats, func(i, j int) bool {
		return strings.ToLower(filepath.Base(cats[i])) < strings.ToLower(filepath.Base(cats[j]))
	})
	return cats, nil
}

func (cm *CategoryManager) loadCategory(categoryPath string) (*Category, error) {
	ents, err := os.ReadDir(categoryPath)
	if err != nil {
		return nil, NewFileSystemError(
			fmt.Sprintf("failed to read category %q", filepath.Base(categoryPath)),
			err,
		)
	}

	var files []string
	for _, e := range ents {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			files = append(files, filepath.Join(categoryPath, e.Name()))
		}
	}

	if len(files) == 0 {
		return nil, NewCategoryError(
			fmt.Sprintf("no files found in category %q", filepath.Base(categoryPath)),
			nil,
		)
	}

	return &Category{Path: categoryPath, Files: files}, nil
}

func (cm *CategoryManager) getSelectedFiles(categoryPath string) map[string]bool {
	m := cm.cache.Load()
	return toSet(m[categoryPath])
}

func (cm *CategoryManager) getUnselectedFiles(category *Category) []string {
	seen := cm.getSelectedFiles(category.Path)
	unselected := make([]string, 0, len(category.Files))

	for _, f := range category.Files {
		if !seen[filepath.Base(f)] {
			unselected = append(unselected, filepath.Base(f))
		}
	}

	sort.Strings(unselected)
	return unselected
}

func (cm *CategoryManager) getAvailableFiles(category *Category) []string {
	seen := cm.getSelectedFiles(category.Path)
	available := make([]string, 0, len(category.Files))

	for _, f := range category.Files {
		if !seen[filepath.Base(f)] {
			available = append(available, f)
		}
	}

	return available
}

func (cm *CategoryManager) displayCategoryInfo(category *Category) {
	seen := cm.getSelectedFiles(category.Path)
	uiInstance := cm.newUIInstance(false)
	uiInstance.CategoryInfo(filepath.Base(category.Path), len(category.Files), len(seen))
}

func (cm *CategoryManager) displayMenu() {
	uiInstance := cm.newUIInstance(false)
	uiInstance.Menu()
}

func (cm *CategoryManager) showSelectedFiles(categoryPath string) {
	seen := cm.getSelectedFiles(categoryPath)
	list := mapKeys(seen)
	uiInstance := cm.newUIInstance(false)
	uiInstance.SelectedFiles(filepath.Base(categoryPath), list)
}

func (cm *CategoryManager) showUnselectedFiles(category *Category) {
	unselected := cm.getUnselectedFiles(category)
	uiInstance := cm.newUIInstance(false)
	uiInstance.UnselectedFiles(unselected)
}

func (cm *CategoryManager) handleKeepAction(file FileEntry) error {
	uiInstance := cm.newUIInstance(true)
	uiInstance.KeepAction(file.FileName)
	cm.cache.Add(file.FileName, file.CategoryPath)
	cm.poolDirty = true // Mark pool as dirty when cache changes
	cm.metrics.RecordSelection()

	done, err := cm.isCategoryComplete(file.CategoryPath)
	if err != nil {
		fmt.Fprintf(cm.stdout, "warning: could not verify completion: %v\n", err)
		return nil
	}

	if done {
		cm.cache.Clear(file.CategoryPath)
		fmt.Fprintf(cm.stdout, "\n🎉 You've picked everything from %s! You can start fresh with this folder now.\n", filepath.Base(file.CategoryPath))
	}

	return nil
}

func (cm *CategoryManager) displayCompletionSummary(categoryPath string) {
	root := filepath.Dir(categoryPath)
	cats, err := listCategories(root)
	if err != nil || len(cats) == 0 {
		return
	}

	completed, total, names := cm.getCompletionSummary(cats)
	uiInstance := cm.newUIInstance(true)
	uiInstance.CompletionSummary(completed, total, names)
}

func (cm *CategoryManager) isCategoryComplete(catPath string) (bool, error) {
	// Handle uncategorized files specially
	if catPath == "UNCATEGORIZED" {
		return false, nil // Uncategorized files don't have completion logic
	}

	total, err := categoryFileCount(catPath)
	if err != nil {
		return false, err
	}
	m := cm.cache.Load()
	seen := len(m[catPath])
	return seen >= total && total > 0, nil
}

func (cm *CategoryManager) getCompletionSummary(categories []string) (int, int, []string) {
	m := cm.cache.Load()
	completed := 0
	names := make([]string, 0, len(categories))

	for _, cat := range categories {
		total, err := categoryFileCount(cat)
		if err != nil || total == 0 {
			continue
		}
		seen := len(m[cat])
		if seen >= total {
			completed++
			names = append(names, filepath.Base(cat))
		}
	}

	sort.Strings(names)
	return completed, len(categories), names
}

// getFilePool returns cached file pool or rebuilds if dirty
func (cm *CategoryManager) getFilePool(categories, uncategorized []string) []FileEntry {
	if cm.poolDirty || cm.filePool == nil {
		cm.filePool = cm.buildFilePool(categories, uncategorized)
		cm.poolDirty = false
	}
	return cm.filePool
}

func (cm *CategoryManager) handleRandomSelection(category *Category, pr *prompter) error {
	available := cm.getAvailableFiles(category)
	skipped := make(map[string]bool)

	for {
		// Filter out skipped files
		currentAvailable := make([]string, 0, len(available))
		for _, f := range available {
			if !skipped[f] {
				currentAvailable = append(currentAvailable, f)
			}
		}

		if len(currentAvailable) == 0 {
			if len(available) == 0 {
				fmt.Fprintf(cm.stdout, "\n🎉 Amazing! You've picked all the outfits from %s!\n", filepath.Base(category.Path))
				cm.cache.Clear(category.Path)
				fmt.Fprintf(cm.stdout, "I've reset this folder so you can pick from it again!\n")
				return nil
			}
			fmt.Fprintln(cm.stdout, "⚠️ You've skipped all available outfits in this category.")
			fmt.Fprint(cm.stdout, "Try again with the same outfits? [y/N]: ")
			response, _ := pr.readLineLower()
			if response == "y" {
				skipped = make(map[string]bool)
				continue
			}
			return nil
		}

		idx := rand.Intn(len(currentAvailable))
		randomFile := currentAvailable[idx]
		file := FileEntry{
			CategoryPath: category.Path,
			FilePath:     randomFile,
			FileName:     filepath.Base(randomFile),
		}

		theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: false}
		uiInstance := ui.NewUI(cm.stdout, theme)
		uiInstance.RandomSelection(file.FileName)

		action, err := pr.readLineLowerDefault("k")
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Fprintln(cm.stdout, "invalid action. please try again.")
			continue
		}

		switch action {
		case "k":
			if err := cm.handleKeepAction(file); err != nil {
				return err
			}
			cm.displayCompletionSummary(category.Path)
			return nil
		case "s":
			uiInstance := cm.newUIInstance(true)
			uiInstance.SkipAction(file.FileName)
			skipped[randomFile] = true
			cm.metrics.RecordSkip()
		case "q":
			fmt.Fprintln(cm.stdout, "Exiting.")
			return nil
		default:
			fmt.Fprintln(cm.stdout, "invalid action. please try again.")
		}
	}
}

func runCategoryFlow(categoryPath string, cache *storage.Manager, pr *prompter, stdout io.Writer) error {
	cm := NewCategoryManager(cache, stdout)

	category, err := cm.loadCategory(categoryPath)
	if err != nil {
		return err
	}

	cm.displayCategoryInfo(category)
	cm.displayMenu()

	choice, _ := pr.readLineLower()
	switch choice {
	case "s":
		cm.showSelectedFiles(category.Path)
	case "u":
		cm.showUnselectedFiles(category)
	case "r":
		return cm.handleRandomSelection(category, pr)
	case "q":
		fmt.Fprintln(stdout, "Exiting.")
	default:
		return fmt.Errorf("invalid selection")
	}
	return nil
}

func randomAcrossAll(categories, uncategorized []string, cache *storage.Manager, pr *prompter, stdout io.Writer) error {
	cm := NewCategoryManager(cache, stdout)
	pool := cm.getFilePool(categories, uncategorized)
	skipped := make(map[string]bool)
	defer cm.metrics.LogSession()

	// Check if pool is empty from the start
	if len(pool) == 0 {
		fmt.Fprintln(stdout, "🎉 Amazing! You've picked all your outfits!")
		for _, cat := range categories {
			cache.Clear(cat)
		}
		if len(uncategorized) > 0 {
			cache.Clear("UNCATEGORIZED")
		}
		fmt.Fprintln(stdout, "Starting fresh - you can pick from all your outfits again!")
		return nil
	}

	for {
		// Filter out skipped files from pool
		available := make([]FileEntry, 0, len(pool))
		for _, file := range pool {
			if !skipped[file.FilePath] {
				available = append(available, file)
			}
		}

		if len(available) == 0 {
			fmt.Fprintln(stdout, "⚠️ You've skipped all available outfits in this session.")
			fmt.Fprint(stdout, "Try again with the same outfits? [y/N]: ")
			response, _ := pr.readLineLower()
			if response == "y" {
				skipped = make(map[string]bool)
				continue
			}
			return nil
		}

		file := available[rand.Intn(len(available))]
		theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: false}
		uiInstance := ui.NewUI(stdout, theme)

		if file.CategoryPath == "UNCATEGORIZED" {
			fmt.Fprintf(stdout, "\n📄 From your other outfits\n")
		} else {
			fmt.Fprintf(stdout, "\n📂 From your %s collection\n", filepath.Base(file.CategoryPath))
		}
		uiInstance.RandomSelection(file.FileName)

		action, err := pr.readLineLowerDefault("k")
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Fprintln(stdout, "invalid action. please try again.")
			continue
		}

		switch action {
		case "k":
			if err := cm.handleKeepAction(file); err != nil {
				return err
			}
			completed, total, names := cm.getCompletionSummary(categories)
			cm.displayCompletionSummaryFormatted(completed, total, names)
			return nil
		case "s":
			uiInstance := cm.newUIInstance(true)
			uiInstance.SkipAction(file.FileName)
			skipped[file.FilePath] = true
			cm.metrics.RecordSkip()
		case "d":
			return handleDeleteFile(file.FilePath, pr, stdout)
		case "q":
			fmt.Fprintln(stdout, "Exiting.")
			return nil
		default:
			fmt.Fprintln(stdout, "invalid action. please try again.")
		}
	}
}

func (cm *CategoryManager) buildFilePool(categories, uncategorized []string) []FileEntry {
	var pool []FileEntry
	m := cm.cache.Load()

	// Add categorized files
	for _, cat := range categories {
		ents, _ := os.ReadDir(cat)
		seen := toSet(m[cat])

		for _, e := range ents {
			if e.IsDir() || strings.HasPrefix(e.Name(), ".") || seen[e.Name()] {
				continue
			}
			pool = append(pool, FileEntry{
				CategoryPath: cat,
				FilePath:     filepath.Join(cat, e.Name()),
				FileName:     e.Name(),
			})
		}
	}

	// Add uncategorized files
	if len(uncategorized) > 0 {
		const uncategorizedKey = "UNCATEGORIZED"
		seen := toSet(m[uncategorizedKey])

		for _, f := range uncategorized {
			name := filepath.Base(f)
			if !seen[name] {
				pool = append(pool, FileEntry{
					CategoryPath: uncategorizedKey,
					FilePath:     f,
					FileName:     name,
				})
			}
		}
	}

	return pool
}

func (cm *CategoryManager) displayCompletionSummaryFormatted(completed, total int, names []string) {
	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: true}
	uiInstance := ui.NewUI(cm.stdout, theme)
	uiInstance.CompletionSummary(completed, total, names)
}

func showSelectedAcrossAll(categories, uncategorized []string, cache *storage.Manager, stdout io.Writer) error {
	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: false}
	uiInstance := ui.NewUI(stdout, theme)
	m := cache.Load()
	var total int

	// Show categorized selected files
	for _, cat := range categories {
		selected := append([]string(nil), m[cat]...)
		if len(selected) == 0 {
			continue
		}
		total += len(selected)
		uiInstance.SelectedFiles(filepath.Base(cat), selected)
	}

	// Show uncategorized selected files
	if len(uncategorized) > 0 {
		const uncategorizedKey = "UNCATEGORIZED"
		selected := append([]string(nil), m[uncategorizedKey]...)
		if len(selected) > 0 {
			total += len(selected)
			uiInstance.SelectedFiles("Uncategorized", selected)
		}
	}

	if total == 0 {
		uiInstance.Info("You haven't picked any outfits from here yet")
	}
	return nil
}

func showUnselectedAcrossAll(categories, uncategorized []string, cache *storage.Manager, stdout io.Writer) error {
	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: false}
	uiInstance := ui.NewUI(stdout, theme)
	m := cache.Load()
	var hasUnselected bool

	// Show categorized unselected files
	for _, cat := range categories {
		ents, _ := os.ReadDir(cat)
		seen := toSet(m[cat])
		var unselected []string

		for _, e := range ents {
			if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") && !seen[e.Name()] {
				unselected = append(unselected, e.Name())
			}
		}

		if len(unselected) > 0 {
			hasUnselected = true
			fmt.Fprintf(stdout, "\n📁 %s\n", filepath.Base(cat))
			uiInstance.UnselectedFiles(unselected)
		}
	}

	// Show uncategorized unselected files
	if len(uncategorized) > 0 {
		const uncategorizedKey = "UNCATEGORIZED"
		seen := toSet(m[uncategorizedKey])
		var unselected []string

		for _, f := range uncategorized {
			name := filepath.Base(f)
			if !seen[name] {
				unselected = append(unselected, name)
			}
		}

		if len(unselected) > 0 {
			hasUnselected = true
			fmt.Fprintf(stdout, "\n📄 Uncategorized\n")
			uiInstance.UnselectedFiles(unselected)
		}
	}

	if !hasUnselected {
		uiInstance.Success("You've picked all the outfits!")
	}
	return nil
}

// Utility functions
func baseNames(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = filepath.Base(p)
	}
	return out
}

func toSet(list []string) map[string]bool {
	out := make(map[string]bool, len(list))
	for _, v := range list {
		out[v] = true
	}
	return out
}

func mapKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func categoryFileCount(catPath string) (int, error) {
	ents, err := os.ReadDir(catPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read category %q: %w", filepath.Base(catPath), err)
	}
	total := 0
	for _, e := range ents {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			total++
		}
	}
	return total, nil
}

// shouldUseColors determines if colors should be used based on environment
func shouldUseColors() bool {
	// Check if output is a terminal and colors are supported
	if term := os.Getenv("TERM"); term == "dumb" || term == "" {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return true
}

// Legacy functions for backward compatibility
func categoryComplete(catPath string, cache *storage.Manager) (bool, error) {
	cm := &CategoryManager{cache: cache}
	return cm.isCategoryComplete(catPath)
}

func categoriesCompletionSummary(categories []string, cache *storage.Manager) (int, int, []string) {
	cm := &CategoryManager{cache: cache}
	return cm.getCompletionSummary(categories)
}
