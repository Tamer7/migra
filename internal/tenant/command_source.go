package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/migra/migra/pkg/migra"
)

// CommandSource loads tenants by executing a command
type CommandSource struct {
	command string
	args    []string
}

// NewCommandSource creates a new command-based tenant source
func NewCommandSource(command string, args ...string) *CommandSource {
	return &CommandSource{
		command: command,
		args:    args,
	}
}

// LoadTenants loads tenants by executing a command
// The command should output JSON array of tenants to stdout
func (s *CommandSource) LoadTenants(ctx context.Context) ([]*migra.Tenant, error) {
	cmd := exec.CommandContext(ctx, s.command, s.args...)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	var tenants []*migra.Tenant
	if err := json.Unmarshal(output, &tenants); err != nil {
		return nil, fmt.Errorf("failed to parse command output as JSON: %w", err)
	}

	if len(tenants) == 0 {
		return nil, fmt.Errorf("command returned no tenants")
	}

	// Validate tenants
	for i, tenant := range tenants {
		if tenant.ID == "" {
			return nil, fmt.Errorf("tenant at index %d has no ID", i)
		}
		if tenant.Connection == nil {
			tenant.Connection = make(map[string]string)
		}
	}

	return tenants, nil
}
