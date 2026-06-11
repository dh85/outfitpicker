package cli

import (
	"math/rand"

	"github.com/dh85/outfitpicker/internal/application/usecases"
	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
)

type RuntimeDependencies struct {
	ConfigManager usecases.ConfigManager
	CacheManager  usecases.CacheManager
	CategorySvc   interfaces.CategoryService
	RandomInt     func(int) int
	ConfigExists  func() bool
}

type Application struct {
	wardrobe  WardrobeReader
	config    ConfigurationController
	commands  OutfitCommandHandler
	randomInt func(int) int
	selection RandomOutfitSelector
	session   *OutfitSession
}

func buildApplication(config *entities.Config, deps RuntimeDependencies) *Application {
	randomInt := deps.RandomInt
	if randomInt == nil {
		randomInt = rand.Intn
	}
	session := NewOutfitSession()
	wardrobe := usecases.NewWardrobeQueries(deps.ConfigManager, deps.CacheManager, deps.CategorySvc)
	configController := NewSessionConfigController(config, deps.ConfigManager, deps.CacheManager, session)
	commands := NewSessionCommandHandler(deps.CategorySvc, deps.ConfigManager, deps.CacheManager, session)
	app := &Application{
		wardrobe:  wardrobe,
		config:    configController,
		commands:  commands,
		randomInt: randomInt,
		session:   session,
	}
	app.selection = NewRuntimeSelectionService(
		deps.CategorySvc,
		deps.ConfigManager,
		deps.CacheManager,
		app.session,
		func(length int) int {
			if length <= 1 {
				return 0
			}
			return app.randomInt(length)
		},
	)
	return app
}
