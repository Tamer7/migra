package logger

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

// JSONLogger implements Logger for JSON output
type JSONLogger struct {
	level  Level
	output io.Writer
	fields []Field
	mu     sync.Mutex
}

// NewJSONLogger creates a new JSON logger
func NewJSONLogger(config Config) *JSONLogger {
	output := config.Output
	if output == nil {
		output = os.Stdout
	}

	return &JSONLogger{
		level:  config.Level,
		output: output,
		fields: make([]Field, 0),
	}
}

// Debug logs a debug message
func (l *JSONLogger) Debug(msg string, fields ...Field) {
	if l.level <= LevelDebug {
		l.log(LevelDebug, msg, fields...)
	}
}

// Info logs an info message
func (l *JSONLogger) Info(msg string, fields ...Field) {
	if l.level <= LevelInfo {
		l.log(LevelInfo, msg, fields...)
	}
}

// Warn logs a warning message
func (l *JSONLogger) Warn(msg string, fields ...Field) {
	if l.level <= LevelWarn {
		l.log(LevelWarn, msg, fields...)
	}
}

// Error logs an error message
func (l *JSONLogger) Error(msg string, fields ...Field) {
	if l.level <= LevelError {
		l.log(LevelError, msg, fields...)
	}
}

// log is the internal logging function
func (l *JSONLogger) log(level Level, msg string, fields ...Field) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := make(map[string]interface{})
	entry["timestamp"] = time.Now().Format(time.RFC3339)
	entry["level"] = level.String()
	entry["message"] = msg

	// Add logger fields
	for _, f := range l.fields {
		entry[f.Key] = f.Value
	}

	// Add call fields
	for _, f := range fields {
		entry[f.Key] = f.Value
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	l.output.Write(data)
	l.output.Write([]byte("\n"))
}

// SetLevel sets the log level
func (l *JSONLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level
func (l *JSONLogger) GetLevel() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// WithContext returns a logger with context (no-op for JSON logger)
func (l *JSONLogger) WithContext(ctx context.Context) Logger {
	return l
}

// WithFields returns a logger with additional fields
func (l *JSONLogger) WithFields(fields ...Field) Logger {
	newLogger := &JSONLogger{
		level:  l.level,
		output: l.output,
		fields: append(l.fields, fields...),
	}
	return newLogger
}

// NewLogger creates a logger based on format
func NewLogger(format string, level Level, verbose, quiet bool) Logger {
	config := Config{
		Level:   level,
		Format:  format,
		Verbose: verbose,
		Quiet:   quiet,
	}

	if format == "json" {
		return NewJSONLogger(config)
	}

	return NewConsoleLogger(config)
}
