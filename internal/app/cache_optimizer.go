package app

import (
	"sync"
	"time"
)

// CacheOptimizer provides caching for expensive operations
type CacheOptimizer struct {
	mu           sync.RWMutex
	fileCountCache map[string]cacheEntry
	ttl          time.Duration
}

type cacheEntry struct {
	value     int
	timestamp time.Time
}

// NewCacheOptimizer creates a new cache optimizer
func NewCacheOptimizer(ttl time.Duration) *CacheOptimizer {
	return &CacheOptimizer{
		fileCountCache: make(map[string]cacheEntry),
		ttl:           ttl,
	}
}

// GetFileCount returns cached file count or computes it
func (c *CacheOptimizer) GetFileCount(categoryPath string) (int, error) {
	c.mu.RLock()
	if entry, exists := c.fileCountCache[categoryPath]; exists {
		if time.Since(entry.timestamp) < c.ttl {
			c.mu.RUnlock()
			return entry.value, nil
		}
	}
	c.mu.RUnlock()
	
	// Compute and cache
	count, err := categoryFileCount(categoryPath)
	if err != nil {
		return 0, err
	}
	
	c.mu.Lock()
	c.fileCountCache[categoryPath] = cacheEntry{
		value:     count,
		timestamp: time.Now(),
	}
	c.mu.Unlock()
	
	return count, nil
}

// Clear removes cached entries
func (c *CacheOptimizer) Clear() {
	c.mu.Lock()
	c.fileCountCache = make(map[string]cacheEntry)
	c.mu.Unlock()
}