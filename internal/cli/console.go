package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Console interface {
	Prompt(message string) string
	Println(args ...any)
	Printf(format string, args ...any)
	Info(message string)
	Error(message string)
	Warning(message string)
	Success(message string)
}

type TerminalConsole struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func NewTerminalConsole() Console {
	return TerminalConsole{}
}

func (c TerminalConsole) Prompt(message string) string {
	if c.input() == os.Stdin && c.output() == os.Stdout {
		return promptFunc(message)
	}
	return readPrompt(c.output(), c.input(), message)
}

func (c TerminalConsole) Println(args ...any) {
	_, _ = fmt.Fprintln(c.output(), args...)
}

func (c TerminalConsole) Printf(format string, args ...any) {
	_, _ = fmt.Fprintf(c.output(), format, args...)
}

func (c TerminalConsole) Info(message string) {
	c.Println("i", Colorize(sanitizeTerminalText(message), uiBlue))
}

func (c TerminalConsole) Error(message string) {
	_, _ = fmt.Fprintln(c.errorOutput(), "x", Colorize(sanitizeTerminalText(message), uiRed))
}

func (c TerminalConsole) Warning(message string) {
	c.Println("!", Colorize(sanitizeTerminalText(message), uiYellow))
}

func (c TerminalConsole) Success(message string) {
	c.Println("✓", Colorize(sanitizeTerminalText(message), uiGreen))
}

func (c TerminalConsole) input() io.Reader {
	if c.stdin != nil {
		return c.stdin
	}
	return os.Stdin
}

func (c TerminalConsole) output() io.Writer {
	if c.stdout != nil {
		return c.stdout
	}
	return os.Stdout
}

func (c TerminalConsole) errorOutput() io.Writer {
	if c.stderr != nil {
		return c.stderr
	}
	return os.Stderr
}

var defaultConsole Console = TerminalConsole{}

func consoleOrDefault(console Console) Console {
	if console != nil {
		return console
	}
	return defaultConsole
}

func optionalConsole(consoles []Console) Console {
	if len(consoles) == 0 {
		return nil
	}
	return consoles[0]
}

func promptWithConsole(console Console, message string) string {
	return consoleOrDefault(console).Prompt(message)
}

func readPrompt(stdout io.Writer, stdin io.Reader, message string) string {
	_, _ = fmt.Fprint(stdout, Colorize(message, uiBold))

	var input strings.Builder
	buffer := make([]byte, 1)
	for {
		n, err := stdin.Read(buffer)
		if n > 0 {
			if buffer[0] == '\n' {
				break
			}
			input.WriteByte(buffer[0])
		}
		if err != nil {
			break
		}
	}

	return strings.TrimSpace(input.String())
}
