package cli

import "github.com/dh85/outfitpicker/internal/domain/entities"

type menuDestination int

const (
	menuDestinationMain menuDestination = iota
	menuDestinationCategory
	menuDestinationAdvanced
	menuDestinationExit
)

type menuTransition struct {
	destination menuDestination
	category    entities.CategoryReference
}

func mainMenuTransition() menuTransition {
	return menuTransition{destination: menuDestinationMain}
}

func categoryMenuTransition(category entities.CategoryReference) menuTransition {
	return menuTransition{destination: menuDestinationCategory, category: category}
}

func advancedMenuTransition() menuTransition {
	return menuTransition{destination: menuDestinationAdvanced}
}

func exitMenuTransition() menuTransition {
	return menuTransition{destination: menuDestinationExit}
}
