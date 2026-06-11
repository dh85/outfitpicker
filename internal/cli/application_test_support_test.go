package cli

import (
	"github.com/dh85/outfitpicker/internal/application/usecases"
	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
)

func newTestApplication(
	config *entities.Config,
	configManager usecases.ConfigManager,
	cacheManager usecases.CacheManager,
	categorySvc interfaces.CategoryService,
) *Application {
	return buildApplication(config, RuntimeDependencies{
		ConfigManager: configManager,
		CacheManager:  cacheManager,
		CategorySvc:   categorySvc,
	})
}
