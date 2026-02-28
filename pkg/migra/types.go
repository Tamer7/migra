package migra

import (
	"context"
	"time"
)

// Version information
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// Adapter defines the interface for framework-specific migration adapters
type Adapter interface {
	Deploy(ctx context.Context, service *Service, tenant *Tenant) (*Result, error)
	Rollback(ctx context.Context, service *Service, tenant *Tenant, steps int) (*Result, error)
	Status(ctx context.Context, service *Service, tenant *Tenant) (*StatusResult, error)
	Name() string
}

// Service represents a microservice configuration
type Service struct {
	Name       string            `yaml:"name" json:"name"`
	Type       string            `yaml:"type" json:"type"`
	Path       string            `yaml:"path" json:"path"`
	Env        map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	WorkingDir string            `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`
}

// Tenant represents a tenant in multi-tenant architecture
type Tenant struct {
	ID         string            `json:"id"`
	Connection map[string]string `json:"connection"`
}

// Result represents the result of an adapter operation
type Result struct {
	Success   bool          `json:"success"`
	Output    string        `json:"output"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// StatusResult represents migration status information
type StatusResult struct {
	Applied   []string `json:"applied"`
	Pending   []string `json:"pending"`
	LastError string   `json:"last_error,omitempty"`
}

// ServiceResult represents the result of a service migration
type ServiceResult struct {
	ServiceName string        `json:"service_name"`
	Success     bool          `json:"success"`
	Duration    time.Duration `json:"duration"`
	Error       string        `json:"error,omitempty"`
	Output      string        `json:"output,omitempty"`
}

// Operation defines the type of migration operation
type Operation string

const (
	OperationDeploy   Operation = "deploy"
	OperationRollback Operation = "rollback"
	OperationStatus   Operation = "status"
)
