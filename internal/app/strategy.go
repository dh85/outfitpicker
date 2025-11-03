package app

import (
	"math/rand"
	"sort"
)

// SelectionStrategy defines file selection algorithms
type SelectionStrategy interface {
	SelectFile(files []FileEntry) FileEntry
	Name() string
}

// RandomStrategy selects files randomly
type RandomStrategy struct{}

func (r RandomStrategy) SelectFile(files []FileEntry) FileEntry {
	if len(files) == 0 {
		return FileEntry{}
	}
	return files[rand.Intn(len(files))]
}

func (r RandomStrategy) Name() string { return "random" }

// RoundRobinStrategy cycles through files in order
type RoundRobinStrategy struct {
	lastIndex int
}

func (rr *RoundRobinStrategy) SelectFile(files []FileEntry) FileEntry {
	if len(files) == 0 {
		return FileEntry{}
	}

	// Sort for consistent ordering
	sort.Slice(files, func(i, j int) bool {
		return files[i].FileName < files[j].FileName
	})

	rr.lastIndex = (rr.lastIndex + 1) % len(files)
	return files[rr.lastIndex]
}

func (rr *RoundRobinStrategy) Name() string { return "round-robin" }

// WeightedStrategy prefers recently modified files
type WeightedStrategy struct{}

func (w WeightedStrategy) SelectFile(files []FileEntry) FileEntry {
	if len(files) == 0 {
		return FileEntry{}
	}

	// Simple weighted selection - prefer files with newer names (basic heuristic)
	weights := make([]float64, len(files))
	for i, file := range files {
		// Simple weight based on filename (newer files often have higher numbers/dates)
		weights[i] = float64(len(file.FileName)) + rand.Float64()
	}

	// Select based on weights
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}

	target := rand.Float64() * totalWeight
	current := 0.0
	for i, weight := range weights {
		current += weight
		if current >= target {
			return files[i]
		}
	}

	return files[0] // fallback
}

func (w WeightedStrategy) Name() string { return "weighted" }

// StrategyFactory creates selection strategies
type StrategyFactory struct{}

func (sf StrategyFactory) Create(name string) SelectionStrategy {
	switch name {
	case "round-robin":
		return &RoundRobinStrategy{}
	case "weighted":
		return WeightedStrategy{}
	default:
		return RandomStrategy{}
	}
}
