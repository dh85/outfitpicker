package usecases

import (
	"github.com/dh85/outfitpicker/internal/domain/entities"
	"github.com/dh85/outfitpicker/internal/domain/interfaces"
)

type CacheUseCase struct {
	repo interfaces.CacheRepository
}

func NewCacheUseCase(repo interfaces.CacheRepository) *CacheUseCase {
	return &CacheUseCase{repo: repo}
}

func (uc *CacheUseCase) LoadOrCreate() (*entities.OutfitCache, error) {
	cache, err := uc.repo.Load()
	if err != nil {
		return nil, err
	}
	if cache == nil {
		cache := entities.NewOutfitCache()
		return &cache, nil
	}
	return cache, nil
}

func (uc *CacheUseCase) Save(cache *entities.OutfitCache) error {
	return uc.repo.Save(cache)
}

func (uc *CacheUseCase) Delete() error {
	return uc.repo.Delete()
}
