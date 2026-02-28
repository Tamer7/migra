package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Validator validates configuration
type Validator struct {
	config *Config
	errors []string
}

// NewValidator creates a new configuration validator
func NewValidator(config *Config) *Validator {
	return &Validator{
		config: config,
		errors: make([]string, 0),
	}
}

// Validate performs comprehensive validation of the configuration
func (v *Validator) Validate() error {
	v.validateServices()
	v.validateExecution()
	v.validateTenancy()
	v.validateLogging()

	if len(v.errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(v.errors, "\n  - "))
	}

	return nil
}

// validateServices validates service configurations
func (v *Validator) validateServices() {
	// Services can be empty if discovery is enabled
	if len(v.config.Services) == 0 && (v.config.Discovery == nil || !v.config.Discovery.Enabled) {
		v.addError("at least one service must be defined, or enable service discovery")
		return
	}

	seenNames := make(map[string]bool)
	supportedTypes := map[string]bool{
		FrameworkDjango:  true,
		FrameworkLaravel: true,
		FrameworkPrisma:  true,
	}

	for i, service := range v.config.Services {
		// Validate name
		if service.Name == "" {
			v.addError(fmt.Sprintf("services[%d]: name is required", i))
		} else if seenNames[service.Name] {
			v.addError(fmt.Sprintf("services[%d]: duplicate service name '%s'", i, service.Name))
		} else {
			seenNames[service.Name] = true
		}

		// Validate type
		if service.Type == "" {
			v.addError(fmt.Sprintf("services[%d] (%s): type is required", i, service.Name))
		} else if !supportedTypes[service.Type] {
			v.addError(fmt.Sprintf("services[%d] (%s): unsupported type '%s' (supported: django, laravel, prisma)", i, service.Name, service.Type))
		}

		// Validate path
		if service.Path == "" {
			v.addError(fmt.Sprintf("services[%d] (%s): path is required", i, service.Name))
		} else {
			// Check if path exists
			absPath, err := filepath.Abs(service.Path)
			if err != nil {
				v.addError(fmt.Sprintf("services[%d] (%s): invalid path '%s': %v", i, service.Name, service.Path, err))
			} else {
				if _, err := os.Stat(absPath); os.IsNotExist(err) {
					v.addError(fmt.Sprintf("services[%d] (%s): path does not exist: %s", i, service.Name, absPath))
				}
			}
		}
	}
}

// validateExecution validates execution configuration
func (v *Validator) validateExecution() {
	// Validate strategy
	if v.config.Execution.Strategy != StrategySequential && v.config.Execution.Strategy != StrategyParallel {
		v.addError(fmt.Sprintf("execution.strategy must be 'sequential' or 'parallel', got '%s'", v.config.Execution.Strategy))
	}

	// Validate parallel limit
	if v.config.Execution.Strategy == StrategyParallel {
		if v.config.Execution.ParallelLimit < 1 {
			v.addError("execution.parallel_limit must be at least 1 for parallel execution")
		}
		if v.config.Execution.ParallelLimit > 100 {
			v.addError("execution.parallel_limit should not exceed 100")
		}
	}

	// Global parallel limit
	if v.config.ParallelLimit < 0 {
		v.addError("parallel_limit cannot be negative")
	}
	if v.config.ParallelLimit > 100 {
		v.addError("parallel_limit should not exceed 100")
	}
}

// validateTenancy validates tenancy configuration
func (v *Validator) validateTenancy() {
	if v.config.Tenancy == nil || !v.config.Tenancy.Enabled {
		return
	}

	tenancy := v.config.Tenancy

	// Validate mode
	if tenancy.Mode != TenancyModeDatabase && tenancy.Mode != TenancyModeSchema {
		v.addError(fmt.Sprintf("tenancy.mode must be 'database_per_tenant' or 'schema_per_tenant', got '%s'", tenancy.Mode))
	}

	// Validate tenant source
	if tenancy.TenantSource == "" {
		v.addError("tenancy.tenant_source is required when tenancy is enabled")
	} else if tenancy.TenantSource != TenantSourceEnv && 
		tenancy.TenantSource != TenantSourceFile && 
		tenancy.TenantSource != TenantSourceCommand {
		v.addError(fmt.Sprintf("tenancy.tenant_source must be 'env', 'file', or 'command', got '%s'", tenancy.TenantSource))
	}

	// Validate max parallel
	if tenancy.MaxParallel < 1 {
		v.addError("tenancy.max_parallel must be at least 1")
	}
	if tenancy.MaxParallel > 1000 {
		v.addError("tenancy.max_parallel should not exceed 1000")
	}
}

// validateLogging validates logging configuration
func (v *Validator) validateLogging() {
	validLevels := map[string]bool{
		LogLevelDebug: true,
		LogLevelInfo:  true,
		LogLevelWarn:  true,
		LogLevelError: true,
	}

	if !validLevels[v.config.Logging.Level] {
		v.addError(fmt.Sprintf("logging.level must be 'debug', 'info', 'warn', or 'error', got '%s'", v.config.Logging.Level))
	}

	validFormats := map[string]bool{
		LogFormatConsole: true,
		LogFormatJSON:    true,
	}

	if !validFormats[v.config.Logging.Format] {
		v.addError(fmt.Sprintf("logging.format must be 'console' or 'json', got '%s'", v.config.Logging.Format))
	}
}

// addError adds a validation error
func (v *Validator) addError(msg string) {
	v.errors = append(v.errors, msg)
}

// Validate is a convenience function to validate a config
func Validate(config *Config) error {
	validator := NewValidator(config)
	return validator.Validate()
}
