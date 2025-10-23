package app

import (
	"os"
	"strings"
)

// FileFilter defines file filtering logic
type FileFilter struct{}

// IsValidFile checks if a file should be included
func (f FileFilter) IsValidFile(entry os.DirEntry) bool {
	return !entry.IsDir() && !strings.HasPrefix(entry.Name(), HiddenFilePrefix)
}

// IsValidCategory checks if a directory should be included as a category
func (f FileFilter) IsValidCategory(entry os.DirEntry) bool {
	return entry.IsDir() &&
		!strings.HasPrefix(entry.Name(), HiddenFilePrefix) &&
		!strings.EqualFold(entry.Name(), ExcludedDir)
}
