package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// ConsoleLogger implements Logger for console output with colors
type ConsoleLogger struct {
	level   Level
	output  io.Writer
	verbose bool
	quiet   bool
	fields  []Field
	mu      sync.Mutex
}

// NewConsoleLogger creates a new console logger
func NewConsoleLogger(config Config) *ConsoleLogger {
	output := config.Output
	if output == nil {
		output = os.Stdout
	}

	return &ConsoleLogger{
		level:   config.Level,
		output:  output,
		verbose: config.Verbose,
		quiet:   config.Quiet,
		fields:  make([]Field, 0),
	}
}

// Debug logs a debug message
func (l *ConsoleLogger) Debug(msg string, fields ...Field) {
	if l.level <= LevelDebug && !l.quiet {
		l.log(LevelDebug, msg, fields...)
	}
}

// Info logs an info message
func (l *ConsoleLogger) Info(msg string, fields ...Field) {
	if l.level <= LevelInfo && !l.quiet {
		l.log(LevelInfo, msg, fields...)
	}
}

// Warn logs a warning message
func (l *ConsoleLogger) Warn(msg string, fields ...Field) {
	if l.level <= LevelWarn {
		l.log(LevelWarn, msg, fields...)
	}
}

// Error logs an error message
func (l *ConsoleLogger) Error(msg string, fields ...Field) {
	if l.level <= LevelError {
		l.log(LevelError, msg, fields...)
	}
}

// log is the internal logging function
func (l *ConsoleLogger) log(level Level, msg string, fields ...Field) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := l.formatLevel(level)

	// Combine logger fields with call fields
	allFields := append(l.fields, fields...)

	var fieldsStr string
	if len(allFields) > 0 && l.verbose {
		parts := make([]string, 0, len(allFields))
		for _, f := range allFields {
			parts = append(parts, fmt.Sprintf("%s=%v", f.Key, f.Value))
		}
		fieldsStr = " " + strings.Join(parts, " ")
	}

	line := fmt.Sprintf("%s [%s] %s%s\n", timestamp, levelStr, msg, fieldsStr)
	fmt.Fprint(l.output, line)
}

// formatLevel formats the log level with color
func (l *ConsoleLogger) formatLevel(level Level) string {
	// ANSI color codes
	const (
		colorReset  = "\033[0m"
		colorGray   = "\033[90m"
		colorBlue   = "\033[34m"
		colorYellow = "\033[33m"
		colorRed    = "\033[31m"
	)

	// Check if output is a terminal (simplified - always use colors for now)
	var color string
	switch level {
	case LevelDebug:
		color = colorGray
	case LevelInfo:
		color = colorBlue
	case LevelWarn:
		color = colorYellow
	case LevelError:
		color = colorRed
	}

	return color + strings.ToUpper(level.String()) + colorReset
}

// SetLevel sets the log level
func (l *ConsoleLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level
func (l *ConsoleLogger) GetLevel() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// WithContext returns a logger with context (no-op for console logger)
func (l *ConsoleLogger) WithContext(ctx context.Context) Logger {
	return l
}

// WithFields returns a logger with additional fields
func (l *ConsoleLogger) WithFields(fields ...Field) Logger {
	newLogger := &ConsoleLogger{
		level:   l.level,
		output:  l.output,
		verbose: l.verbose,
		quiet:   l.quiet,
		fields:  append(l.fields, fields...),
	}
	return newLogger
}
