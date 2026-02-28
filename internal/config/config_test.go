package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/migra/migra/pkg/migra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "migra.yaml")

	cfgContent := `
services:
  - name: test-service
    type: django
    path: ./test

execution:
  strategy: sequential
  stop_on_failure: true

logging:
  level: info
  format: console
`

	err := os.WriteFile(cfgPath, []byte(cfgContent), 0644)
	require.NoError(t, err)

	// Load config
	cfg, err := LoadFromFile(cfgPath)
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Len(t, cfg.Services, 1)
	assert.Equal(t, "test-service", cfg.Services[0].Name)
	assert.Equal(t, "django", cfg.Services[0].Type)
	assert.Equal(t, StrategySequential, cfg.Execution.Strategy)
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Services: []migra.Service{
					{Name: "svc1", Type: FrameworkDjango, Path: "."},
				},
				Execution: ExecutionConfig{
					Strategy:      StrategySequential,
					StopOnFailure: true,
				},
				Logging: LoggingConfig{
					Level:  LogLevelInfo,
					Format: LogFormatConsole,
				},
			},
			wantErr: false,
		},
		{
			name: "missing services",
			config: &Config{
				Services: []migra.Service{},
				Execution: ExecutionConfig{
					Strategy: StrategySequential,
				},
				Logging: LoggingConfig{
					Level:  LogLevelInfo,
					Format: LogFormatConsole,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid strategy",
			config: &Config{
				Services: []migra.Service{
					{Name: "svc1", Type: FrameworkDjango, Path: "."},
				},
				Execution: ExecutionConfig{
					Strategy: "invalid",
				},
				Logging: LoggingConfig{
					Level:  LogLevelInfo,
					Format: LogFormatConsole,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
