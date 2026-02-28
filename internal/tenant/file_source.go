package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/migra/migra/pkg/migra"
	"gopkg.in/yaml.v3"
)

// FileSource loads tenants from a file
type FileSource struct {
	filePath string
}

// NewFileSource creates a new file-based tenant source
func NewFileSource(filePath string) *FileSource {
	return &FileSource{
		filePath: filePath,
	}
}

// LoadTenants loads tenants from a file
// Supports both JSON and YAML formats
func (s *FileSource) LoadTenants(ctx context.Context) ([]*migra.Tenant, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tenants file: %w", err)
	}

	var tenants []*migra.Tenant

	// Try JSON first
	if err := json.Unmarshal(data, &tenants); err != nil {
		// Try YAML
		if err := yaml.Unmarshal(data, &tenants); err != nil {
			return nil, fmt.Errorf("failed to parse tenants file (tried JSON and YAML): %w", err)
		}
	}

	if len(tenants) == 0 {
		return nil, fmt.Errorf("no tenants found in file %s", s.filePath)
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
