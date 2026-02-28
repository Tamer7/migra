package adapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/migra/migra/pkg/migra"
)

// LaravelAdapter implements the Adapter interface for Laravel
type LaravelAdapter struct {
	*BaseAdapter
}

// NewLaravelAdapter creates a new Laravel adapter
func NewLaravelAdapter() *LaravelAdapter {
	return &LaravelAdapter{
		BaseAdapter: NewBaseAdapter("laravel"),
	}
}

// Deploy runs Laravel migrations
func (a *LaravelAdapter) Deploy(ctx context.Context, service *migra.Service, tenant *migra.Tenant) (*migra.Result, error) {
	// Laravel: php artisan migrate --force
	result, err := a.executeCommand(ctx, service, tenant, "php", "artisan", "migrate", "--force")
	if err != nil {
		return result, fmt.Errorf("laravel deploy failed: %w", err)
	}

	result.Output = a.sanitizeOutput(result.Output)
	return result, nil
}

// Rollback rolls back Laravel migrations
func (a *LaravelAdapter) Rollback(ctx context.Context, service *migra.Service, tenant *migra.Tenant, steps int) (*migra.Result, error) {
	if steps <= 0 {
		return nil, fmt.Errorf("rollback steps must be positive")
	}

	// Laravel: php artisan migrate:rollback --step=N --force
	result, err := a.executeCommand(ctx, service, tenant, "php", "artisan", "migrate:rollback", 
		fmt.Sprintf("--step=%d", steps), "--force")
	if err != nil {
		return result, fmt.Errorf("laravel rollback failed: %w", err)
	}

	result.Output = a.sanitizeOutput(result.Output)
	return result, nil
}

// Status returns the migration status for Laravel
func (a *LaravelAdapter) Status(ctx context.Context, service *migra.Service, tenant *migra.Tenant) (*migra.StatusResult, error) {
	result, err := a.executeCommand(ctx, service, tenant, "php", "artisan", "migrate:status")
	if err != nil {
		return &migra.StatusResult{
			LastError: err.Error(),
		}, err
	}

	status := &migra.StatusResult{
		Applied: make([]string, 0),
		Pending: make([]string, 0),
	}

	// Parse Laravel migration status output
	lines := strings.Split(result.Output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "+") || strings.HasPrefix(line, "|") {
			continue
		}

		// Laravel outputs: Ran | Y | migration_name
		// or: Pending | N | migration_name
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			statusStr := strings.TrimSpace(parts[1])
			migration := strings.TrimSpace(parts[2])

			if statusStr == "Y" || strings.ToLower(statusStr) == "ran" {
				status.Applied = append(status.Applied, migration)
			} else if statusStr == "N" || strings.ToLower(statusStr) == "pending" {
				status.Pending = append(status.Pending, migration)
			}
		}
	}

	return status, nil
}
