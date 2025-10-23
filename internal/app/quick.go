package app

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/dh85/outfitpicker/internal/storage"
)

// QuickModeRandom handles power user quick selection
func QuickModeRandom(rootPath string, categoryName string, stdout io.Writer) error {
	return QuickModeRandomWithI18n(rootPath, categoryName, stdout, nil)
}

// QuickModeRandomWithI18n handles power user quick selection with i18n support
func QuickModeRandomWithI18n(rootPath string, categoryName string, stdout io.Writer, i18n *I18n) error {
	cache, err := storage.NewManager(rootPath)
	if err != nil {
		return err
	}

	rootAbs, err := filepath.Abs(rootPath)
	if err != nil {
		return err
	}

	categories, err := listCategories(rootAbs)
	if err != nil {
		return err
	}

	var targetCategory string
	if categoryName != "" {
		for _, cat := range categories {
			if strings.EqualFold(filepath.Base(cat), categoryName) {
				targetCategory = cat
				break
			}
		}
		if targetCategory == "" {
			if i18n != nil {
				return fmt.Errorf("%s", i18n.T("category_not_found", categoryName))
			}
			return fmt.Errorf("category %q not found", categoryName)
		}
	}

	var pool []FileEntry
	if targetCategory != "" {
		pool = buildQuickFilePool([]string{targetCategory}, nil, cache)
	} else {
		uncategorized, _ := listUncategorizedFiles(rootAbs)
		pool = buildQuickFilePool(categories, uncategorized, cache)
	}

	if len(pool) == 0 {
		msg := "No outfits available"
		if i18n != nil {
			msg = i18n.T("no_outfits_available")
		}
		fmt.Fprintln(stdout, msg)
		return nil
	}

	file := pool[rand.Intn(len(pool))]
	cache.Add(file.FileName, file.CategoryPath)

	msg := "✅ Selected: %s"
	if i18n != nil {
		msg = i18n.T("selected_outfit")
	}
	fmt.Fprintf(stdout, msg+"\n", file.FileName)
	return nil
}

func buildQuickFilePool(categories, uncategorized []string, cache *storage.Manager) []FileEntry {
	var pool []FileEntry
	m := cache.Load()

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
