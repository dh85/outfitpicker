package usecases

import (
	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
)

type ConfigUseCase struct {
	repo interfaces.ConfigRepository
}

func NewConfigUseCase(repo interfaces.ConfigRepository) *ConfigUseCase {
	return &ConfigUseCase{repo}
}

func (uc *ConfigUseCase) LoadOrCreate() (*entities.Config, error) {
	return uc.repo.Load()
}

func (uc *ConfigUseCase) Save(config *entities.Config) error {
	return uc.repo.Save(config)
}

func (uc *ConfigUseCase) Delete() error {
	return uc.repo.Delete()
}
