package main

import (
	"fmt"
	"io"
	"os"

	"github.com/dh85/outfitpicker/internal/application/usecases"
	"github.com/dh85/outfitpicker/internal/cli"
	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/infrastructure/persistence"
	infraServices "github.com/dh85/outfitpicker/internal/infrastructure/services"
	"github.com/dh85/outfitpicker/internal/infrastructure/system"
)

var version = "dev"

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

var executeCommand = cli.ExecuteCommand

var exitProcess = os.Exit

func main() {
	if printVersion(os.Args[1:], os.Stdout) {
		return
	}

	console := cli.NewTerminalConsole()
	if len(os.Args) > 1 {
		if handled, code := executeCommand(os.Args[1:], nil, console); handled {
			if code != 0 {
				exitProcess(code)
			}
			return
		}
	}

	app, ok := bootstrapApplication(console)
	if !ok {
		return
	}

	if handled, code := executeCommand(os.Args[1:], app, console); handled {
		if code != 0 {
			exitProcess(code)
		}
		return
	}

	showMainMenu(app, console)
}

func printVersion(args []string, output io.Writer) bool {
	if len(args) != 1 {
		return false
	}
	switch args[0] {
	case "version", "--version", "-v":
		fmt.Fprintf(output, "outfitpicker %s\n", version)
		return true
	default:
		return false
	}
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
		PathProvider: cli.FuncStoragePathProvider{
			ConfigPathFunc: configFileService.FilePath,
			CachePathFunc:  cacheFileService.FilePath,
		},
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
