package cli

import (
	"strings"
	"unicode"
)

const (
	uiReset  = "\x1b[0m"
	uiBold   = "\x1b[1m"
	uiGreen  = "\x1b[32m"
	uiBlue   = "\x1b[34m"
	uiCyan   = "\x1b[36m"
	uiYellow = "\x1b[33m"
	uiRed    = "\x1b[31m"
)

func Header(title string) {
	HeaderWithConsole(nil, title)
}

func HeaderWithConsole(console Console, title string) {
	separator := repeatLine("─", len(title)+4)
	terminal := consoleOrDefault(console)
	terminal.Println()
	terminal.Println(Colorize(separator, uiCyan))
	terminal.Printf("%s %s %s\n", Colorize("─", uiCyan), Colorize(title, uiBold+uiCyan), Colorize("─", uiCyan))
	terminal.Println(Colorize(separator, uiCyan))
	terminal.Println()
}

func Colorize(text, color string) string {
	return color + text + uiReset
}

func sanitizeTerminalText(text string) string {
	if text == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(text))
	for _, char := range text {
		switch {
		case char == '\n' || char == '\r' || char == '\t':
			builder.WriteByte(' ')
		case unicode.IsControl(char):
			builder.WriteByte('?')
		default:
			builder.WriteRune(char)
		}
	}

	return builder.String()
}

func displayOutfitName(fileName string) string {
	return sanitizeTerminalText(strings.TrimSuffix(fileName, ".avatar"))
}

func Section(title, icon, color string) {
	SectionWithConsole(nil, title, icon, color)
}

func SectionWithConsole(console Console, title, icon, color string) {
	label := title
	if icon != "" {
		label = icon + " " + title
	}
	terminal := consoleOrDefault(console)
	terminal.Println(Colorize(label, uiBold+color))
	terminal.Println(Colorize(repeatLine("─", 40), color))
}

func KeyLabel(value string) string {
	return Colorize("[", uiCyan) + Colorize(value, uiBold+uiGreen) + Colorize("]", uiCyan)
}

func Dim(text string) string {
	return Colorize(text, uiYellow)
}

func Accent(text string) string {
	return Colorize(text, uiBold+uiCyan)
}

func repeatLine(char string, count int) string {
	line := ""
	for i := 0; i < count; i++ {
		line += char
	}
	return line
}

func Success(message string) {
	consoleOrDefault(nil).Success(message)
}

func Warning(message string) {
	consoleOrDefault(nil).Warning(message)
}
