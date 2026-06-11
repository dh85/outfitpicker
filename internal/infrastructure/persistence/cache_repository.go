package persistence

import (
	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
)

// CacheRepository implements outfit cache persistence using FileService.
type CacheRepository struct {
	fileService FileServiceInterface[entities.OutfitCache]
}

// NewCacheRepository creates a new cache repository.
func NewCacheRepository(fileService FileServiceInterface[entities.OutfitCache]) *CacheRepository {
	return &CacheRepository{
		fileService: fileService,
	}
}

// Load retrieves the outfit cache from storage.
func (r *CacheRepository) Load() (*entities.OutfitCache, error) {
	return r.fileService.Load()
}

// Save persists the outfit cache to storage.
func (r *CacheRepository) Save(cache *entities.OutfitCache) error {
	return r.fileService.Save(*cache)
}

// Delete removes the outfit cache from storage.
func (r *CacheRepository) Delete() error {
	return r.fileService.Delete()
}

// Ensure CacheRepository implements the interface
var _ interfaces.CacheRepository = (*CacheRepository)(nil)
