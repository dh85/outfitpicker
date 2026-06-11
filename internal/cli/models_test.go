package cli

import "testing"

func TestMenuChoiceDescriptions(t *testing.T) {
	if got := MenuChoiceRandom.Description(); got != "Pick a random outfit" {
		t.Fatalf("MenuChoiceRandom.Description() = %q", got)
	}
	if got := MenuChoiceManual.Description(); got != "Choose an outfit manually" {
		t.Fatalf("MenuChoiceManual.Description() = %q", got)
	}
	if got := MenuChoiceWorn.Description(); got != "Show outfits already worn" {
		t.Fatalf("MenuChoiceWorn.Description() = %q", got)
	}
	if got := MenuChoiceUnworn.Description(); got != "Show outfits not yet worn" {
		t.Fatalf("MenuChoiceUnworn.Description() = %q", got)
	}
	if got := MenuChoiceAdvanced.Description(); got != "Advanced settings" {
		t.Fatalf("MenuChoiceAdvanced.Description() = %q", got)
	}
	if got := MenuChoiceQuit.Description(); got != "Quit" {
		t.Fatalf("MenuChoiceQuit.Description() = %q", got)
	}
	if got := MenuChoice("invalid").Description(); got != "" {
		t.Fatalf("MenuChoice(invalid).Description() = %q, want empty string", got)
	}
}

func TestAdvancedChoiceDescriptions(t *testing.T) {
	if got := AdvancedChoiceChangePath.Description(); got != "Change outfit path" {
		t.Fatalf("AdvancedChoiceChangePath.Description() = %q", got)
	}
	if got := AdvancedChoiceChangeLanguage.Description(); got != "Change language" {
		t.Fatalf("AdvancedChoiceChangeLanguage.Description() = %q", got)
	}
	if got := AdvancedChoiceChangeExcluded.Description(); got != "Manage categories excluded from random selection" {
		t.Fatalf("AdvancedChoiceChangeExcluded.Description() = %q", got)
	}
	if got := AdvancedChoiceResetCategory.Description(); got != "Reset worn outfits for category" {
		t.Fatalf("AdvancedChoiceResetCategory.Description() = %q", got)
	}
	if got := AdvancedChoiceResetAll.Description(); got != "Reset all worn outfits" {
		t.Fatalf("AdvancedChoiceResetAll.Description() = %q", got)
	}
	if got := AdvancedChoiceResetSettings.Description(); got != "Reset user settings and worn outfits" {
		t.Fatalf("AdvancedChoiceResetSettings.Description() = %q", got)
	}
	if got := AdvancedChoiceBack.Description(); got != "Back to main menu" {
		t.Fatalf("AdvancedChoiceBack.Description() = %q", got)
	}
	if got := AdvancedChoiceQuit.Description(); got != "Quit" {
		t.Fatalf("AdvancedChoiceQuit.Description() = %q", got)
	}
	if got := AdvancedChoice("invalid").Description(); got != "" {
		t.Fatalf("AdvancedChoice(invalid).Description() = %q, want empty string", got)
	}
}

func TestChoiceRawValuesAndParsers(t *testing.T) {
	if MenuChoiceRandom != "r" || MenuChoiceWorn != "w" || MenuChoiceUnworn != "u" || MenuChoiceAdvanced != "a" || MenuChoiceQuit != "q" {
		t.Fatal("menu choice raw values changed unexpectedly")
	}
	if AdvancedChoiceChangePath != "p" || AdvancedChoiceChangeLanguage != "l" || AdvancedChoiceChangeExcluded != "e" || AdvancedChoiceResetCategory != "c" || AdvancedChoiceResetAll != "r" || AdvancedChoiceResetSettings != "s" || AdvancedChoiceBack != "b" || AdvancedChoiceQuit != "q" {
		t.Fatal("advanced choice raw values changed unexpectedly")
	}

	if got, ok := ParseMenuChoice("r"); !ok || got != MenuChoiceRandom {
		t.Fatal("ParseMenuChoice(r) failed")
	}
	if _, ok := ParseMenuChoice("invalid"); ok {
		t.Fatal("ParseMenuChoice(invalid) should fail")
	}
	if got, ok := ParseAdvancedChoice("p"); !ok || got != AdvancedChoiceChangePath {
		t.Fatal("ParseAdvancedChoice(p) failed")
	}
	if _, ok := ParseAdvancedChoice("invalid"); ok {
		t.Fatal("ParseAdvancedChoice(invalid) should fail")
	}
}

func TestAllChoices(t *testing.T) {
	if got := len(AllMenuChoices()); got != 6 {
		t.Fatalf("len(AllMenuChoices()) = %d, want 6", got)
	}
	if got := len(AllAdvancedChoices()); got != 8 {
		t.Fatalf("len(AllAdvancedChoices()) = %d, want 8", got)
	}
}

func TestOutfitChoiceValues(t *testing.T) {
	choices := []OutfitChoice{OutfitChoiceWorn, OutfitChoiceSkipped, OutfitChoiceBack, OutfitChoiceQuit}
	if len(choices) != 4 {
		t.Fatalf("len(choices) = %d, want 4", len(choices))
	}
	if OutfitChoiceWorn == OutfitChoiceSkipped {
		t.Fatal("OutfitChoiceWorn should not equal OutfitChoiceSkipped")
	}
	if OutfitChoiceBack == OutfitChoiceQuit {
		t.Fatal("OutfitChoiceBack should not equal OutfitChoiceQuit")
	}
}
