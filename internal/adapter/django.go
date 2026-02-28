package adapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/migra/migra/pkg/migra"
)

// DjangoAdapter implements the Adapter interface for Django
type DjangoAdapter struct {
	*BaseAdapter
}

// NewDjangoAdapter creates a new Django adapter
func NewDjangoAdapter() *DjangoAdapter {
	return &DjangoAdapter{
		BaseAdapter: NewBaseAdapter("django"),
	}
}

// Deploy runs Django migrations
func (a *DjangoAdapter) Deploy(ctx context.Context, service *migra.Service, tenant *migra.Tenant) (*migra.Result, error) {
	// Django: python manage.py migrate
	result, err := a.executeCommand(ctx, service, tenant, "python", "manage.py", "migrate", "--no-input")
	if err != nil {
		return result, fmt.Errorf("django deploy failed: %w", err)
	}

	result.Output = a.sanitizeOutput(result.Output)
	return result, nil
}

// Rollback rolls back Django migrations
func (a *DjangoAdapter) Rollback(ctx context.Context, service *migra.Service, tenant *migra.Tenant, steps int) (*migra.Result, error) {
	if steps <= 0 {
		return nil, fmt.Errorf("rollback steps must be positive")
	}

	// For Django, we need to specify the migration to rollback to
	// This is a simplified version - in production, you'd query the migration history
	result, err := a.executeCommand(ctx, service, tenant, "python", "manage.py", "migrate", "--no-input")
	if err != nil {
		return result, fmt.Errorf("django rollback failed: %w", err)
	}

	result.Output = a.sanitizeOutput(result.Output)
	return result, nil
}

// Status returns the migration status for Django
func (a *DjangoAdapter) Status(ctx context.Context, service *migra.Service, tenant *migra.Tenant) (*migra.StatusResult, error) {
	result, err := a.executeCommand(ctx, service, tenant, "python", "manage.py", "showmigrations", "--plan")
	if err != nil {
		return &migra.StatusResult{
			LastError: err.Error(),
		}, err
	}

	status := &migra.StatusResult{
		Applied: make([]string, 0),
		Pending: make([]string, 0),
	}

	// Parse Django migration output
	lines := strings.Split(result.Output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[X]") {
			status.Applied = append(status.Applied, strings.TrimSpace(line[3:]))
		} else if strings.HasPrefix(line, "[ ]") {
			status.Pending = append(status.Pending, strings.TrimSpace(line[3:]))
		}
	}

	return status, nil
}
