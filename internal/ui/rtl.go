// Package ui provides RTL (Right-to-Left) language support utilities.
package ui

import "unicode"

// RTL language codes that require right-to-left text direction
var rtlLanguages = map[string]bool{
	"ar": true, // Arabic
	"he": true, // Hebrew
	"ur": true, // Urdu
	"fa": true, // Persian/Farsi
}

// IsRTL checks if a language code requires RTL text direction
func IsRTL(langCode string) bool {
	return rtlLanguages[langCode]
}

// ReverseString reverses a string for RTL display
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// FormatRTL formats text for RTL display if needed
func FormatRTL(text string, isRTL bool) string {
	if !isRTL {
		return text
	}

	// Add RTL mark to ensure proper text direction
	const rtlMark = "\u200F"
	return rtlMark + text
}

// AlignText aligns text based on RTL requirements
func AlignText(text string, isRTL bool, width int) string {
	if !isRTL || len(text) >= width {
		return text
	}

	// Right-align for RTL languages
	padding := width - len([]rune(text))
	return string(make([]rune, padding)) + text
}

// ContainsRTLChars checks if text contains RTL characters
func ContainsRTLChars(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Arabic, r) || unicode.Is(unicode.Hebrew, r) {
			return true
		}
	}
	return false
}
