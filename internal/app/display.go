package app

import (
	"io"

	"github.com/dh85/outfitpicker/internal/ui"
)

// Display handles all UI output using the enhanced UI package
type Display struct {
	ui *ui.UI
}

// NewDisplay creates a new Display with enhanced UI
func NewDisplay(writer io.Writer, config AppConfig) *Display {
	theme := ui.Theme{
		UseColors: shouldUseColors(),
		UseEmojis: config.ShowEmojis,
		Compact:   false, // Can be made configurable
	}
	return &Display{ui: ui.NewUI(writer, theme)}
}

// CategoryInfo displays category information
func (d *Display) CategoryInfo(name string, totalFiles, selectedFiles int) {
	d.ui.CategoryInfo(name, totalFiles, selectedFiles)
}

// Menu displays the category menu
func (d *Display) Menu() {
	d.ui.Menu()
}

// SelectedFiles displays previously selected files
func (d *Display) SelectedFiles(categoryName string, files []string) {
	d.ui.SelectedFiles(categoryName, files)
}

// UnselectedFiles displays unselected files
func (d *Display) UnselectedFiles(files []string) {
	d.ui.UnselectedFiles(files)
}

// RandomSelection displays random selection prompt
func (d *Display) RandomSelection(filename string) {
	d.ui.RandomSelection(filename)
}

// KeepAction displays keep confirmation
func (d *Display) KeepAction(filename string) {
	d.ui.KeepAction(filename)
}

// SkipAction displays skip confirmation
func (d *Display) SkipAction(filename string) {
	d.ui.SkipAction(filename)
}

// CompletionSummary displays completion status
func (d *Display) CompletionSummary(completed, total int, names []string) {
	d.ui.CompletionSummary(completed, total, names)
}

// Error displays error messages
func (d *Display) Error(message string) {
	d.ui.Error(message)
}

// Success displays success messages
func (d *Display) Success(message string) {
	d.ui.Success(message)
}

// Info displays info messages
func (d *Display) Info(message string) {
	d.ui.Info(message)
}

// Warning displays warning messages
func (d *Display) Warning(message string) {
	d.ui.Warning(message)
}
