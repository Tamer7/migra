package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Loader handles loading configuration from files
type Loader struct {
	configPath string
}

// NewLoader creates a new configuration loader
func NewLoader(configPath string) *Loader {
	return &Loader{
		configPath: configPath,
	}
}

// Load reads and parses the configuration file
func (l *Loader) Load() (*Config, error) {
	data, err := os.ReadFile(l.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	l.applyDefaults(&config)

	// Auto-discover services if enabled
	if config.Discovery != nil && config.Discovery.Enabled {
		if err := l.discoverServices(&config); err != nil {
			return nil, fmt.Errorf("failed to discover services: %w", err)
		}
	}

	return &config, nil
}

// applyDefaults sets default values for missing configuration
func (l *Loader) applyDefaults(config *Config) {
	// Execution defaults
	if config.Execution.Strategy == "" {
		config.Execution.Strategy = DefaultStrategy
	}
	if config.Execution.ParallelLimit == 0 && config.Execution.Strategy == StrategyParallel {
		config.Execution.ParallelLimit = DefaultParallelLimit
	}

	// Global parallel limit
	if config.ParallelLimit == 0 && config.Execution.Strategy == StrategyParallel {
		config.ParallelLimit = DefaultParallelLimit
	}

	// Logging defaults
	if config.Logging.Level == "" {
		config.Logging.Level = DefaultLogLevel
	}
	if config.Logging.Format == "" {
		config.Logging.Format = DefaultLogFormat
	}

	// Tenancy defaults
	if config.Tenancy != nil && config.Tenancy.Enabled {
		if config.Tenancy.Mode == "" {
			config.Tenancy.Mode = TenancyModeDatabase
		}
		if config.Tenancy.MaxParallel == 0 {
			config.Tenancy.MaxParallel = DefaultParallelLimit
		}
	}

	// Service defaults - merge global env and set working directory
	for i := range config.Services {
		if config.Services[i].Env == nil {
			config.Services[i].Env = make(map[string]string)
		}
		for k, v := range config.GlobalEnv {
			if _, exists := config.Services[i].Env[k]; !exists {
				config.Services[i].Env[k] = v
			}
		}
		
		// Set working directory to path if not specified
		if config.Services[i].WorkingDir == "" {
			config.Services[i].WorkingDir = config.Services[i].Path
		}
	}
}

// discoverServices auto-discovers services and adds them to config
func (l *Loader) discoverServices(config *Config) error {
	if config.Discovery.Root == "" {
		return fmt.Errorf("discovery.root is required when discovery is enabled")
	}

	discoverer := NewDiscoverer(config.Discovery.Root)
	discovered, err := discoverer.Discover()
	if err != nil {
		return err
	}

	// Merge discovered services with explicitly defined ones
	// Explicit services take precedence
	existingNames := make(map[string]bool)
	for _, svc := range config.Services {
		existingNames[svc.Name] = true
	}

	for _, svc := range discovered {
		if !existingNames[svc.Name] {
			// Merge global env into discovered service
			if svc.Env == nil {
				svc.Env = make(map[string]string)
			}
			for k, v := range config.GlobalEnv {
				if _, exists := svc.Env[k]; !exists {
					svc.Env[k] = v
				}
			}
			config.Services = append(config.Services, svc)
		}
	}

	return nil
}

// LoadFromFile is a convenience function to load config from a file path
func LoadFromFile(path string) (*Config, error) {
	loader := NewLoader(path)
	return loader.Load()
}
