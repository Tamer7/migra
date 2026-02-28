package state

import (
	"time"
)

// State represents the orchestration state
type State struct {
	LastExecution time.Time                  `json:"last_execution"`
	Services      map[string]*ServiceState   `json:"services"`
	Tenants       map[string]*TenantState    `json:"tenants,omitempty"`
	Version       string                     `json:"version"`
}

// ServiceState represents the state of a service
type ServiceState struct {
	LastRun      time.Time `json:"last_run"`
	LastResult   string    `json:"last_result"`
	LastError    string    `json:"last_error,omitempty"`
	SuccessCount int       `json:"success_count"`
	FailureCount int       `json:"failure_count"`
	LastDuration string    `json:"last_duration"`
}

// TenantState represents the state of a tenant
type TenantState struct {
	LastRun      time.Time                  `json:"last_run"`
	Services     map[string]*ServiceState   `json:"services"`
	SuccessCount int                        `json:"success_count"`
	FailureCount int                        `json:"failure_count"`
}

// NewState creates a new state instance
func NewState() *State {
	return &State{
		Services: make(map[string]*ServiceState),
		Tenants:  make(map[string]*TenantState),
		Version:  "1.0",
	}
}

// GetServiceState gets or creates a service state
func (s *State) GetServiceState(serviceName string) *ServiceState {
	if s.Services == nil {
		s.Services = make(map[string]*ServiceState)
	}

	if state, ok := s.Services[serviceName]; ok {
		return state
	}

	state := &ServiceState{}
	s.Services[serviceName] = state
	return state
}

// GetTenantState gets or creates a tenant state
func (s *State) GetTenantState(tenantID string) *TenantState {
	if s.Tenants == nil {
		s.Tenants = make(map[string]*TenantState)
	}

	if state, ok := s.Tenants[tenantID]; ok {
		return state
	}

	state := &TenantState{
		Services: make(map[string]*ServiceState),
	}
	s.Tenants[tenantID] = state
	return state
}

// RecordServiceExecution records a service execution result
func (s *State) RecordServiceExecution(serviceName string, success bool, duration time.Duration, err error) {
	state := s.GetServiceState(serviceName)
	state.LastRun = time.Now()
	state.LastDuration = duration.String()

	if success {
		state.LastResult = "success"
		state.SuccessCount++
		state.LastError = ""
	} else {
		state.LastResult = "failure"
		state.FailureCount++
		if err != nil {
			state.LastError = err.Error()
		}
	}

	s.LastExecution = time.Now()
}

// RecordTenantExecution records a tenant execution result
func (s *State) RecordTenantExecution(tenantID, serviceName string, success bool, duration time.Duration, err error) {
	tenantState := s.GetTenantState(tenantID)
	tenantState.LastRun = time.Now()

	if tenantState.Services == nil {
		tenantState.Services = make(map[string]*ServiceState)
	}

	serviceState, ok := tenantState.Services[serviceName]
	if !ok {
		serviceState = &ServiceState{}
		tenantState.Services[serviceName] = serviceState
	}

	serviceState.LastRun = time.Now()
	serviceState.LastDuration = duration.String()

	if success {
		serviceState.LastResult = "success"
		serviceState.SuccessCount++
		serviceState.LastError = ""
		tenantState.SuccessCount++
	} else {
		serviceState.LastResult = "failure"
		serviceState.FailureCount++
		tenantState.FailureCount++
		if err != nil {
			serviceState.LastError = err.Error()
		}
	}

	s.LastExecution = time.Now()
}
