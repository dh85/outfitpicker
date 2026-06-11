package cli

import (
	"errors"

	"github.com/dh85/outfitpicker/internal/application/usecases"
	"github.com/dh85/outfitpicker/internal/domain/entities"
	domainerrors "github.com/dh85/outfitpicker/internal/domain/errors"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
)

type SessionConfigController struct {
	current       *entities.Config
	configManager usecases.ConfigManager
	cacheManager  usecases.CacheManager
	session       *OutfitSession
}

func NewSessionConfigController(current *entities.Config, configManager usecases.ConfigManager, cacheManager usecases.CacheManager, session *OutfitSession) *SessionConfigController {
	return &SessionConfigController{current: current, configManager: configManager, cacheManager: cacheManager, session: session}
}

func (c *SessionConfigController) GetConfiguration() (*entities.Config, error) {
	config, err := c.configManager.LoadOrCreate()
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, domainerrors.ErrConfigurationNotFound
	}
	c.current = config
	return config, nil
}

func (c *SessionConfigController) UpdateConfiguration(config *entities.Config) error {
	currentRoot := ""
	if c.current != nil {
		currentRoot = c.current.Root
	}
	rootChanged := currentRoot != "" && currentRoot != config.Root

	if err := c.configManager.Save(config); err != nil {
		return err
	}
	if rootChanged {
		if err := c.cacheManager.Delete(); err != nil {
			return err
		}
		c.session.ResetAll()
	}
	c.current = config
	return nil
}

type SessionCommandHandler struct {
	categorySvc   interfaces.CategoryService
	configManager usecases.ConfigManager
	cacheManager  usecases.CacheManager
	session       *OutfitSession
}

func NewSessionCommandHandler(categorySvc interfaces.CategoryService, configManager usecases.ConfigManager, cacheManager usecases.CacheManager, session *OutfitSession) *SessionCommandHandler {
	return &SessionCommandHandler{categorySvc: categorySvc, configManager: configManager, cacheManager: cacheManager, session: session}
}

func (h *SessionCommandHandler) WearOutfit(outfit entities.OutfitReference) error {
	err := usecases.NewWearOutfitUseCase(h.categorySvc, h.configManager, h.cacheManager).Execute(outfit)
	if err == nil {
		h.session.ResetAll()
		return nil
	}

	var rotationCompleted *domainerrors.RotationCompletedError
	if errors.As(err, &rotationCompleted) {
		h.session.ResetAll()
	}
	return err
}

func (h *SessionCommandHandler) ResetCategory(categoryName string) error {
	if err := usecases.NewResetCategoryUseCase(h.configManager, h.cacheManager).Execute(categoryName); err != nil {
		return err
	}
	h.session.ResetCategory(categoryName)
	return nil
}

func (h *SessionCommandHandler) ResetAllCategories() error {
	if err := usecases.NewResetCategoryUseCase(h.configManager, h.cacheManager).ExecuteAll(); err != nil {
		return err
	}
	h.session.ResetAll()
	return nil
}

func (h *SessionCommandHandler) FactoryReset() error {
	if err := h.configManager.Delete(); err != nil {
		return err
	}
	if err := h.cacheManager.Delete(); err != nil {
		return err
	}
	h.session.ResetAll()
	return nil
}
