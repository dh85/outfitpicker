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

// I18n interface for localization
type I18n interface {
	T(key string, args ...interface{}) string
}

// UI handles all user interface operations
type UI struct {
	writer   io.Writer
	theme    Theme
	i18n     I18n
	langCode string
	isRTL    bool
}

// NewUI creates a new UI instance
func NewUI(writer io.Writer, theme Theme) *UI {
	return &UI{writer: writer, theme: theme}
}

// NewUIWithI18n creates a new UI instance with internationalization
func NewUIWithI18n(writer io.Writer, theme Theme, i18n I18n) *UI {
	return &UI{writer: writer, theme: theme, i18n: i18n}
}

// NewUIWithI18nAndLang creates a new UI instance with internationalization and language support
func NewUIWithI18nAndLang(writer io.Writer, theme Theme, i18n I18n, langCode string) *UI {
	return &UI{
		writer:   writer,
		theme:    theme,
		i18n:     i18n,
		langCode: langCode,
		isRTL:    IsRTL(langCode),
	}
}

// Header displays a formatted header with title
func (u *UI) Header(title string) {
	if u.theme.Compact {
		fmt.Fprintf(u.writer, "%s%s\n", u.icon(IconMenu), title)
		return
	}

	separator := strings.Repeat(IconSeparator, len(title)+4)
	fmt.Fprintf(u.writer, "\n%s\n", u.colorize(separator, Cyan))
	fmt.Fprintf(u.writer, "%s%s %s\n", u.colorize(IconSeparator, Cyan), u.colorize(title, Bold+Cyan), u.colorize(IconSeparator, Cyan))
	fmt.Fprintf(u.writer, "%s\n\n", u.colorize(separator, Cyan))
}

// CategoryInfo displays category information with enhanced formatting
func (u *UI) CategoryInfo(name string, totalFiles, selectedFiles int) {
	if u.theme.Compact {
		fmt.Fprintf(u.writer, "%s%s (%d/%d)\n", u.icon(IconFolder), name, selectedFiles, totalFiles)
		return
	}

	fmt.Fprintf(u.writer, "\n%s%s\n", u.icon(IconFolder), u.colorize(name, Bold+Blue))
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

	fmt.Fprintf(u.writer, "\n%s %s\n", u.icon(IconMenu), u.colorize("What would you like to do?", Bold))
	options := []struct {
		key, desc, icon string
	}{
		{"r", "Pick a random outfit from this folder", IconRandom},
		{"s", "Show outfits I've already picked", IconCheck},
		{"u", "Show outfits I haven't picked yet", IconFile},
		{"q", "Go back", IconExit},
	}

	for _, opt := range options {
		fmt.Fprintf(u.writer, "  %s [%s] %s\n",
			u.icon(opt.icon),
			u.colorize(opt.key, Bold+Green),
			opt.desc)
	}
	fmt.Fprint(u.writer, "\n"+u.colorize("Choose a letter: ", Bold))
}

// MainMenu displays the main category selection menu
func (u *UI) MainMenu(categories, uncategorized []string) {
	title := "Outfit Picker"
	if u.i18n != nil {
		title = u.i18n.T("outfit_picker")
	}
	title = FormatRTL(title, u.isRTL)
	u.Header(title)

	if len(categories) > 0 {
		folderTitle := "Outfit Folders"
		if u.i18n != nil {
			folderTitle = u.i18n.T("outfit_folders")
		}
		fmt.Fprintf(u.writer, "%s%s\n", u.icon(IconFolder), u.colorize(folderTitle, Bold+Blue))
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
	}

	if len(uncategorized) > 0 {
		nextNum := len(categories) + 1
		otherOutfits := "Other Outfits"
		filesCount := "(%d files)"
		if u.i18n != nil {
			otherOutfits = u.i18n.T("other_outfits")
			filesCount = u.i18n.T("files_count")
		}
		fmt.Fprintf(u.writer, "  [%s] %s %s %s\n",
			u.colorize(fmt.Sprintf("%d", nextNum), Bold+Green),
			u.icon(IconFile),
			otherOutfits,
			fmt.Sprintf(filesCount, len(uncategorized)))
	}

	whatToDo := "What would you like to do?"
	if u.i18n != nil {
		whatToDo = u.i18n.T("what_would_you_like")
	}
	fmt.Fprintf(u.writer, "\n%s %s\n", u.icon(IconMenu), u.colorize(whatToDo, Bold+Magenta))

	globalOptions := []struct {
		key, descKey, icon string
	}{
		{"r", "pick_random_outfit", IconRandom},
		{"s", "show_already_picked", IconCheck},
		{"u", "show_not_picked", IconFile},
		{"m", "let_me_choose", IconMenu},
		{"q", "exit", IconExit},
	}

	for _, opt := range globalOptions {
		desc := opt.descKey
		if u.i18n != nil {
			desc = u.i18n.T(opt.descKey)
		}
		fmt.Fprintf(u.writer, "  [%s] %s %s\n",
			u.colorize(opt.key, Bold+Green),
			u.icon(opt.icon),
			desc)
	}

	choosePrompt := "Choose a number or letter: "
	if u.i18n != nil {
		choosePrompt = u.i18n.T("choose_number_letter")
	}
	choosePrompt = FormatRTL(choosePrompt, u.isRTL)
	fmt.Fprint(u.writer, "\n"+u.colorize(choosePrompt, Bold))
}

// SelectedFiles displays previously selected files
func (u *UI) SelectedFiles(categoryName string, files []string) {
	if len(files) == 0 {
		fmt.Fprintf(u.writer, "\n%s %s\n",
			u.icon(IconInfo),
			u.colorize("You haven't picked any outfits from here yet", Yellow))
		return
	}

	if u.theme.Compact {
		fmt.Fprintf(u.writer, "\nYou picked (%d):\n", len(files))
	} else {
		fmt.Fprintf(u.writer, "\n%s %s (%d outfits)\n",
			u.icon(IconCheck),
			u.colorize("Outfits You've Already Picked", Bold+Green),
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
			u.colorize("You've picked all the outfits from here!", Green))
		return
	}

	if u.theme.Compact {
		fmt.Fprintf(u.writer, "\nNot picked yet (%d):\n", len(files))
	} else {
		fmt.Fprintf(u.writer, "\n%s %s (%d outfits)\n",
			u.icon(IconFile),
			u.colorize("Outfits You Haven't Picked Yet", Bold+Yellow),
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
	pickedMsg := "I picked this outfit for you"
	prompt := "Do you want to (k)eep it, (s)kip it, (d)elete it, or (q)uit? "
	if u.i18n != nil {
		pickedMsg = u.i18n.T("picked_outfit_for_you")
		prompt = u.i18n.T("keep_skip_delete_quit")
	}
	fmt.Fprintf(u.writer, "\n%s %s: %s\n",
		u.icon(IconRandom),
		u.colorize(pickedMsg, Bold+Magenta),
		u.colorize(filename, Bold+Cyan))
	fmt.Fprint(u.writer, u.colorize(prompt, Bold))
}

// KeepAction displays keep confirmation
func (u *UI) KeepAction(filename string) {
	fmt.Fprintf(u.writer, "%s%s: %s\n",
		u.icon(IconCheck),
		u.colorize("Great choice! I've saved", Green),
		filename)
}

// SkipAction displays skip confirmation
func (u *UI) SkipAction(filename string) {
	fmt.Fprintf(u.writer, "%s%s: %s\n",
		u.icon(IconWarning),
		u.colorize("Skipped", Yellow),
		filename)
}

// CompletionSummary displays completion status
func (u *UI) CompletionSummary(completed, total int, names []string) {
	if completed == 0 {
		fmt.Fprintf(u.writer, "%sOutfit folders complete: %s\n",
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

	fmt.Fprintf(u.writer, "%sOutfit folders complete: %s%s\n",
		u.icon(IconCheck),
		u.colorize(fmt.Sprintf("%d/%d", completed, total), color),
		suffix)
}

// Error displays error messages
func (u *UI) Error(message string) {
	fmt.Fprintf(u.writer, "%s%s: %s\n",
		u.icon(IconCross),
		u.colorize("Error", Bold+Red),
		message)
}

// Success displays success messages
func (u *UI) Success(message string) {
	fmt.Fprintf(u.writer, "%s%s\n",
		u.icon(IconCheck),
		u.colorize(message, Green))
}

// Info displays info messages
func (u *UI) Info(message string) {
	fmt.Fprintf(u.writer, "%s%s\n",
		u.icon(IconInfo),
		message)
}

// Warning displays warning messages
func (u *UI) Warning(message string) {
	fmt.Fprintf(u.writer, "%s%s\n",
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
	return emoji + " "
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

// UncategorizedOnlyMenu displays menu for uncategorized files only
func (u *UI) UncategorizedOnlyMenu(fileCount int) {
	u.Header("Outfit Picker")

	fmt.Fprintf(u.writer, "%s%s (%d outfits available)\n",
		u.icon(IconFile),
		u.colorize("Your Outfits", Bold+Blue),
		fileCount)

	fmt.Fprintf(u.writer, "\n%s %s\n", u.icon(IconMenu), u.colorize("What would you like to do?", Bold+Magenta))
	options := []struct {
		key, desc, icon string
	}{
		{"r", "Pick a random outfit for me", IconRandom},
		{"s", "Show outfits I've already picked", IconCheck},
		{"u", "Show outfits I haven't picked yet", IconFile},
		{"m", "Let me choose an outfit myself", IconMenu},
		{"q", "Exit", IconExit},
	}

	for _, opt := range options {
		fmt.Fprintf(u.writer, "  [%s] %s %s\n",
			u.colorize(opt.key, Bold+Green),
			u.icon(opt.icon),
			opt.desc)
	}
	fmt.Fprint(u.writer, "\n"+u.colorize("Choose a letter: ", Bold))
}

// UncategorizedInfo displays uncategorized file information
func (u *UI) UncategorizedInfo(totalFiles, selectedFiles int) {
	if u.theme.Compact {
		fmt.Fprintf(u.writer, "%sUncategorized (%d/%d)\n", u.icon(IconFile), selectedFiles, totalFiles)
		return
	}

	fmt.Fprintf(u.writer, "\n%s %s\n", u.icon(IconFile), u.colorize("Uncategorized Files", Bold+Blue))
	fmt.Fprintf(u.writer, "   %s Total files: %s\n", u.icon(IconFile), u.colorize(fmt.Sprintf("%d", totalFiles), Green))
	fmt.Fprintf(u.writer, "   %s Selected: %s\n", u.icon(IconCheck), u.colorize(fmt.Sprintf("%d", selectedFiles), Yellow))

	if selectedFiles > 0 && totalFiles > 0 {
		percentage := float64(selectedFiles) / float64(totalFiles) * 100
		progressBar := u.createProgressBar(percentage, 20)
		fmt.Fprintf(u.writer, "   Progress: %s %.1f%%\n", progressBar, percentage)
	}
}

// ManualSelectionMenu displays grouped file selection menu
func (u *UI) ManualSelectionMenu(groupCount, totalFiles int) {
	u.Header("Choose Your Outfit")

	fmt.Fprintf(u.writer, "%s%d outfit collections with %d total outfits\n",
		u.icon(IconFolder), groupCount, totalFiles)

	fmt.Fprintf(u.writer, "\n  [%s] %s Go back\n",
		u.colorize("q", Bold+Green),
		u.icon(IconExit))

	fmt.Fprint(u.writer, "\n"+u.colorize("Choose a number or 'q' to go back: ", Bold))
}

// DisplayFileGroup displays a group of files with numbering
func (u *UI) DisplayFileGroup(groupName string, files []string, selectedFiles map[string]bool, startIndex int) int {
	if groupName == "Uncategorized" {
		fmt.Fprintf(u.writer, "\n%s %s\n", u.icon(IconFile), u.colorize("Other Outfits", Bold+Blue))
	} else {
		fmt.Fprintf(u.writer, "\n%s %s\n", u.icon(IconFolder), u.colorize(groupName, Bold+Blue))
	}

	fileIndex := startIndex
	for _, fileName := range files {
		status := ""
		if selectedFiles[fileName] {
			status = u.colorize(" (already picked)", Green)
		}

		fmt.Fprintf(u.writer, "  [%s] %s %s%s\n",
			u.colorize(fmt.Sprintf("%d", fileIndex), Bold+Green),
			u.icon(IconFile),
			fileName,
			status)
		fileIndex++
	}

	return fileIndex
}
