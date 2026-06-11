package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/dh85/outfitpicker/internal/cli"
)

func TestMain(t *testing.T) {
	originalBootstrap := bootstrapApplication
	originalShowMainMenu := showMainMenu
	originalExecuteCommand := executeCommand
	originalExitProcess := exitProcess
	originalArgs := os.Args
	t.Cleanup(func() {
		bootstrapApplication = originalBootstrap
		showMainMenu = originalShowMainMenu
		executeCommand = originalExecuteCommand
		exitProcess = originalExitProcess
		os.Args = originalArgs
	})

	t.Run("shows help before bootstrap", func(t *testing.T) {
		os.Args = []string{"outfitpicker", "--help"}
		bootstrapCalls := 0
		showCalls := 0
		executeCalls := 0

		bootstrapApplication = func(cli.Console) (*cli.Application, bool) {
			bootstrapCalls++
			return nil, false
		}
		showMainMenu = func(*cli.Application, cli.Console) {
			showCalls++
		}
		executeCommand = func(args []string, received cli.CommandRuntime, _ cli.Console) (bool, int) {
			executeCalls++
			if received != nil {
				t.Fatalf("executeCommand runtime = %v, want nil", received)
			}
			if len(args) != 1 || args[0] != "--help" {
				t.Fatalf("executeCommand args = %#v, want --help", args)
			}
			return true, 0
		}

		main()

		if executeCalls != 1 {
			t.Fatalf("executeCalls = %d, want 1", executeCalls)
		}
		if bootstrapCalls != 0 {
			t.Fatalf("bootstrapCalls = %d, want 0", bootstrapCalls)
		}
		if showCalls != 0 {
			t.Fatalf("showCalls = %d, want 0", showCalls)
		}
	})

	t.Run("returns when bootstrap fails", func(t *testing.T) {
		os.Args = []string{"outfitpicker"}
		bootstrapCalls := 0
		showCalls := 0

		bootstrapApplication = func(cli.Console) (*cli.Application, bool) {
			bootstrapCalls++
			return nil, false
		}
		showMainMenu = func(*cli.Application, cli.Console) {
			showCalls++
		}
		executeCommand = func([]string, cli.CommandRuntime, cli.Console) (bool, int) {
			return false, 0
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
		os.Args = []string{"outfitpicker"}
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
		executeCommand = func(args []string, _ cli.CommandRuntime, _ cli.Console) (bool, int) {
			if len(args) != 0 {
				t.Fatalf("executeCommand args = %#v, want empty", args)
			}
			return false, 0
		}

		main()

		if bootstrapCalls != 1 {
			t.Fatalf("bootstrapCalls = %d, want 1", bootstrapCalls)
		}
		if showCalls != 1 {
			t.Fatalf("showCalls = %d, want 1", showCalls)
		}
	})

	t.Run("runs command and skips menu when command is handled", func(t *testing.T) {
		os.Args = []string{"outfitpicker", "list", "categories"}
		app := &cli.Application{}
		showCalls := 0
		exitCalls := 0

		bootstrapApplication = func(cli.Console) (*cli.Application, bool) {
			return app, true
		}
		showMainMenu = func(*cli.Application, cli.Console) {
			showCalls++
		}
		executeCalls := 0
		executeCommand = func(args []string, received cli.CommandRuntime, _ cli.Console) (bool, int) {
			executeCalls++
			if len(args) != 2 || args[0] != "list" || args[1] != "categories" {
				t.Fatalf("executeCommand args = %#v, want list categories", args)
			}
			if executeCalls == 1 {
				if received != nil {
					t.Fatalf("first executeCommand runtime = %v, want nil", received)
				}
				return false, 0
			}
			if received != app {
				t.Fatalf("second executeCommand runtime = %p, want %p", received, app)
			}
			return true, 0
		}
		exitProcess = func(int) {
			exitCalls++
		}

		main()

		if executeCalls != 2 {
			t.Fatalf("executeCalls = %d, want 2", executeCalls)
		}
		if showCalls != 0 {
			t.Fatalf("showCalls = %d, want 0", showCalls)
		}
		if exitCalls != 0 {
			t.Fatalf("exitCalls = %d, want 0", exitCalls)
		}
	})

	t.Run("exits with command status when handled command fails", func(t *testing.T) {
		os.Args = []string{"outfitpicker", "reset"}
		app := &cli.Application{}
		gotExitCode := -1

		bootstrapApplication = func(cli.Console) (*cli.Application, bool) {
			return app, true
		}
		showMainMenu = func(*cli.Application, cli.Console) {
			t.Fatal("showMainMenu should not be called")
		}
		executeCalls := 0
		executeCommand = func([]string, cli.CommandRuntime, cli.Console) (bool, int) {
			executeCalls++
			if executeCalls == 1 {
				return false, 0
			}
			return true, 2
		}
		exitProcess = func(code int) {
			gotExitCode = code
		}

		main()

		if executeCalls != 2 {
			t.Fatalf("executeCalls = %d, want 2", executeCalls)
		}
		if gotExitCode != 2 {
			t.Fatalf("exit code = %d, want 2", gotExitCode)
		}
	})
}

func TestPrintVersion(t *testing.T) {
	originalVersion := version
	version = "1.2.3"
	t.Cleanup(func() {
		version = originalVersion
	})

	tests := []struct {
		name      string
		args      []string
		wantPrint bool
		want      string
	}{
		{name: "long flag", args: []string{"--version"}, wantPrint: true, want: "outfitpicker 1.2.3\n"},
		{name: "short flag", args: []string{"-v"}, wantPrint: true, want: "outfitpicker 1.2.3\n"},
		{name: "version command", args: []string{"version"}, wantPrint: true, want: "outfitpicker 1.2.3\n"},
		{name: "no args", args: nil},
		{name: "other arg", args: []string{"--help"}},
		{name: "too many args", args: []string{"version", "extra"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			gotPrint := printVersion(tt.args, &output)

			if gotPrint != tt.wantPrint {
				t.Fatalf("printVersion() = %t, want %t", gotPrint, tt.wantPrint)
			}
			if output.String() != tt.want {
				t.Fatalf("output = %q, want %q", output.String(), tt.want)
			}
		})
	}
}
