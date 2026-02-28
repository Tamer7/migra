package logger

import (
	"context"
	"io"
)

// Level represents log level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns string representation of log level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// ParseLevel parses a log level string
func ParseLevel(s string) Level {
	switch s {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Logger defines the logging interface
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	SetLevel(level Level)
	GetLevel() Level
	WithContext(ctx context.Context) Logger
	WithFields(fields ...Field) Logger
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// F is a convenience function for creating fields
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Config represents logger configuration
type Config struct {
	Level   Level
	Format  string
	Output  io.Writer
	Verbose bool
	Quiet   bool
}
