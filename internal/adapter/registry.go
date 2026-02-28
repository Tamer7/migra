package adapter

import (
	"fmt"

	"github.com/migra/migra/pkg/migra"
)

// Registry manages adapter instances
type Registry struct {
	adapters map[string]migra.Adapter
}

// NewRegistry creates a new adapter registry
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[string]migra.Adapter),
	}
}

// Register registers an adapter with the registry
func (r *Registry) Register(name string, adapter migra.Adapter) {
	r.adapters[name] = adapter
}

// Get retrieves an adapter by name
func (r *Registry) Get(name string) (migra.Adapter, error) {
	adapter, ok := r.adapters[name]
	if !ok {
		return nil, fmt.Errorf("adapter '%s' not found", name)
	}
	return adapter, nil
}

// GetForService retrieves the appropriate adapter for a service
func (r *Registry) GetForService(service *migra.Service) (migra.Adapter, error) {
	return r.Get(service.Type)
}

// NewDefaultRegistry creates a registry with all built-in adapters
func NewDefaultRegistry() *Registry {
	registry := NewRegistry()
	registry.Register("django", NewDjangoAdapter())
	registry.Register("laravel", NewLaravelAdapter())
	registry.Register("prisma", NewPrismaAdapter())
	return registry
}
