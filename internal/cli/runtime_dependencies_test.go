package cli

import (
	"os"

	"github.com/dh85/outfitpicker/internal/application/usecases"
	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/infrastructure/persistence"
	infraServices "github.com/dh85/outfitpicker/internal/infrastructure/services"
	"github.com/dh85/outfitpicker/internal/infrastructure/system"
)

func newProductionStyleRuntimeDependencies() RuntimeDependencies {
	configFileService := system.NewFileService[entities.Config](configFileName)
	cacheFileService := system.NewFileService[entities.OutfitCache](cacheFileName)
	configRepo := persistence.NewConfigRepository(configFileService)
	cacheRepo := persistence.NewCacheRepository(cacheFileService)

	return RuntimeDependencies{
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

func newMenuSystemForRuntime(runtime interface {
	WardrobeReader
	ConfigurationController
	OutfitCommandHandler
	RandomOutfitSelector
}, consoles ...Console) MenuSystem {
	console := optionalConsole(consoles)
	outfitService := NewOutfitServiceFromRuntime(runtime)
	presentation := NewOutfitPresentation(runtime, console)
	renderer := NewMenuRenderer(console)
	return NewMenuSystem(outfitService, runtime, presentation, renderer, console)
}
