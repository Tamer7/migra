package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/migra/migra/internal/adapter"
	"github.com/migra/migra/internal/logger"
	"github.com/migra/migra/internal/state"
	"github.com/migra/migra/pkg/migra"
)

// SequentialEngine executes migrations sequentially
type SequentialEngine struct {
	registry      *adapter.Registry
	stateManager  *state.Manager
	logger        logger.Logger
	stopOnFailure bool
	dryRun        bool
}

// NewSequentialEngine creates a new sequential execution engine
func NewSequentialEngine(registry *adapter.Registry, stateManager *state.Manager, log logger.Logger, stopOnFailure, dryRun bool) *SequentialEngine {
	return &SequentialEngine{
		registry:      registry,
		stateManager:  stateManager,
		logger:        log,
		stopOnFailure: stopOnFailure,
		dryRun:        dryRun,
	}
}

// Execute executes migrations sequentially
func (e *SequentialEngine) Execute(ctx context.Context, services []migra.Service, operation migra.Operation) ([]migra.ServiceResult, error) {
	results := make([]migra.ServiceResult, 0, len(services))

	for _, service := range services {
		e.logger.Info(fmt.Sprintf("Executing %s for service: %s", operation, service.Name),
			logger.F("service", service.Name),
			logger.F("type", service.Type),
		)

		result := e.executeService(ctx, &service, operation)
		results = append(results, result)

		if !result.Success {
			e.logger.Error(fmt.Sprintf("Service %s failed", service.Name),
				logger.F("service", service.Name),
				logger.F("error", result.Error),
			)

			if e.stopOnFailure {
				e.logger.Warn("Stopping execution due to failure")
				break
			}
		} else {
			e.logger.Info(fmt.Sprintf("Service %s completed successfully", service.Name),
				logger.F("service", service.Name),
				logger.F("duration", result.Duration.String()),
			)
		}
	}

	return results, nil
}

// executeService executes migration for a single service
func (e *SequentialEngine) executeService(ctx context.Context, service *migra.Service, operation migra.Operation) migra.ServiceResult {
	start := time.Now()

	result := migra.ServiceResult{
		ServiceName: service.Name,
	}

	if e.dryRun {
		e.logger.Info("[DRY RUN] Would execute migration", logger.F("service", service.Name))
		result.Success = true
		result.Duration = time.Since(start)
		result.Output = "Dry run - no actual execution"
		return result
	}

	// Get adapter for service
	adp, err := e.registry.GetForService(service)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(start)
		e.stateManager.RecordServiceExecution(service.Name, false, result.Duration, err)
		return result
	}

	// Execute operation
	var opResult *migra.Result
	switch operation {
	case migra.OperationDeploy:
		opResult, err = adp.Deploy(ctx, service, nil)
	case migra.OperationRollback:
		opResult, err = adp.Rollback(ctx, service, nil, 1)
	case migra.OperationStatus:
		statusResult, statusErr := adp.Status(ctx, service, nil)
		if statusErr != nil {
			err = statusErr
		}
		opResult = &migra.Result{
			Success:   statusErr == nil,
			Output:    fmt.Sprintf("Applied: %d, Pending: %d", len(statusResult.Applied), len(statusResult.Pending)),
			Timestamp: time.Now(),
		}
	default:
		err = fmt.Errorf("unsupported operation: %s", operation)
	}

	result.Duration = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		if opResult != nil {
			result.Output = opResult.Output
		}
		e.stateManager.RecordServiceExecution(service.Name, false, result.Duration, err)
		return result
	}

	result.Success = opResult.Success
	result.Output = opResult.Output
	if !opResult.Success {
		result.Error = opResult.Error
	}

	e.stateManager.RecordServiceExecution(service.Name, result.Success, result.Duration, nil)
	return result
}
