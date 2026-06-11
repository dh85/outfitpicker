package cli

type MenuChoice string

const (
	MenuChoiceRandom   MenuChoice = "r"
	MenuChoiceManual   MenuChoice = "m"
	MenuChoiceWorn     MenuChoice = "w"
	MenuChoiceUnworn   MenuChoice = "u"
	MenuChoiceAdvanced MenuChoice = "a"
	MenuChoiceQuit     MenuChoice = "q"
)

func AllMenuChoices() []MenuChoice {
	return []MenuChoice{
		MenuChoiceRandom,
		MenuChoiceManual,
		MenuChoiceWorn,
		MenuChoiceUnworn,
		MenuChoiceAdvanced,
		MenuChoiceQuit,
	}
}

func ParseMenuChoice(value string) (MenuChoice, bool) {
	choice := MenuChoice(value)
	switch choice {
	case MenuChoiceRandom, MenuChoiceManual, MenuChoiceWorn, MenuChoiceUnworn, MenuChoiceAdvanced, MenuChoiceQuit:
		return choice, true
	default:
		return "", false
	}
}

func (m MenuChoice) Description() string {
	switch m {
	case MenuChoiceRandom:
		return "Pick a random outfit"
	case MenuChoiceManual:
		return "Choose an outfit manually"
	case MenuChoiceWorn:
		return "Show outfits already worn"
	case MenuChoiceUnworn:
		return "Show outfits not yet worn"
	case MenuChoiceAdvanced:
		return "Advanced settings"
	case MenuChoiceQuit:
		return "Quit"
	default:
		return ""
	}
}

type AdvancedChoice string

const (
	AdvancedChoiceChangePath     AdvancedChoice = "p"
	AdvancedChoiceChangeLanguage AdvancedChoice = "l"
	AdvancedChoiceChangeExcluded AdvancedChoice = "e"
	AdvancedChoiceResetCategory  AdvancedChoice = "c"
	AdvancedChoiceResetAll       AdvancedChoice = "r"
	AdvancedChoiceResetSettings  AdvancedChoice = "s"
	AdvancedChoiceBack           AdvancedChoice = "b"
	AdvancedChoiceQuit           AdvancedChoice = "q"
)

func AllAdvancedChoices() []AdvancedChoice {
	return []AdvancedChoice{
		AdvancedChoiceChangePath,
		AdvancedChoiceChangeLanguage,
		AdvancedChoiceChangeExcluded,
		AdvancedChoiceResetCategory,
		AdvancedChoiceResetAll,
		AdvancedChoiceResetSettings,
		AdvancedChoiceBack,
		AdvancedChoiceQuit,
	}
}

func ParseAdvancedChoice(value string) (AdvancedChoice, bool) {
	choice := AdvancedChoice(value)
	switch choice {
	case AdvancedChoiceChangePath, AdvancedChoiceChangeLanguage, AdvancedChoiceChangeExcluded, AdvancedChoiceResetCategory, AdvancedChoiceResetAll, AdvancedChoiceResetSettings, AdvancedChoiceBack, AdvancedChoiceQuit:
		return choice, true
	default:
		return "", false
	}
}

func (a AdvancedChoice) Description() string {
	switch a {
	case AdvancedChoiceChangePath:
		return "Change outfit path"
	case AdvancedChoiceChangeLanguage:
		return "Change language"
	case AdvancedChoiceChangeExcluded:
		return "Manage categories excluded from random selection"
	case AdvancedChoiceResetCategory:
		return "Reset worn outfits for category"
	case AdvancedChoiceResetAll:
		return "Reset all worn outfits"
	case AdvancedChoiceResetSettings:
		return "Reset user settings and worn outfits"
	case AdvancedChoiceBack:
		return "Back to main menu"
	case AdvancedChoiceQuit:
		return "Quit"
	default:
		return ""
	}
}

type OutfitChoice int

const (
	OutfitChoiceWorn OutfitChoice = iota
	OutfitChoiceSkipped
	OutfitChoiceBack
	OutfitChoiceQuit
)
