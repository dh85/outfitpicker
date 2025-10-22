package app

import (
	"bufio"
	"errors"
	"strings"
	"testing"
)

func TestPrompter_ReadLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"normal input", "hello\n", "hello", false},
		{"input with spaces", "  hello world  \n", "hello world", false},
		{"empty line", "\n", "", false},
		{"eof with content", "hello", "hello", false},
		{"eof without content", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &prompter{r: bufio.NewReader(strings.NewReader(tt.input))}
			result, err := p.readLine()

			if tt.wantErr && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrompter_ReadLine_ReadError(t *testing.T) {
	p := &prompter{r: bufio.NewReader(&errorReader{})}
	_, err := p.readLine()
	if err == nil {
		t.Fatal("expected error but got none")
	}
}

func TestPrompter_ReadLineLower(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"uppercase", "HELLO\n", "hello", false},
		{"mixed case", "HeLLo WoRLd\n", "hello world", false},
		{"already lowercase", "hello\n", "hello", false},
		{"eof with content", "HELLO", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &prompter{r: bufio.NewReader(strings.NewReader(tt.input))}
			result, err := p.readLineLower()

			if tt.wantErr && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrompter_ReadLineLower_Error(t *testing.T) {
	p := &prompter{r: bufio.NewReader(&errorReader{})}
	_, err := p.readLineLower()
	if err == nil {
		t.Fatal("expected error but got none")
	}
}

func TestPrompter_ReadLineLowerDefault(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		def      string
		expected string
		wantErr  bool
	}{
		{"normal input", "HELLO\n", "default", "hello", false},
		{"empty input uses default", "\n", "DEFAULT", "default", false},
		{"eof uses default", "", "DEFAULT", "default", false},
		{"spaces only uses default", "   \n", "DEFAULT", "default", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &prompter{r: bufio.NewReader(strings.NewReader(tt.input))}
			result, err := p.readLineLowerDefault(tt.def)

			if tt.wantErr && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrompter_ReadLineLowerDefault_Error(t *testing.T) {
	p := &prompter{r: bufio.NewReader(&errorReader{})}
	_, err := p.readLineLowerDefault("default")
	if err == nil {
		t.Fatal("expected error but got none")
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}
