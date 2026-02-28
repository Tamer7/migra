package adapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/migra/migra/pkg/migra"
)

// PrismaAdapter implements the Adapter interface for Prisma
type PrismaAdapter struct {
	*BaseAdapter
}

// NewPrismaAdapter creates a new Prisma adapter
func NewPrismaAdapter() *PrismaAdapter {
	return &PrismaAdapter{
		BaseAdapter: NewBaseAdapter("prisma"),
	}
}

// Deploy runs Prisma migrations
func (a *PrismaAdapter) Deploy(ctx context.Context, service *migra.Service, tenant *migra.Tenant) (*migra.Result, error) {
	// Prisma: npx prisma migrate deploy
	result, err := a.executeCommand(ctx, service, tenant, "npx", "prisma", "migrate", "deploy")
	if err != nil {
		return result, fmt.Errorf("prisma deploy failed: %w", err)
	}

	result.Output = a.sanitizeOutput(result.Output)
	return result, nil
}

// Rollback rolls back Prisma migrations
func (a *PrismaAdapter) Rollback(ctx context.Context, service *migra.Service, tenant *migra.Tenant, steps int) (*migra.Result, error) {
	// Prisma doesn't have built-in rollback, this would require custom implementation
	// For now, we return an error indicating it's not supported
	return nil, fmt.Errorf("prisma does not support automatic rollback - manual intervention required")
}

// Status returns the migration status for Prisma
func (a *PrismaAdapter) Status(ctx context.Context, service *migra.Service, tenant *migra.Tenant) (*migra.StatusResult, error) {
	result, err := a.executeCommand(ctx, service, tenant, "npx", "prisma", "migrate", "status")
	if err != nil {
		return &migra.StatusResult{
			LastError: err.Error(),
		}, err
	}

	status := &migra.StatusResult{
		Applied: make([]string, 0),
		Pending: make([]string, 0),
	}

	// Parse Prisma migration status output
	lines := strings.Split(result.Output, "\n")
	inAppliedSection := false
	inPendingSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Detect sections
		if strings.Contains(strings.ToLower(line), "migrations applied") ||
		   strings.Contains(strings.ToLower(line), "following migration") {
			inAppliedSection = true
			inPendingSection = false
			continue
		}
		if strings.Contains(strings.ToLower(line), "pending migration") ||
		   strings.Contains(strings.ToLower(line), "not yet applied") {
			inAppliedSection = false
			inPendingSection = true
			continue
		}

		// Parse migration names (usually in format: 20240101000000_migration_name)
		if strings.Contains(line, "_") && len(line) > 15 {
			if inAppliedSection {
				status.Applied = append(status.Applied, line)
			} else if inPendingSection {
				status.Pending = append(status.Pending, line)
			}
		}
	}

	return status, nil
}
