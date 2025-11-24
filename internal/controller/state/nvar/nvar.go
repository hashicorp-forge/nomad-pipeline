package nvar

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hashicorp/nomad/api"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"

	serverstate "github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/logger"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

const (
	// Nomad Variables path prefixes for different resource types
	variablePathPrefix   = "nomad-pipeline"
	namespacesPathPrefix = variablePathPrefix + "/namespaces"
	flowsPathPrefix      = variablePathPrefix + "/flows"
	runsPathPrefix       = variablePathPrefix + "/runs"
	triggersPathPrefix   = variablePathPrefix + "/triggers"
)

// State implements the serverstate.State interface using Nomad Variables
// for persistent storage. This allows the state to be shared across multiple
// controller instances and survives controller restarts.
type State struct {
	client *api.Client
	logger *zap.Logger

	// enableCache indicates whether local caching is enabled. When enabled, the
	// maps and locks below are used to cache state data in memory for faster
	// read access.
	enableCache bool

	namespacesCache map[string]*state.Namespace
	namespacesLock  sync.RWMutex

	flowsCache map[flowCompositeKey]*state.Flow
	flowsLock  sync.RWMutex

	runsCache map[runCompositeKey]*state.Run
	runsLock  sync.RWMutex

	triggersCache map[triggerCompositeKey]*state.Trigger
	triggersLock  sync.RWMutex
}

// New creates a new Nomad Variables-backed state implementation
func New(cache bool, zLogger *zap.Logger, client *api.Client) (serverstate.State, error) {

	s := &State{
		client:          client,
		enableCache:     cache,
		logger:          zLogger.Named(logger.ComponentNameState),
		namespacesCache: make(map[string]*state.Namespace),
		flowsCache:      make(map[flowCompositeKey]*state.Flow),
		runsCache:       make(map[runCompositeKey]*state.Run),
		triggersCache:   make(map[triggerCompositeKey]*state.Trigger),
	}

	// Initialize the state by loading all data into cache if caching is enabled
	if s.enableCache {
		if err := s.loadCache(); err != nil {
			s.logger.Warn("failed to load initial cache", zap.Error(err))
		}
	}

	return s, nil
}

// loadCache loads all state data from Nomad Variables into the local cache
func (s *State) loadCache() error {
	s.logger.Debug("loading cache from Nomad Variables")

	// Load namespaces
	if err := s.loadNamespacesCache(); err != nil {
		return fmt.Errorf("failed to load namespaces: %w", err)
	}

	// Load flows
	if err := s.loadFlowsCache(); err != nil {
		return fmt.Errorf("failed to load flows: %w", err)
	}

	// Load runs
	if err := s.loadRunsCache(); err != nil {
		return fmt.Errorf("failed to load runs: %w", err)
	}

	// Load triggers
	if err := s.loadTriggersCache(); err != nil {
		return fmt.Errorf("failed to load triggers: %w", err)
	}

	s.logger.Debug("cache loaded successfully")
	return nil
}

// Interface implementation methods
func (s *State) Flows() serverstate.Flows {
	return &Flows{s: s}
}

func (s *State) Namespaces() serverstate.Namespaces {
	return &Namespaces{s: s}
}

func (s *State) Runs() serverstate.Runs {
	return &Runs{s: s}
}

func (s *State) Triggers() serverstate.Triggers {
	return &Triggers{s: s}
}

// Composite key types for multi-field indexing
type flowCompositeKey struct {
	id        string
	namespace string
}

type runCompositeKey struct {
	id        ulid.ULID
	namespace string
}

type triggerCompositeKey struct {
	id        string
	namespace string
}

// Helper functions for variable path generation
func namespaceVarPath(name string) string {
	return fmt.Sprintf("%s/%s", namespacesPathPrefix, name)
}

func flowVarPath(namespace, id string) string {
	return fmt.Sprintf("%s/%s/%s", flowsPathPrefix, namespace, id)
}

func runVarPath(namespace string, id ulid.ULID) string {
	return fmt.Sprintf("%s/%s/%s", runsPathPrefix, namespace, id.String())
}

func triggerVarPath(namespace, id string) string {
	return fmt.Sprintf("%s/%s/%s", triggersPathPrefix, namespace, id)
}

// Helper functions for serialization
func encodeToVariable(path string, data any) (*api.Variable, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	return &api.Variable{
		Path: path,
		Items: map[string]string{
			"data": string(jsonData),
		},
	}, nil
}

func decodeFromVariable(v *api.Variable, target any) error {
	data, ok := v.Items["data"]
	if !ok {
		return fmt.Errorf("variable missing 'data' field")
	}

	if err := json.Unmarshal([]byte(data), target); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// getVariable retrieves a variable from Nomad
func (s *State) getVariable(path string) (*api.Variable, error) {
	v, _, err := s.client.Variables().Read(path, &api.QueryOptions{})
	if err != nil {
		return nil, err
	}
	return v, nil
}

// putVariable stores a variable in Nomad
func (s *State) putVariable(v *api.Variable) error {
	_, _, err := s.client.Variables().Update(v, &api.WriteOptions{})
	return err
}

// deleteVariable deletes a variable from Nomad
func (s *State) deleteVariable(path string) error {
	_, err := s.client.Variables().Delete(path, &api.WriteOptions{})
	return err
}

// listVariablesByPrefix lists all variables with a given prefix
func (s *State) listVariablesByPrefix(prefix string) ([]*api.VariableMetadata, error) {
	vars, _, err := s.client.Variables().PrefixList(prefix, &api.QueryOptions{})
	if err != nil {
		return nil, err
	}
	return vars, nil
}
