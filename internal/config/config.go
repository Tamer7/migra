package config

import (
	"github.com/migra/migra/pkg/migra"
)

// Config represents the main configuration structure
type Config struct {
	Services      []migra.Service  `yaml:"services" json:"services"`
	Discovery     *DiscoveryConfig `yaml:"discovery,omitempty" json:"discovery,omitempty"`
	Execution     ExecutionConfig  `yaml:"execution" json:"execution"`
	Tenancy       *TenancyConfig   `yaml:"tenancy,omitempty" json:"tenancy,omitempty"`
	Logging       LoggingConfig    `yaml:"logging" json:"logging"`
	GlobalEnv     map[string]string `yaml:"global_env,omitempty" json:"global_env,omitempty"`
	ParallelLimit int              `yaml:"parallel_limit,omitempty" json:"parallel_limit,omitempty"`
}

// ExecutionConfig defines how migrations should be executed
type ExecutionConfig struct {
	Strategy      string `yaml:"strategy" json:"strategy"`
	StopOnFailure bool   `yaml:"stop_on_failure" json:"stop_on_failure"`
	ParallelLimit int    `yaml:"parallel_limit,omitempty" json:"parallel_limit,omitempty"`
}

// TenancyConfig defines multi-tenant configuration
type TenancyConfig struct {
	Enabled       bool   `yaml:"enabled" json:"enabled"`
	Mode          string `yaml:"mode" json:"mode"`
	TenantSource  string `yaml:"tenant_source" json:"tenant_source"`
	StopOnFailure bool   `yaml:"stop_on_failure" json:"stop_on_failure"`
	MaxParallel   int    `yaml:"max_parallel,omitempty" json:"max_parallel,omitempty"`
}

// LoggingConfig defines logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"`
	File   string `yaml:"file,omitempty" json:"file,omitempty"`
}

// Constants for configuration values
const (
	StrategySequential = "sequential"
	StrategyParallel   = "parallel"

	TenancyModeDatabase = "database_per_tenant"
	TenancyModeSchema   = "schema_per_tenant"

	TenantSourceEnv     = "env"
	TenantSourceFile    = "file"
	TenantSourceCommand = "command"

	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"

	LogFormatConsole = "console"
	LogFormatJSON    = "json"

	FrameworkDjango = "django"
	FrameworkLaravel = "laravel"
	FrameworkPrisma = "prisma"
)

// Default values
const (
	DefaultStrategy      = StrategySequential
	DefaultStopOnFailure = true
	DefaultParallelLimit = 5
	DefaultLogLevel      = LogLevelInfo
	DefaultLogFormat     = LogFormatConsole
)
