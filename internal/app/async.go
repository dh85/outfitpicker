package app

import (
	"context"
	"sync"
)

// AsyncOperations handles non-blocking operations
type AsyncOperations struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewAsyncOperations(ctx context.Context) *AsyncOperations {
	ctx, cancel := context.WithCancel(ctx)
	return &AsyncOperations{
		ctx:    ctx,
		cancel: cancel,
	}
}

// LoadCategoriesAsync loads categories in background
func (ao *AsyncOperations) LoadCategoriesAsync(rootPath string, callback func([]string, error)) {
	ao.wg.Add(1)
	go func() {
		defer ao.wg.Done()

		select {
		case <-ao.ctx.Done():
			callback(nil, ao.ctx.Err())
			return
		default:
		}

		categories, err := listCategories(rootPath)
		callback(categories, err)
	}()
}

// PreloadCacheAsync preloads cache data
func (ao *AsyncOperations) PreloadCacheAsync(categories []string, optimizer *CacheOptimizer, callback func(error)) {
	ao.wg.Add(1)
	go func() {
		defer ao.wg.Done()

		for _, cat := range categories {
			select {
			case <-ao.ctx.Done():
				callback(ao.ctx.Err())
				return
			default:
			}

			_, err := optimizer.GetFileCount(cat)
			if err != nil {
				callback(err)
				return
			}
		}
		callback(nil)
	}()
}

func (ao *AsyncOperations) Wait() {
	ao.wg.Wait()
}

func (ao *AsyncOperations) Cancel() {
	ao.cancel()
}
