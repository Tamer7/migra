package tenant

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/migra/migra/pkg/migra"
)

// EnvSource loads tenants from environment variable
type EnvSource struct {
	envVar string
}

// NewEnvSource creates a new environment variable tenant source
func NewEnvSource(envVar string) *EnvSource {
	if envVar == "" {
		envVar = "MIGRA_TENANTS"
	}
	return &EnvSource{
		envVar: envVar,
	}
}

// LoadTenants loads tenants from environment variable
// Expected format: "tenant1:db1,tenant2:db2" or "tenant1,tenant2"
func (s *EnvSource) LoadTenants(ctx context.Context) ([]*migra.Tenant, error) {
	value := os.Getenv(s.envVar)
	if value == "" {
		return nil, fmt.Errorf("environment variable %s is not set", s.envVar)
	}

	tenants := make([]*migra.Tenant, 0)
	parts := strings.Split(value, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		tenant := &migra.Tenant{
			Connection: make(map[string]string),
		}

		// Check if it contains connection info
		if strings.Contains(part, ":") {
			subparts := strings.SplitN(part, ":", 2)
			tenant.ID = strings.TrimSpace(subparts[0])
			tenant.Connection["DATABASE_URL"] = strings.TrimSpace(subparts[1])
		} else {
			tenant.ID = part
		}

		tenants = append(tenants, tenant)
	}

	if len(tenants) == 0 {
		return nil, fmt.Errorf("no tenants found in environment variable %s", s.envVar)
	}

	return tenants, nil
}
