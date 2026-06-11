package main

import (
	"testing"

	"github.com/dh85/outfitpicker/internal/cli"
)

func TestMain(t *testing.T) {
	originalBootstrap := bootstrapApplication
	originalShowMainMenu := showMainMenu
	t.Cleanup(func() {
		bootstrapApplication = originalBootstrap
		showMainMenu = originalShowMainMenu
	})

	t.Run("returns when bootstrap fails", func(t *testing.T) {
		bootstrapCalls := 0
		showCalls := 0

		bootstrapApplication = func(cli.Console) (*cli.Application, bool) {
			bootstrapCalls++
			return nil, false
		}
		showMainMenu = func(*cli.Application, cli.Console) {
			showCalls++
		}

		main()

		if bootstrapCalls != 1 {
			t.Fatalf("bootstrapCalls = %d, want 1", bootstrapCalls)
		}
		if showCalls != 0 {
			t.Fatalf("showCalls = %d, want 0", showCalls)
		}
	})

	t.Run("shows menu when bootstrap succeeds", func(t *testing.T) {
		bootstrapCalls := 0
		showCalls := 0
		app := &cli.Application{}

		bootstrapApplication = func(cli.Console) (*cli.Application, bool) {
			bootstrapCalls++
			return app, true
		}
		showMainMenu = func(received *cli.Application, _ cli.Console) {
			showCalls++
			if received != app {
				t.Fatalf("showMainMenu() received %p, want %p", received, app)
			}
		}

		main()

		if bootstrapCalls != 1 {
			t.Fatalf("bootstrapCalls = %d, want 1", bootstrapCalls)
		}
		if showCalls != 1 {
			t.Fatalf("showCalls = %d, want 1", showCalls)
		}
	})
}
