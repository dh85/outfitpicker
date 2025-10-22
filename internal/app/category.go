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
	cache  *storage.Manager
	stdout io.Writer
}

func NewCategoryManager(cache *storage.Manager, stdout io.Writer) *CategoryManager {
	return &CategoryManager{cache: cache, stdout: stdout}
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
		return nil, fmt.Errorf("failed to read category %q: %w", filepath.Base(categoryPath), err)
	}

	var files []string
	for _, e := range ents {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			files = append(files, filepath.Join(categoryPath, e.Name()))
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in category %q", filepath.Base(categoryPath))
	}

	return &Category{Path: categoryPath, Files: files}, nil
}

func (cm *CategoryManager) getSelectedFiles(categoryPath string) map[string]bool {
	m := cm.cache.Load()
	return toSet(m[categoryPath])
}

func (cm *CategoryManager) getUnselectedFiles(category *Category) []string {
	seen := cm.getSelectedFiles(category.Path)
	var unselected []string

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
	var available []string

	for _, f := range category.Files {
		if !seen[filepath.Base(f)] {
			available = append(available, f)
		}
	}

	return available
}

func (cm *CategoryManager) displayCategoryInfo(category *Category) {
	seen := cm.getSelectedFiles(category.Path)
	fmt.Fprintf(cm.stdout, "\n📂 Category: %s\n", filepath.Base(category.Path))
	fmt.Fprintf(cm.stdout, "Total files in %q: %d\n", filepath.Base(category.Path), len(category.Files))
	fmt.Fprintf(cm.stdout, "Selected: %d of %d\n", len(seen), len(category.Files))
}

func (cm *CategoryManager) displayMenu() {
	fmt.Fprintln(cm.stdout, "\nOptions:")
	fmt.Fprintln(cm.stdout, "[r] Select a random file in this category")
	fmt.Fprintln(cm.stdout, "[s] Show previously selected files in this category")
	fmt.Fprintln(cm.stdout, "[u] Show unselected files in this category")
	fmt.Fprintln(cm.stdout, "[q] Quit")
	fmt.Fprint(cm.stdout, "Enter your choice: ")
}

func (cm *CategoryManager) showSelectedFiles(categoryPath string) {
	seen := cm.getSelectedFiles(categoryPath)
	if len(seen) == 0 {
		fmt.Fprintln(cm.stdout, "\nno files have been selected yet for this category")
		return
	}

	fmt.Fprintln(cm.stdout, "\n--- Previously Selected Files ---")
	list := mapKeys(seen)
	sort.Strings(list)
	for _, fn := range list {
		fmt.Fprintf(cm.stdout, " - %s\n", fn)
	}
	fmt.Fprintln(cm.stdout, "---------------------------------")
}

func (cm *CategoryManager) showUnselectedFiles(category *Category) {
	fmt.Fprintln(cm.stdout, "\n--- Unselected Files ---")
	unselected := cm.getUnselectedFiles(category)

	if len(unselected) == 0 {
		fmt.Fprintln(cm.stdout, "🎉 all files in this category have been selected")
		return
	}

	for _, u := range unselected {
		fmt.Fprintf(cm.stdout, " - %s\n", u)
	}
	fmt.Fprintln(cm.stdout, "------------------------")
}

func (cm *CategoryManager) handleKeepAction(file FileEntry) error {
	fmt.Fprintf(cm.stdout, "✅ kept and cached: %s\n", file.FileName)
	cm.cache.Add(file.FileName, file.CategoryPath)

	done, err := cm.isCategoryComplete(file.CategoryPath)
	if err != nil {
		fmt.Fprintf(cm.stdout, "warning: could not verify completion: %v\n", err)
		return nil
	}

	if done {
		cm.cache.Clear(file.CategoryPath)
		fmt.Fprintf(cm.stdout, "cache cleared for %q — next random will restart the cycle\n", filepath.Base(file.CategoryPath))
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
	if completed == 0 {
		fmt.Fprintf(cm.stdout, "categories complete: %d/%d\n", completed, total)
	} else {
		suffix := ""
		if len(names) > 0 {
			suffix = " — " + strings.Join(names, ", ")
		}
		fmt.Fprintf(cm.stdout, "categories complete: %d/%d%s\n", completed, total, suffix)
	}
}

func (cm *CategoryManager) isCategoryComplete(catPath string) (bool, error) {
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
	var names []string

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

func (cm *CategoryManager) handleRandomSelection(category *Category, pr *prompter) error {
	available := cm.getAvailableFiles(category)

	for {
		if len(available) == 0 {
			fmt.Fprintf(cm.stdout, "\n🎉 all files in %q have been selected\n", filepath.Base(category.Path))
			cm.cache.Clear(category.Path)
			fmt.Fprintf(cm.stdout, "cache cleared for %q — next random will restart the cycle\n", filepath.Base(category.Path))
			return nil
		}

		idx := rand.Intn(len(available))
		randomFile := available[idx]
		file := FileEntry{
			CategoryPath: category.Path,
			FilePath:     randomFile,
			FileName:     filepath.Base(randomFile),
		}

		fmt.Fprintf(cm.stdout, "\nRandomly selected: %s\n", file.FileName)
		fmt.Fprint(cm.stdout, "Enter (k)eep, (s)kip, or (q)uit: ")

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
			fmt.Fprintln(cm.stdout, "⏩ skipped. selecting another...")
			available = append(available[:idx], available[idx+1:]...)
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

func randomAcrossAll(categories []string, cache *storage.Manager, pr *prompter, stdout io.Writer) error {
	cm := NewCategoryManager(cache, stdout)
	pool := cm.buildFilePool(categories)

	if len(pool) == 0 {
		fmt.Fprintln(stdout, "🎉 all files in all categories have been selected")
		for _, cat := range categories {
			cache.Clear(cat)
		}
		return nil
	}

	file := pool[rand.Intn(len(pool))]
	fmt.Fprintf(stdout, "\n📂 Category: %s\n", filepath.Base(file.CategoryPath))
	fmt.Fprintf(stdout, "Randomly selected: %s\n", file.FileName)
	fmt.Fprint(stdout, "Enter (k)eep, (s)kip, or (q)uit: ")

	action, err := pr.readLineLowerDefault("k")
	if err != nil && !errors.Is(err, io.EOF) {
		fmt.Fprintln(stdout, "invalid action. please try again.")
		return nil
	}

	switch action {
	case "k":
		if err := cm.handleKeepAction(file); err != nil {
			return err
		}
		completed, total, names := cm.getCompletionSummary(categories)
		cm.displayCompletionSummaryFormatted(completed, total, names)
	case "s":
		fmt.Fprintln(stdout, "⏩ skipped. run again for another pick")
	case "q":
		fmt.Fprintln(stdout, "Exiting.")
	default:
		fmt.Fprintln(stdout, "invalid action. please try again.")
	}
	return nil
}

func (cm *CategoryManager) buildFilePool(categories []string) []FileEntry {
	var pool []FileEntry
	m := cm.cache.Load()

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
	return pool
}

func (cm *CategoryManager) displayCompletionSummaryFormatted(completed, total int, names []string) {
	if completed == 0 {
		fmt.Fprintf(cm.stdout, "categories complete: %d/%d\n", completed, total)
	} else {
		suffix := ""
		if len(names) > 0 {
			suffix = " — " + strings.Join(names, ", ")
		}
		fmt.Fprintf(cm.stdout, "categories complete: %d/%d%s\n", completed, total, suffix)
	}
}

func showSelectedAcrossAll(categories []string, cache *storage.Manager, stdout io.Writer) error {
	m := cache.Load()
	var total int

	for _, cat := range categories {
		selected := append([]string(nil), m[cat]...)
		if len(selected) == 0 {
			continue
		}
		total += len(selected)
		sort.Strings(selected)
		fmt.Fprintf(stdout, "\n--- Selected in %s ---\n", filepath.Base(cat))
		for _, s := range selected {
			fmt.Fprintf(stdout, " - %s\n", s)
		}
	}

	if total == 0 {
		fmt.Fprintln(stdout, "\nno files have been selected yet across all categories")
	}
	return nil
}

func showUnselectedAcrossAll(categories []string, cache *storage.Manager, stdout io.Writer) error {
	m := cache.Load()
	var hasUnselected bool

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
			sort.Strings(unselected)
			fmt.Fprintf(stdout, "\n--- Unselected in %s ---\n", filepath.Base(cat))
			for _, u := range unselected {
				fmt.Fprintf(stdout, " - %s\n", u)
			}
		}
	}

	if !hasUnselected {
		fmt.Fprintln(stdout, "\n🎉 all files in all categories have been selected")
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

// Legacy functions for backward compatibility
func categoryComplete(catPath string, cache *storage.Manager) (bool, error) {
	cm := &CategoryManager{cache: cache}
	return cm.isCategoryComplete(catPath)
}

func categoriesCompletionSummary(categories []string, cache *storage.Manager) (int, int, []string) {
	cm := &CategoryManager{cache: cache}
	return cm.getCompletionSummary(categories)
}
