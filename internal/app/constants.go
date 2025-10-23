package app

const (
	// File filtering
	HiddenFilePrefix = "."
	ExcludedDir      = "Downloads"
	
	// User actions
	ActionKeep = "k"
	ActionSkip = "s"
	ActionQuit = "q"
	ActionRandom = "r"
	ActionShowSelected = "s"
	ActionShowUnselected = "u"
	
	// Display messages
	ExitMessage = "Exiting."
	InvalidActionMessage = "invalid action. please try again."
	CacheClearedMessage = "cache cleared for %q — next random will restart the cycle"
)