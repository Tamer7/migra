package logger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsoleLogger(t *testing.T) {
	var buf bytes.Buffer

	t.Run("log messages", func(t *testing.T) {
		buf.Reset()
		logger := NewConsoleLogger(Config{
			Level:  LevelDebug,
			Output: &buf,
		})

		logger.Debug("debug message")
		logger.Info("info message")
		logger.Warn("warn message")
		logger.Error("error message")

		output := buf.String()
		assert.Contains(t, output, "debug message")
		assert.Contains(t, output, "info message")
		assert.Contains(t, output, "warn message")
		assert.Contains(t, output, "error message")
	})

	t.Run("respect log level", func(t *testing.T) {
		buf.Reset()
		logger := NewConsoleLogger(Config{
			Level:  LevelWarn,
			Output: &buf,
		})

		logger.Debug("debug message")
		logger.Info("info message")
		logger.Warn("warn message")

		output := buf.String()
		assert.NotContains(t, output, "debug message")
		assert.NotContains(t, output, "info message")
		assert.Contains(t, output, "warn message")
	})

	t.Run("quiet mode", func(t *testing.T) {
		buf.Reset()
		logger := NewConsoleLogger(Config{
			Level:  LevelInfo,
			Output: &buf,
			Quiet:  true,
		})

		logger.Info("info message")
		logger.Warn("warn message")

		output := buf.String()
		assert.NotContains(t, output, "info message")
		assert.Contains(t, output, "warn message")
	})

	t.Run("with fields", func(t *testing.T) {
		buf.Reset()
		logger := NewConsoleLogger(Config{
			Level:   LevelInfo,
			Output:  &buf,
			Verbose: true,
		})

		logger.Info("test message", F("key", "value"))

		output := buf.String()
		assert.Contains(t, output, "test message")
		assert.Contains(t, output, "key=value")
	})
}

func TestJSONLogger(t *testing.T) {
	var buf bytes.Buffer

	t.Run("log messages as JSON", func(t *testing.T) {
		buf.Reset()
		logger := NewJSONLogger(Config{
			Level:  LevelInfo,
			Output: &buf,
		})

		logger.Info("test message")

		output := buf.String()
		assert.Contains(t, output, `"message":"test message"`)
		assert.Contains(t, output, `"level":"info"`)
	})

	t.Run("with fields", func(t *testing.T) {
		buf.Reset()
		logger := NewJSONLogger(Config{
			Level:  LevelInfo,
			Output: &buf,
		})

		logger.Info("test", F("key", "value"), F("num", 42))

		output := buf.String()
		assert.Contains(t, output, `"key":"value"`)
		assert.Contains(t, output, `"num":42`)
	})
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", LevelDebug},
		{"info", LevelInfo},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"unknown", LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := ParseLevel(tt.input)
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestLevelString(t *testing.T) {
	assert.Equal(t, "debug", LevelDebug.String())
	assert.Equal(t, "info", LevelInfo.String())
	assert.Equal(t, "warn", LevelWarn.String())
	assert.Equal(t, "error", LevelError.String())
}

func TestNewLogger(t *testing.T) {
	t.Run("console logger", func(t *testing.T) {
		logger := NewLogger("console", LevelInfo, false, false)
		assert.NotNil(t, logger)
		_, ok := logger.(*ConsoleLogger)
		assert.True(t, ok)
	})

	t.Run("json logger", func(t *testing.T) {
		logger := NewLogger("json", LevelInfo, false, false)
		assert.NotNil(t, logger)
		_, ok := logger.(*JSONLogger)
		assert.True(t, ok)
	})
}
