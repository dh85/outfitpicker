package persistence

import (
	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
)

// FileServiceInterface defines the operations needed from FileService.
type FileServiceInterface[T any] interface {
	Load() (*T, error)
	Save(obj T) error
	Delete() error
}

// ConfigRepository implements configuration persistence using FileService.
type ConfigRepository struct {
	fileService FileServiceInterface[entities.Config]
}

// NewConfigRepository creates a new configuration repository.
func NewConfigRepository(fileService FileServiceInterface[entities.Config]) *ConfigRepository {
	return &ConfigRepository{
		fileService: fileService,
	}
}

// Load retrieves the configuration from storage.
func (r *ConfigRepository) Load() (*entities.Config, error) {
	return r.fileService.Load()
}

// Save persists the configuration to storage.
func (r *ConfigRepository) Save(config *entities.Config) error {
	return r.fileService.Save(*config)
}

// Delete removes the configuration from storage.
func (r *ConfigRepository) Delete() error {
	return r.fileService.Delete()
}

// Ensure ConfigRepository implements the interface
var _ interfaces.ConfigRepository = (*ConfigRepository)(nil)
