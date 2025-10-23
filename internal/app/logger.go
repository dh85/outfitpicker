package app

import (
	"io"
	"log"
	"os"
)

// Logger provides structured logging
type Logger struct {
	debug *log.Logger
	info  *log.Logger
	error *log.Logger
}

// NewLogger creates a new logger
func NewLogger(debugOut, infoOut, errorOut io.Writer) *Logger {
	return &Logger{
		debug: log.New(debugOut, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		info:  log.New(infoOut, "INFO: ", log.Ldate|log.Ltime),
		error: log.New(errorOut, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// DefaultLogger returns a logger with sensible defaults
func DefaultLogger() *Logger {
	return NewLogger(io.Discard, os.Stdout, os.Stderr)
}

func (l *Logger) Debug(v ...interface{}) { l.debug.Println(v...) }
func (l *Logger) Info(v ...interface{})  { l.info.Println(v...) }
func (l *Logger) Error(v ...interface{}) { l.error.Println(v...) }
