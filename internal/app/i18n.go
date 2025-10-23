package app

import (
	"fmt"
	"os"
	"strings"
)

// I18n handles internationalization
type I18n struct {
	locale   string
	messages map[string]map[string]string
}

func NewI18n(locale string) *I18n {
	i18n := &I18n{
		locale:   locale,
		messages: getTranslations(),
	}

	// Default to English if locale not found
	if _, exists := i18n.messages[locale]; !exists {
		i18n.locale = "en"
	}

	return i18n
}

// DetectLocale detects the user's locale from environment variables
func DetectLocale() string {
	// Check LANG environment variable first
	if lang := os.Getenv("LANG"); lang != "" {
		return parseLocale(lang)
	}

	// Check LC_ALL and LC_MESSAGES as fallbacks
	for _, env := range []string{"LC_ALL", "LC_MESSAGES"} {
		if lang := os.Getenv(env); lang != "" {
			return parseLocale(lang)
		}
	}

	// Default to English
	return "en"
}

func parseLocale(lang string) string {
	// Handle common CI/system locales that should default to English
	lang = strings.ToLower(lang)
	if lang == "c" || lang == "c.utf-8" || lang == "posix" {
		return "en"
	}

	// Extract language code (e.g., "es_ES.UTF-8" -> "es")
	if parts := strings.Split(lang, "_"); len(parts) > 0 {
		langCode := strings.Split(parts[0], ".")[0] // Handle "en.UTF-8" format
		return langCode
	}

	return "en"
}

func (i *I18n) T(key string, args ...interface{}) string {
	if msg, exists := i.messages[i.locale][key]; exists {
		if len(args) > 0 {
			return fmt.Sprintf(msg, args...)
		}
		return msg
	}

	// Fallback to English
	if msg, exists := i.messages["en"][key]; exists {
		if len(args) > 0 {
			return fmt.Sprintf(msg, args...)
		}
		return msg
	}

	return key // Return key if translation not found
}

// GetLocale returns the current locale
func (i *I18n) GetLocale() string {
	return i.locale
}

// SetLocale changes the current locale
func (i *I18n) SetLocale(locale string) {
	if _, exists := i.messages[locale]; exists {
		i.locale = locale
	} else {
		i.locale = "en" // fallback to English
	}
}
