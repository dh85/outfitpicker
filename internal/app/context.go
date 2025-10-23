package app

import (
	"context"
	"io"
	"time"

	"github.com/dh85/outfitpicker/internal/storage"
)

// AppContext holds application context and dependencies
type AppContext struct {
	ctx       context.Context
	config    AppConfig
	logger    *Logger
	cache     *storage.Manager
	stdout    io.Writer
	optimizer *CacheOptimizer
}

// NewAppContext creates a new application context
func NewAppContext(ctx context.Context, cache *storage.Manager, stdout io.Writer) *AppContext {
	return &AppContext{
		ctx:       ctx,
		config:    DefaultAppConfig(),
		logger:    DefaultLogger(),
		cache:     cache,
		stdout:    stdout,
		optimizer: NewCacheOptimizer(time.Minute * 5),
	}
}

// WithConfig sets the configuration
func (ac *AppContext) WithConfig(config AppConfig) *AppContext {
	ac.config = config
	return ac
}

// WithLogger sets the logger
func (ac *AppContext) WithLogger(logger *Logger) *AppContext {
	ac.logger = logger
	return ac
}
