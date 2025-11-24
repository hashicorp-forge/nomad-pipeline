package nvar

import (
	"errors"
	"fmt"

	"go.uber.org/zap"

	serverstate "github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type Namespaces struct {
	s *State
}

func (n *Namespaces) Create(req *serverstate.NamespacesCreateReq) (*serverstate.NamespacesCreateResp, *serverstate.ErrorResp) {
	n.s.namespacesLock.Lock()
	defer n.s.namespacesLock.Unlock()

	// Check if namespace already exists
	if n.s.enableCache {
		if _, ok := n.s.namespacesCache[req.Namespace.ID]; ok {
			return nil, serverstate.NewErrorResp(errors.New("namespace already exists"), 409)
		}
	} else {
		// Check directly in Nomad Variables
		_, err := n.s.getVariable(namespaceVarPath(req.Namespace.ID))
		if err == nil {
			return nil, serverstate.NewErrorResp(errors.New("namespace already exists"), 409)
		}
	}

	// Create the variable
	v, err := encodeToVariable(namespaceVarPath(req.Namespace.ID), req.Namespace)
	if err != nil {
		n.s.logger.Error("failed to encode namespace", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to encode namespace: %w", err), 500)
	}

	if err := n.s.putVariable(v); err != nil {
		n.s.logger.Error("failed to store namespace in Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to store namespace: %w", err), 500)
	}

	// Update cache
	if n.s.enableCache {
		n.s.namespacesCache[req.Namespace.ID] = req.Namespace
	}

	n.s.logger.Debug("namespace created", zap.String("namespace", req.Namespace.ID))
	return &serverstate.NamespacesCreateResp{}, nil
}

func (n *Namespaces) Delete(req *serverstate.NamespacesDeleteReq) (*serverstate.NamespacesDeleteResp, *serverstate.ErrorResp) {
	// Check if any flows are using the namespace
	n.s.flowsLock.Lock()
	if n.s.enableCache {
		for _, flow := range n.s.flowsCache {
			if flow.Namespace == req.Name {
				n.s.flowsLock.Unlock()
				return nil, serverstate.NewErrorResp(errors.New("cannot delete in-use namespace"), 409)
			}
		}
	} else {
		// Check flows in Nomad Variables
		flowVars, err := n.s.listVariablesByPrefix(fmt.Sprintf("%s/%s/", flowsPathPrefix, req.Name))
		if err != nil {
			n.s.flowsLock.Unlock()
			n.s.logger.Error("failed to list flows for namespace", zap.Error(err))
			return nil, serverstate.NewErrorResp(fmt.Errorf("failed to check flows: %w", err), 500)
		}
		if len(flowVars) > 0 {
			n.s.flowsLock.Unlock()
			return nil, serverstate.NewErrorResp(errors.New("cannot delete in-use namespace"), 409)
		}
	}
	n.s.flowsLock.Unlock()

	n.s.namespacesLock.Lock()
	defer n.s.namespacesLock.Unlock()

	// Check if namespace exists
	if n.s.enableCache {
		if _, ok := n.s.namespacesCache[req.Name]; !ok {
			return nil, serverstate.NewErrorResp(errors.New("namespace not found"), 404)
		}
	} else {
		_, err := n.s.getVariable(namespaceVarPath(req.Name))
		if err != nil {
			return nil, serverstate.NewErrorResp(errors.New("namespace not found"), 404)
		}
	}

	// Delete from Nomad Variables
	if err := n.s.deleteVariable(namespaceVarPath(req.Name)); err != nil {
		n.s.logger.Error("failed to delete namespace from Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to delete namespace: %w", err), 500)
	}

	// Update cache
	if n.s.enableCache {
		delete(n.s.namespacesCache, req.Name)
	}

	n.s.logger.Debug("namespace deleted", zap.String("namespace", req.Name))
	return &serverstate.NamespacesDeleteResp{}, nil
}

func (n *Namespaces) Get(req *serverstate.NamespacesGetReq) (*serverstate.NamespacesGetResp, *serverstate.ErrorResp) {
	n.s.namespacesLock.RLock()
	defer n.s.namespacesLock.RUnlock()

	// Check cache first
	if n.s.enableCache {
		if ns, ok := n.s.namespacesCache[req.Name]; ok {
			return &serverstate.NamespacesGetResp{Namespace: ns}, nil
		}
		return nil, serverstate.NewErrorResp(errors.New("namespace not found"), 404)
	}

	// Read from Nomad Variables
	v, err := n.s.getVariable(namespaceVarPath(req.Name))
	if err != nil {
		if isNotFoundError(err) {
			return nil, serverstate.NewErrorResp(errors.New("namespace not found"), 404)
		}
		n.s.logger.Error("failed to get namespace from Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to get namespace: %w", err), 500)
	}

	var ns state.Namespace
	if err := decodeFromVariable(v, &ns); err != nil {
		n.s.logger.Error("failed to decode namespace", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to decode namespace: %w", err), 500)
	}

	return &serverstate.NamespacesGetResp{Namespace: &ns}, nil
}

func (n *Namespaces) List(_ *serverstate.NamespacesListReq) (*serverstate.NamespacesListResp, *serverstate.ErrorResp) {
	n.s.namespacesLock.RLock()
	defer n.s.namespacesLock.RUnlock()

	resp := serverstate.NamespacesListResp{}

	// Use cache if enabled
	if n.s.enableCache {
		for _, ns := range n.s.namespacesCache {
			resp.Namespaces = append(resp.Namespaces, ns.Stub())
		}
		return &resp, nil
	}

	// List from Nomad Variables
	vars, err := n.s.listVariablesByPrefix(namespacesPathPrefix + "/")
	if err != nil {
		n.s.logger.Error("failed to list namespaces from Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to list namespaces: %w", err), 500)
	}

	for _, varMeta := range vars {
		v, err := n.s.getVariable(varMeta.Path)
		if err != nil {
			n.s.logger.Warn("failed to get namespace", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		var ns state.Namespace
		if err := decodeFromVariable(v, &ns); err != nil {
			n.s.logger.Warn("failed to decode namespace", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		resp.Namespaces = append(resp.Namespaces, ns.Stub())
	}

	return &resp, nil
}

// loadNamespacesCache loads all namespaces into the cache
func (s *State) loadNamespacesCache() error {
	s.namespacesLock.Lock()
	defer s.namespacesLock.Unlock()

	vars, err := s.listVariablesByPrefix(namespacesPathPrefix + "/")
	if err != nil {
		// If no variables found, that's okay - just means empty state
		if isNotFoundError(err) {
			return nil
		}
		return err
	}

	for _, varMeta := range vars {
		v, err := s.getVariable(varMeta.Path)
		if err != nil {
			s.logger.Warn("failed to get namespace for cache", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		var ns state.Namespace
		if err := decodeFromVariable(v, &ns); err != nil {
			s.logger.Warn("failed to decode namespace for cache", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		s.namespacesCache[ns.ID] = &ns
	}

	s.logger.Debug("loaded namespaces into cache", zap.Int("count", len(s.namespacesCache)))
	return nil
}

// isNotFoundError checks if an error is a "not found" error from Nomad API
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Nomad API returns nil variable and nil error when not found
	// But also check for API error responses
	errStr := err.Error()
	return errStr == "Unexpected response code: 404" || errStr == "variable not found"
}
