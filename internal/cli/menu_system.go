package cli

type MenuSystem struct {
	outfitService OutfitService
	selector      RandomOutfitSelector
	presentation  OutfitPresentation
	renderer      MenuRenderer
	console       Console
}

func NewMenuSystem(outfitService OutfitService, selector RandomOutfitSelector, presentation OutfitPresentation, renderer MenuRenderer, consoles ...Console) MenuSystem {
	return MenuSystem{
		outfitService: outfitService,
		selector:      selector,
		presentation:  presentation,
		renderer:      renderer,
		console:       optionalConsole(consoles),
	}
}

func (m MenuSystem) ShowMainMenu() {
	transition := mainMenuTransition()

	for {
		nextTransition, ok := m.dispatchTransition(transition)
		if !ok {
			return
		}
		transition = nextTransition
	}
}

func (m MenuSystem) dispatchTransition(transition menuTransition) (menuTransition, bool) {
	switch transition.destination {
	case menuDestinationMain:
		return MainMenu{outfitService: m.outfitService, selector: m.selector, presentation: m.presentation, renderer: m.renderer, console: m.console}.Show(), true
	case menuDestinationCategory:
		return CategoryMenu{outfitService: m.outfitService, selector: m.selector, presentation: m.presentation, renderer: m.renderer, category: transition.category, console: m.console}.Show(), true
	case menuDestinationAdvanced:
		return AdvancedMenu{outfitService: m.outfitService, console: m.console}.Show(), true
	case menuDestinationExit:
		return menuTransition{}, false
	default:
		return menuTransition{}, false
	}
}
