package tenant

import (
	"context"

	"github.com/migra/migra/pkg/migra"
)

// Source defines the interface for loading tenants
type Source interface {
	LoadTenants(ctx context.Context) ([]*migra.Tenant, error)
}
