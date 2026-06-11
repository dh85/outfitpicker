package interfaces

import "github.com/dh85/outfitpicker/internal/domain/entities"

// ConfigRepository handles configuration persistence.
type ConfigRepository interface {
	Load() (*entities.Config, error)
	Save(config *entities.Config) error
	Delete() error
}

// CacheRepository handles outfit cache persistence.
type CacheRepository interface {
	Load() (*entities.OutfitCache, error)
	Save(cache *entities.OutfitCache) error
	Delete() error
}