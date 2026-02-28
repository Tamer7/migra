package adapter

import (
	"context"
	"testing"

	"github.com/migra/migra/pkg/migra"
	"github.com/stretchr/testify/assert"
)

func TestAdapterRegistry(t *testing.T) {
	registry := NewDefaultRegistry()

	t.Run("get django adapter", func(t *testing.T) {
		adapter, err := registry.Get("django")
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.Equal(t, "django", adapter.Name())
	})

	t.Run("get laravel adapter", func(t *testing.T) {
		adapter, err := registry.Get("laravel")
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.Equal(t, "laravel", adapter.Name())
	})

	t.Run("get prisma adapter", func(t *testing.T) {
		adapter, err := registry.Get("prisma")
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.Equal(t, "prisma", adapter.Name())
	})

	t.Run("get unknown adapter", func(t *testing.T) {
		adapter, err := registry.Get("unknown")
		assert.Error(t, err)
		assert.Nil(t, adapter)
	})

	t.Run("get adapter for service", func(t *testing.T) {
		service := &migra.Service{
			Name: "test",
			Type: "django",
			Path: ".",
		}
		adapter, err := registry.GetForService(service)
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.Equal(t, "django", adapter.Name())
	})
}

func TestDjangoAdapter(t *testing.T) {
	adapter := NewDjangoAdapter()
	assert.Equal(t, "django", adapter.Name())

	// Note: We can't test actual execution without Django installed
	// These would be integration tests
}

func TestLaravelAdapter(t *testing.T) {
	adapter := NewLaravelAdapter()
	assert.Equal(t, "laravel", adapter.Name())
}

func TestPrismaAdapter(t *testing.T) {
	adapter := NewPrismaAdapter()
	assert.Equal(t, "prisma", adapter.Name())

	// Test that rollback returns error (Prisma doesn't support it)
	service := &migra.Service{Name: "test", Type: "prisma", Path: "."}
	result, err := adapter.Rollback(context.Background(), service, nil, 1)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "does not support")
}
