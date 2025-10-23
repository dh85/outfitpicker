package metrics

import (
	"sync"
	"time"
)

type Metrics struct {
	mu sync.RWMutex
	counters map[string]int64
	timers   map[string]time.Duration
}

func New() *Metrics {
	return &Metrics{
		counters: make(map[string]int64),
		timers:   make(map[string]time.Duration),
	}
}

func (m *Metrics) Inc(name string) {
	m.mu.Lock()
	m.counters[name]++
	m.mu.Unlock()
}

func (m *Metrics) Time(name string, duration time.Duration) {
	m.mu.Lock()
	m.timers[name] = duration
	m.mu.Unlock()
}

func (m *Metrics) Get(name string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.counters[name]
}

func (m *Metrics) GetTime(name string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.timers[name]
}