package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateManager(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("create new state manager", func(t *testing.T) {
		manager := NewManager(tmpDir)
		assert.NotNil(t, manager)

		err := manager.Load()
		assert.NoError(t, err)

		state := manager.GetState()
		assert.NotNil(t, state)
		assert.Empty(t, state.Services)
	})

	t.Run("record service execution", func(t *testing.T) {
		manager := NewManager(tmpDir)
		err := manager.Load()
		require.NoError(t, err)

		err = manager.RecordServiceExecution("test-service", true, time.Second, nil)
		assert.NoError(t, err)

		state := manager.GetState()
		svcState := state.Services["test-service"]
		assert.NotNil(t, svcState)
		assert.Equal(t, "success", svcState.LastResult)
		assert.Equal(t, 1, svcState.SuccessCount)
		assert.Equal(t, 0, svcState.FailureCount)
	})

	t.Run("save and load state", func(t *testing.T) {
		manager := NewManager(tmpDir)
		err := manager.Load()
		require.NoError(t, err)

		err = manager.RecordServiceExecution("test-service-2", false, time.Second, assert.AnError)
		require.NoError(t, err)

		// Create new manager and load
		manager2 := NewManager(tmpDir)
		err = manager2.Load()
		require.NoError(t, err)

		state := manager2.GetState()
		svcState := state.Services["test-service-2"]
		assert.NotNil(t, svcState)
		assert.Equal(t, "failure", svcState.LastResult)
		assert.Equal(t, 1, svcState.FailureCount)
	})

	t.Run("state file atomicity", func(t *testing.T) {
		manager := NewManager(tmpDir)
		err := manager.Load()
		require.NoError(t, err)

		// Record multiple executions
		for i := 0; i < 5; i++ {
			err = manager.RecordServiceExecution("test-service", true, time.Second, nil)
			assert.NoError(t, err)
		}

		// Verify state file exists and is valid
		stateFile := filepath.Join(tmpDir, ".migra", "state.json")
		_, err = os.Stat(stateFile)
		assert.NoError(t, err)

		// Load state again
		manager2 := NewManager(tmpDir)
		err = manager2.Load()
		require.NoError(t, err)

		state := manager2.GetState()
		assert.Equal(t, 5, state.Services["test-service"].SuccessCount)
	})
}

func TestStateModel(t *testing.T) {
	t.Run("new state", func(t *testing.T) {
		state := NewState()
		assert.NotNil(t, state)
		assert.NotNil(t, state.Services)
		assert.NotNil(t, state.Tenants)
		assert.Equal(t, "1.0", state.Version)
	})

	t.Run("get or create service state", func(t *testing.T) {
		state := NewState()
		
		svcState := state.GetServiceState("test")
		assert.NotNil(t, svcState)
		
		// Should return same instance
		svcState2 := state.GetServiceState("test")
		assert.Equal(t, svcState, svcState2)
	})

	t.Run("record execution", func(t *testing.T) {
		state := NewState()
		
		state.RecordServiceExecution("test", true, time.Second, nil)
		
		svcState := state.Services["test"]
		assert.Equal(t, "success", svcState.LastResult)
		assert.Equal(t, 1, svcState.SuccessCount)
	})
}
