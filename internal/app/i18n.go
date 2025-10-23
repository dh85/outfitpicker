package app

import "fmt"

// I18n handles internationalization
type I18n struct {
	locale   string
	messages map[string]map[string]string
}

func NewI18n(locale string) *I18n {
	i18n := &I18n{
		locale: locale,
		messages: map[string]map[string]string{
			"en": {
				"category":        "Category",
				"total_files":     "Total files in %q: %d",
				"selected":        "Selected: %d of %d",
				"options":         "Options:",
				"random_select":   "[r] Select a random file in this category",
				"show_selected":   "[s] Show previously selected files in this category",
				"show_unselected": "[u] Show unselected files in this category",
				"quit":            "[q] Quit",
				"enter_choice":    "Enter your choice: ",
				"kept_cached":     "kept and cached: %s",
				"exiting":         "Exiting.",
			},
			"es": {
				"category":        "Categoría",
				"total_files":     "Total de archivos en %q: %d",
				"selected":        "Seleccionados: %d de %d",
				"options":         "Opciones:",
				"random_select":   "[r] Seleccionar un archivo aleatorio en esta categoría",
				"show_selected":   "[s] Mostrar archivos seleccionados previamente",
				"show_unselected": "[u] Mostrar archivos no seleccionados",
				"quit":            "[q] Salir",
				"enter_choice":    "Ingrese su elección: ",
				"kept_cached":     "guardado y almacenado: %s",
				"exiting":         "Saliendo.",
			},
		},
	}

	// Default to English if locale not found
	if _, exists := i18n.messages[locale]; !exists {
		i18n.locale = "en"
	}

	return i18n
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
