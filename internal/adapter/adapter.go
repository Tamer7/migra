package adapter

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/migra/migra/pkg/migra"
)

// BaseAdapter provides common functionality for all adapters
type BaseAdapter struct {
	name string
}

// NewBaseAdapter creates a new base adapter
func NewBaseAdapter(name string) *BaseAdapter {
	return &BaseAdapter{
		name: name,
	}
}

// Name returns the adapter name
func (a *BaseAdapter) Name() string {
	return a.name
}

// executeCommand executes a command in the service directory
func (a *BaseAdapter) executeCommand(ctx context.Context, service *migra.Service, tenant *migra.Tenant, command string, args ...string) (*migra.Result, error) {
	start := time.Now()

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = service.WorkingDir

	// Build environment variables
	// Start with parent process environment
	env := cmd.Environ()
	
	// Add service-specific environment variables
	for k, v := range service.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Add tenant-specific environment variables if tenant is provided
	if tenant != nil {
		for k, v := range tenant.Connection {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	cmd.Env = env

	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	result := &migra.Result{
		Success:   err == nil && cmd.ProcessState.ExitCode() == 0,
		Output:    string(output),
		Duration:  duration,
		Timestamp: time.Now(),
	}

	if err != nil {
		result.Error = err.Error()
		if len(output) > 0 {
			result.Error = fmt.Sprintf("%s: %s", err.Error(), string(output))
		}
	}

	return result, nil
}

// sanitizeOutput removes sensitive information from command output
func (a *BaseAdapter) sanitizeOutput(output string) string {
	// Remove common password patterns
	sensitivePatterns := []string{
		"password=",
		"PASSWORD=",
		"pass=",
		"PASS=",
		"secret=",
		"SECRET=",
		"token=",
		"TOKEN=",
	}

	sanitized := output
	for _, pattern := range sensitivePatterns {
		if strings.Contains(strings.ToLower(sanitized), strings.ToLower(pattern)) {
			parts := strings.Split(sanitized, pattern)
			if len(parts) > 1 {
				sanitized = parts[0] + pattern + "[REDACTED]"
			}
		}
	}

	return sanitized
}
