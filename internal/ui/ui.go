package ui

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Colors for terminal output
const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
)

// Icons for better visual presentation
const (
	IconFolder    = "📂"
	IconFile      = "📄"
	IconCheck     = "✅"
	IconCross     = "❌"
	IconInfo      = "ℹ️"
	IconWarning   = "⚠️"
	IconSuccess   = "🎉"
	IconRandom    = "🎲"
	IconMenu      = "📋"
	IconConfig    = "⚙️"
	IconExit      = "👋"
	IconArrow     = "➤"
	IconBullet    = "•"
	IconSeparator = "─"
)

// Theme represents UI styling configuration
type Theme struct {
	UseColors bool
	UseEmojis bool
	Compact   bool
}

// UI handles all user interface operations
type UI struct {
	writer io.Writer
	theme  Theme
}

// NewUI creates a new UI instance
func NewUI(writer io.Writer, theme Theme) *UI {
	return &UI{writer: writer, theme: theme}
}

// Header displays a formatted header with title
func (u *UI) Header(title string) {
	if u.theme.Compact {
		fmt.Fprintf(u.writer, "%s %s\n", u.icon(IconMenu), title)
		return
	}

	separator := strings.Repeat(IconSeparator, len(title)+4)
	fmt.Fprintf(u.writer, "\n%s\n", u.colorize(separator, Cyan))
	fmt.Fprintf(u.writer, "%s %s %s\n", u.colorize(IconSeparator, Cyan), u.colorize(title, Bold+Cyan), u.colorize(IconSeparator, Cyan))
	fmt.Fprintf(u.writer, "%s\n\n", u.colorize(separator, Cyan))
}

// CategoryInfo displays category information with enhanced formatting
func (u *UI) CategoryInfo(name string, totalFiles, selectedFiles int) {
	if u.theme.Compact {
		fmt.Fprintf(u.writer, "%s %s (%d/%d)\n", u.icon(IconFolder), name, selectedFiles, totalFiles)
		return
	}

	fmt.Fprintf(u.writer, "\n%s %s\n", u.icon(IconFolder), u.colorize(name, Bold+Blue))
	fmt.Fprintf(u.writer, "   %s Total files: %s\n", u.icon(IconFile), u.colorize(fmt.Sprintf("%d", totalFiles), Green))
	fmt.Fprintf(u.writer, "   %s Selected: %s\n", u.icon(IconCheck), u.colorize(fmt.Sprintf("%d", selectedFiles), Yellow))

	if selectedFiles > 0 && totalFiles > 0 {
		percentage := float64(selectedFiles) / float64(totalFiles) * 100
		progressBar := u.createProgressBar(percentage, 20)
		fmt.Fprintf(u.writer, "   Progress: %s %.1f%%\n", progressBar, percentage)
	}
}

// Menu displays the main menu with enhanced formatting
func (u *UI) Menu() {
	if u.theme.Compact {
		fmt.Fprint(u.writer, "[r]andom [s]elected [u]nselected [q]uit: ")
		return
	}

	fmt.Fprintf(u.writer, "\n%s %s\n", u.icon(IconMenu), u.colorize("Options", Bold))
	options := []struct {
		key, desc, icon string
	}{
		{"r", "Select a random file in this category", IconRandom},
		{"s", "Show previously selected files", IconCheck},
		{"u", "Show unselected files", IconFile},
		{"q", "Quit", IconExit},
	}

	for _, opt := range options {
		fmt.Fprintf(u.writer, "  %s [%s] %s\n",
			u.icon(opt.icon),
			u.colorize(opt.key, Bold+Green),
			opt.desc)
	}
	fmt.Fprint(u.writer, "\n"+u.colorize("Enter your choice: ", Bold))
}

// MainMenu displays the main category selection menu
func (u *UI) MainMenu(categories []string) {
	u.Header("Outfit Picker")

	fmt.Fprintf(u.writer, "%s %s\n", u.icon(IconFolder), u.colorize("Categories", Bold+Blue))
	for i, c := range categories {
		name := strings.TrimSuffix(c, "/")
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}
		fmt.Fprintf(u.writer, "  [%s] %s %s\n",
			u.colorize(fmt.Sprintf("%d", i+1), Bold+Green),
			u.icon(IconFolder),
			name)
	}

	fmt.Fprintf(u.writer, "\n%s %s\n", u.icon(IconMenu), u.colorize("All-categories options", Bold+Magenta))
	globalOptions := []struct {
		key, desc, icon string
	}{
		{"r", "Select a random file from any category", IconRandom},
		{"s", "Show previously selected files from all categories", IconCheck},
		{"u", "Show unselected files from all categories", IconFile},
		{"q", "Quit", IconExit},
	}

	for _, opt := range globalOptions {
		fmt.Fprintf(u.writer, "  [%s] %s %s\n",
			u.colorize(opt.key, Bold+Green),
			u.icon(opt.icon),
			opt.desc)
	}
	fmt.Fprint(u.writer, "\n"+u.colorize("Enter a category number or option: ", Bold))
}

// SelectedFiles displays previously selected files
func (u *UI) SelectedFiles(categoryName string, files []string) {
	if len(files) == 0 {
		fmt.Fprintf(u.writer, "\n%s %s\n",
			u.icon(IconInfo),
			u.colorize("No files have been selected yet for this category", Yellow))
		return
	}

	if u.theme.Compact {
		fmt.Fprintf(u.writer, "\nSelected (%d):\n", len(files))
	} else {
		fmt.Fprintf(u.writer, "\n%s %s (%d files)\n",
			u.icon(IconCheck),
			u.colorize("Previously Selected Files", Bold+Green),
			len(files))
	}

	sort.Strings(files)
	for i, f := range files {
		if u.theme.Compact {
			fmt.Fprintf(u.writer, "  %d. %s\n", i+1, f)
		} else {
			fmt.Fprintf(u.writer, "  %s %s %s\n",
				u.icon(IconBullet),
				u.icon(IconFile),
				f)
		}
	}
}

// UnselectedFiles displays unselected files
func (u *UI) UnselectedFiles(files []string) {
	if len(files) == 0 {
		fmt.Fprintf(u.writer, "\n%s %s\n",
			u.icon(IconSuccess),
			u.colorize("All files in this category have been selected!", Green))
		return
	}

	if u.theme.Compact {
		fmt.Fprintf(u.writer, "\nUnselected (%d):\n", len(files))
	} else {
		fmt.Fprintf(u.writer, "\n%s %s (%d files)\n",
			u.icon(IconFile),
			u.colorize("Unselected Files", Bold+Yellow),
			len(files))
	}

	for i, f := range files {
		if u.theme.Compact {
			fmt.Fprintf(u.writer, "  %d. %s\n", i+1, f)
		} else {
			fmt.Fprintf(u.writer, "  %s %s %s\n",
				u.icon(IconBullet),
				u.icon(IconFile),
				f)
		}
	}
}

// RandomSelection displays random selection prompt
func (u *UI) RandomSelection(filename string) {
	fmt.Fprintf(u.writer, "\n%s %s: %s\n",
		u.icon(IconRandom),
		u.colorize("Randomly selected", Bold+Magenta),
		u.colorize(filename, Bold+Cyan))
	fmt.Fprint(u.writer, u.colorize("Enter (k)eep, (s)kip, or (q)uit: ", Bold))
}

// KeepAction displays keep confirmation
func (u *UI) KeepAction(filename string) {
	fmt.Fprintf(u.writer, "%s %s: %s\n",
		u.icon(IconCheck),
		u.colorize("Kept and cached", Green),
		filename)
}

// SkipAction displays skip confirmation
func (u *UI) SkipAction(filename string) {
	fmt.Fprintf(u.writer, "%s %s: %s\n",
		u.icon(IconWarning),
		u.colorize("Skipped", Yellow),
		filename)
}

// CompletionSummary displays completion status
func (u *UI) CompletionSummary(completed, total int, names []string) {
	if completed == 0 {
		fmt.Fprintf(u.writer, "%s Categories complete: %s\n",
			u.icon(IconInfo),
			u.colorize(fmt.Sprintf("%d/%d", completed, total), Yellow))
		return
	}

	suffix := ""
	if len(names) > 0 {
		suffix = " — " + strings.Join(names, ", ")
	}

	color := Yellow
	if completed == total {
		color = Green
	}

	fmt.Fprintf(u.writer, "%s Categories complete: %s%s\n",
		u.icon(IconCheck),
		u.colorize(fmt.Sprintf("%d/%d", completed, total), color),
		suffix)
}

// Error displays error messages
func (u *UI) Error(message string) {
	fmt.Fprintf(u.writer, "%s %s: %s\n",
		u.icon(IconCross),
		u.colorize("Error", Bold+Red),
		message)
}

// Success displays success messages
func (u *UI) Success(message string) {
	fmt.Fprintf(u.writer, "%s %s\n",
		u.icon(IconCheck),
		u.colorize(message, Green))
}

// Info displays info messages
func (u *UI) Info(message string) {
	fmt.Fprintf(u.writer, "%s %s\n",
		u.icon(IconInfo),
		message)
}

// Warning displays warning messages
func (u *UI) Warning(message string) {
	fmt.Fprintf(u.writer, "%s %s\n",
		u.icon(IconWarning),
		u.colorize(message, Yellow))
}

// Separator displays a visual separator
func (u *UI) Separator() {
	if u.theme.Compact {
		fmt.Fprintln(u.writer, "---")
		return
	}
	fmt.Fprintln(u.writer, u.colorize(strings.Repeat(IconSeparator, 50), Gray))
}

// Helper methods

func (u *UI) colorize(text, color string) string {
	if !u.theme.UseColors {
		return text
	}
	return color + text + Reset
}

func (u *UI) icon(emoji string) string {
	if !u.theme.UseEmojis {
		return ""
	}
	return emoji
}

func (u *UI) createProgressBar(percentage float64, width int) string {
	if !u.theme.UseColors && !u.theme.UseEmojis {
		filled := int(percentage / 100 * float64(width))
		return "[" + strings.Repeat("=", filled) + strings.Repeat("-", width-filled) + "]"
	}

	filled := int(percentage / 100 * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return u.colorize("["+bar+"]", Green)
}
