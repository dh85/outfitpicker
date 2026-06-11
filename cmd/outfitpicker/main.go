package main

import (
	"os"

	"github.com/dh85/outfitpicker/internal/application/usecases"
	"github.com/dh85/outfitpicker/internal/cli"
	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/infrastructure/persistence"
	infraServices "github.com/dh85/outfitpicker/internal/infrastructure/services"
	"github.com/dh85/outfitpicker/internal/infrastructure/system"
)

var bootstrapApplication = func(console cli.Console) (*cli.Application, bool) {
	deps := newRuntimeDependencies()
	return cli.BootstrapApplication(deps, console)
}

var showMainMenu = func(app *cli.Application, console cli.Console) {
	outfitService := cli.NewOutfitServiceFromRuntime(app)
	presentation := cli.NewOutfitPresentation(app, console)
	renderer := cli.NewMenuRenderer(console)
	cli.NewMenuSystem(outfitService, app, presentation, renderer, console).ShowMainMenu()
}

func main() {
	console := cli.NewTerminalConsole()
	app, ok := bootstrapApplication(console)
	if !ok {
		return
	}

	showMainMenu(app, console)
}

func newRuntimeDependencies() cli.RuntimeDependencies {
	configFileService := system.NewFileService[entities.Config](cliConfigFileName())
	cacheFileService := system.NewFileService[entities.OutfitCache](cliCacheFileName())
	configRepo := persistence.NewConfigRepository(configFileService)
	cacheRepo := persistence.NewCacheRepository(cacheFileService)

	return cli.RuntimeDependencies{
		ConfigManager: usecases.NewConfigUseCase(configRepo),
		CacheManager:  usecases.NewCacheUseCase(cacheRepo),
		CategorySvc:   infraServices.NewCategoryScanner(system.NewDefaultFileManager()),
		ConfigExists: func() bool {
			path, err := configFileService.FilePath()
			if err != nil {
				return false
			}
			_, err = os.Stat(path)
			return err == nil
		},
	}
}

func cliConfigFileName() string { return "config.json" }

func cliCacheFileName() string { return "cache.json" }
