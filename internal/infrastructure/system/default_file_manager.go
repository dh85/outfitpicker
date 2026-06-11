package system

import (
	"os"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

// DefaultFileManager provides default filesystem operations.
type DefaultFileManager struct{}

// NewDefaultFileManager creates a new default file manager.
func NewDefaultFileManager() *DefaultFileManager {
	return &DefaultFileManager{}
}

// ReadDir reads directory entries and returns them as FileEntry slice.
func (d *DefaultFileManager) ReadDir(path string) ([]entities.FileEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := make([]entities.FileEntry, len(entries))
	for i, entry := range entries {
		result[i] = entities.FileEntry{
			FileName:    entry.Name(),
			IsDirectory: entry.IsDir(),
		}
	}

	return result, nil
}

// FileExists checks if a file exists.
func (d *DefaultFileManager) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
