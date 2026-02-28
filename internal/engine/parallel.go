package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/migra/migra/internal/adapter"
	"github.com/migra/migra/internal/logger"
	"github.com/migra/migra/internal/state"
	"github.com/migra/migra/pkg/migra"
)

// ParallelEngine executes migrations in parallel
type ParallelEngine struct {
	registry      *adapter.Registry
	stateManager  *state.Manager
	logger        logger.Logger
	stopOnFailure bool
	dryRun        bool
	maxParallel   int
}

// NewParallelEngine creates a new parallel execution engine
func NewParallelEngine(registry *adapter.Registry, stateManager *state.Manager, log logger.Logger, stopOnFailure, dryRun bool, maxParallel int) *ParallelEngine {
	if maxParallel <= 0 {
		maxParallel = 5
	}
	return &ParallelEngine{
		registry:      registry,
		stateManager:  stateManager,
		logger:        log,
		stopOnFailure: stopOnFailure,
		dryRun:        dryRun,
		maxParallel:   maxParallel,
	}
}

// Execute executes migrations in parallel
func (e *ParallelEngine) Execute(ctx context.Context, services []migra.Service, operation migra.Operation) ([]migra.ServiceResult, error) {
	results := make([]migra.ServiceResult, len(services))
	resultsMu := sync.Mutex{}

	// Create semaphore for limiting concurrency
	semaphore := make(chan struct{}, e.maxParallel)
	var wg sync.WaitGroup

	// Context for cancellation on failure
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	failed := false
	var failMu sync.Mutex

	for i, service := range services {
		wg.Add(1)

		go func(idx int, svc migra.Service) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-execCtx.Done():
				return
			}

			// Check if we should stop due to previous failure
			failMu.Lock()
			if failed && e.stopOnFailure {
				failMu.Unlock()
				return
			}
			failMu.Unlock()

			e.logger.Info(fmt.Sprintf("Executing %s for service: %s", operation, svc.Name),
				logger.F("service", svc.Name),
				logger.F("type", svc.Type),
			)

			result := e.executeService(execCtx, &svc, operation)

			resultsMu.Lock()
			results[idx] = result
			resultsMu.Unlock()

			if !result.Success {
				e.logger.Error(fmt.Sprintf("Service %s failed", svc.Name),
					logger.F("service", svc.Name),
					logger.F("error", result.Error),
				)

				if e.stopOnFailure {
					failMu.Lock()
					failed = true
					failMu.Unlock()
					cancel()
				}
			} else {
				e.logger.Info(fmt.Sprintf("Service %s completed successfully", svc.Name),
					logger.F("service", svc.Name),
					logger.F("duration", result.Duration.String()),
				)
			}
		}(i, service)
	}

	wg.Wait()

	// Filter out empty results (from stopped executions)
	filteredResults := make([]migra.ServiceResult, 0, len(results))
	for _, r := range results {
		if r.ServiceName != "" {
			filteredResults = append(filteredResults, r)
		}
	}

	return filteredResults, nil
}

// executeService executes migration for a single service
func (e *ParallelEngine) executeService(ctx context.Context, service *migra.Service, operation migra.Operation) migra.ServiceResult {
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
