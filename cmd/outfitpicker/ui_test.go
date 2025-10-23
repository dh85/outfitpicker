package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestShouldUseColors(t *testing.T) {
	tests := []struct {
		name     string
		term     string
		noColor  string
		expected bool
	}{
		{"normal terminal", "xterm-256color", "", true},
		{"dumb terminal", "dumb", "", false},
		{"empty term", "", "", false},
		{"no color set", "xterm", "1", false},
		{"no color empty", "xterm", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origTerm := os.Getenv("TERM")
			origNoColor := os.Getenv("NO_COLOR")
			defer func() {
				os.Setenv("TERM", origTerm)
				os.Setenv("NO_COLOR", origNoColor)
			}()

			os.Setenv("TERM", tt.term)
			os.Setenv("NO_COLOR", tt.noColor)

			result := shouldUseColors()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEnhancedUIMessages(t *testing.T) {
	tests := []struct {
		name     string
		setupCmd func() *cobra.Command
		args     []string
		contains []string
	}{
		{
			name: "config show with enhanced UI",
			setupCmd: func() *cobra.Command {
				return newConfigCmd()
			},
			args:     []string{"show"},
			contains: []string{"config file:"},
		},
		{
			name: "version flag with enhanced output",
			setupCmd: func() *cobra.Command {
				cmd := newRootCmd()
				cmd.SetArgs([]string{"--version"})
				return cmd
			},
			args:     []string{},
			contains: []string{}, // Version output is simple
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := tt.setupCmd()
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			
			if len(tt.args) > 0 {
				cmd.SetArgs(tt.args)
			}
			
			// Don't fail on execution errors for this test
			_ = cmd.Execute()
			
			output := buf.String()
			for _, expected := range tt.contains {
				if expected != "" && !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestUIIntegration(t *testing.T) {
	// Test that UI components are properly integrated
	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}
	
	output := buf.String()
	expectedHelp := []string{
		"outfitpicker",
		"Interactive CLI to pick outfits from category folders",
		"--category",
		"--root",
		"--set-root",
		"--version",
	}
	
	for _, expected := range expectedHelp {
		if !strings.Contains(output, expected) {
			t.Errorf("expected help to contain %q", expected)
		}
	}
}

func TestConfigCommandsWithUI(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "config show",
			args:    []string{"config", "show"},
			wantErr: false,
		},
		{
			name:    "config reset",
			args:    []string{"config", "reset"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := newRootCmd()
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)
			
			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestCompletionCommandWithUI(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}
	
	for _, shell := range shells {
		t.Run("completion_"+shell, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := newRootCmd()
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs([]string{"completion", shell})
			
			err := cmd.Execute()
			if err != nil {
				t.Errorf("completion command failed for %s: %v", shell, err)
			}
			
			output := buf.String()
			if len(output) == 0 {
				t.Errorf("expected completion output for %s", shell)
			}
		})
	}
}

// Benchmark tests for UI performance
func BenchmarkShouldUseColors(b *testing.B) {
	os.Setenv("TERM", "xterm-256color")
	os.Setenv("NO_COLOR", "")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shouldUseColors()
	}
}

func BenchmarkRootCommandHelp(b *testing.B) {
	cmd := newRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = cmd.Execute()
	}
}