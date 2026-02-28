package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/migra/migra/pkg/migra"
)

// Discoverer discovers services by scanning directories
type Discoverer struct {
	rootPath string
}

// NewDiscoverer creates a new service discoverer
func NewDiscoverer(rootPath string) *Discoverer {
	return &Discoverer{
		rootPath: rootPath,
	}
}

// DiscoveryConfig represents auto-discovery configuration
type DiscoveryConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Root    string `yaml:"root" json:"root"`
}

// Discover scans for services and returns discovered service definitions
func (d *Discoverer) Discover() ([]migra.Service, error) {
	services := make([]migra.Service, 0)

	err := filepath.Walk(d.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a directory
		if !info.IsDir() {
			return nil
		}

		// Skip hidden directories and common non-service directories
		if info.Name()[0] == '.' || info.Name() == "node_modules" || info.Name() == "venv" {
			return filepath.SkipDir
		}

		// Check for framework indicators
		frameworkType := d.detectFramework(path)
		if frameworkType != "" {
			// Extract service name from directory
			serviceName := filepath.Base(path)
			
			services = append(services, migra.Service{
				Name:       serviceName,
				Type:       frameworkType,
				Path:       path,
				WorkingDir: path,
				Env:        make(map[string]string),
			})

			// Don't descend into discovered service directories
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	return services, nil
}

// detectFramework checks for framework-specific files
func (d *Discoverer) detectFramework(path string) string {
	// Check for Django
	if fileExists(filepath.Join(path, "manage.py")) {
		return FrameworkDjango
	}

	// Check for Laravel
	if fileExists(filepath.Join(path, "artisan")) {
		return FrameworkLaravel
	}

	// Check for Prisma
	prismaDir := filepath.Join(path, "prisma")
	if dirExists(prismaDir) {
		// Look for .prisma schema files
		entries, err := os.ReadDir(prismaDir)
		if err == nil {
			for _, entry := range entries {
				if filepath.Ext(entry.Name()) == ".prisma" {
					return FrameworkPrisma
				}
			}
		}
	}

	return ""
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
