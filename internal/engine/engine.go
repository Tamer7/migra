package engine

import (
	"context"
	"time"

	"github.com/migra/migra/pkg/migra"
)

// Engine defines the interface for migration execution
type Engine interface {
	Execute(ctx context.Context, services []migra.Service, operation migra.Operation) ([]migra.ServiceResult, error)
}

// ExecutionOptions contains options for execution
type ExecutionOptions struct {
	StopOnFailure bool
	DryRun        bool
	ServiceFilter string
}

// Result represents a complete execution result
type Result struct {
	Services     []migra.ServiceResult
	TotalSuccess int
	TotalFailure int
	Duration     time.Duration
}
