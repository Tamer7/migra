package tenant

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

// Executor handles tenant-aware migration execution
type Executor struct {
	source        Source
	registry      *adapter.Registry
	stateManager  *state.Manager
	logger        logger.Logger
	stopOnFailure bool
	maxParallel   int
}

// NewExecutor creates a new tenant executor
func NewExecutor(source Source, registry *adapter.Registry, stateManager *state.Manager, log logger.Logger, stopOnFailure bool, maxParallel int) *Executor {
	if maxParallel <= 0 {
		maxParallel = 5
	}
	return &Executor{
		source:        source,
		registry:      registry,
		stateManager:  stateManager,
		logger:        log,
		stopOnFailure: stopOnFailure,
		maxParallel:   maxParallel,
	}
}

// TenantResult represents the result of a tenant migration
type TenantResult struct {
	TenantID     string
	Success      bool
	Duration     time.Duration
	Error        string
	ServiceCount int
}

// Execute executes migrations for all tenants
func (e *Executor) Execute(ctx context.Context, services []migra.Service, operation migra.Operation) ([]TenantResult, error) {
	// Load tenants
	e.logger.Info("Loading tenants...")
	tenants, err := e.source.LoadTenants(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load tenants: %w", err)
	}

	e.logger.Info(fmt.Sprintf("Found %d tenants", len(tenants)))

	results := make([]TenantResult, len(tenants))
	resultsMu := sync.Mutex{}

	// Create semaphore for limiting concurrency
	semaphore := make(chan struct{}, e.maxParallel)
	var wg sync.WaitGroup

	// Context for cancellation on failure
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	failed := false
	var failMu sync.Mutex

	for i, tenant := range tenants {
		wg.Add(1)

		go func(idx int, tnt *migra.Tenant) {
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

			e.logger.Info(fmt.Sprintf("Executing %s for tenant: %s", operation, tnt.ID),
				logger.F("tenant", tnt.ID),
			)

			result := e.executeTenant(execCtx, tnt, services, operation)

			resultsMu.Lock()
			results[idx] = result
			resultsMu.Unlock()

			if !result.Success {
				e.logger.Error(fmt.Sprintf("Tenant %s failed", tnt.ID),
					logger.F("tenant", tnt.ID),
					logger.F("error", result.Error),
				)

				if e.stopOnFailure {
					failMu.Lock()
					failed = true
					failMu.Unlock()
					cancel()
				}
			} else {
				e.logger.Info(fmt.Sprintf("Tenant %s completed successfully", tnt.ID),
					logger.F("tenant", tnt.ID),
					logger.F("duration", result.Duration.String()),
				)
			}
		}(i, tenant)
	}

	wg.Wait()

	// Filter out empty results
	filteredResults := make([]TenantResult, 0, len(results))
	for _, r := range results {
		if r.TenantID != "" {
			filteredResults = append(filteredResults, r)
		}
	}

	return filteredResults, nil
}

// executeTenant executes migrations for a single tenant
func (e *Executor) executeTenant(ctx context.Context, tenant *migra.Tenant, services []migra.Service, operation migra.Operation) TenantResult {
	start := time.Now()

	result := TenantResult{
		TenantID:     tenant.ID,
		ServiceCount: len(services),
	}

	successCount := 0
	for _, service := range services {
		// Get adapter
		adp, err := e.registry.GetForService(&service)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to get adapter for service %s: %v", service.Name, err)
			result.Duration = time.Since(start)
			return result
		}

		// Execute operation
		var opResult *migra.Result
		switch operation {
		case migra.OperationDeploy:
			opResult, err = adp.Deploy(ctx, &service, tenant)
		case migra.OperationRollback:
			opResult, err = adp.Rollback(ctx, &service, tenant, 1)
		default:
			err = fmt.Errorf("unsupported operation: %s", operation)
		}

		if err != nil || !opResult.Success {
			result.Success = false
			if err != nil {
				result.Error = fmt.Sprintf("service %s failed: %v", service.Name, err)
			} else {
				result.Error = fmt.Sprintf("service %s failed: %s", service.Name, opResult.Error)
			}
			result.Duration = time.Since(start)

			// Record failure
			e.stateManager.RecordTenantExecution(tenant.ID, service.Name, false, result.Duration, err)
			return result
		}

		successCount++
		e.stateManager.RecordTenantExecution(tenant.ID, service.Name, true, opResult.Duration, nil)
	}

	result.Success = successCount == len(services)
	result.Duration = time.Since(start)
	return result
}
