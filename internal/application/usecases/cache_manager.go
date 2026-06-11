package usecases

import "github.com/dh85/outfitpicker/internal/domain/entities"

type CacheManager interface {
	LoadOrCreate() (*entities.OutfitCache, error)
	Save(cache *entities.OutfitCache) error
	Delete() error
}
