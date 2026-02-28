package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	defaultStateDir  = ".migra"
	defaultStateFile = "state.json"
)

// Manager handles state persistence
type Manager struct {
	stateDir  string
	stateFile string
	state     *State
	mu        sync.RWMutex
}

// NewManager creates a new state manager
func NewManager(workDir string) *Manager {
	stateDir := filepath.Join(workDir, defaultStateDir)
	return &Manager{
		stateDir:  stateDir,
		stateFile: filepath.Join(stateDir, defaultStateFile),
		state:     NewState(),
	}
}

// Load loads the state from disk
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create state directory if it doesn't exist
	if err := os.MkdirAll(m.stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Check if state file exists
	if _, err := os.Stat(m.stateFile); os.IsNotExist(err) {
		// Initialize new state
		m.state = NewState()
		return nil
	}

	// Read state file
	data, err := os.ReadFile(m.stateFile)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	// Parse state
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	m.state = &state
	return nil
}

// Save saves the state to disk atomically
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Ensure state directory exists
	if err := os.MkdirAll(m.stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal state to JSON
	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temporary file
	tempFile := m.stateFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary state file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, m.stateFile); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to save state file: %w", err)
	}

	return nil
}

// GetState returns a copy of the current state
func (m *Manager) GetState() *State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// UpdateState updates the state with a function
func (m *Manager) UpdateState(fn func(*State)) error {
	m.mu.Lock()
	fn(m.state)
	m.mu.Unlock()
	return m.Save()
}

// RecordServiceExecution records a service execution and saves state
func (m *Manager) RecordServiceExecution(serviceName string, success bool, duration time.Duration, err error) error {
	return m.UpdateState(func(s *State) {
		s.RecordServiceExecution(serviceName, success, duration, err)
	})
}

// RecordTenantExecution records a tenant execution and saves state
func (m *Manager) RecordTenantExecution(tenantID, serviceName string, success bool, duration time.Duration, err error) error {
	return m.UpdateState(func(s *State) {
		s.RecordTenantExecution(tenantID, serviceName, success, duration, err)
	})
}

// Clear clears all state
func (m *Manager) Clear() error {
	m.mu.Lock()
	m.state = NewState()
	m.mu.Unlock()
	return m.Save()
}
