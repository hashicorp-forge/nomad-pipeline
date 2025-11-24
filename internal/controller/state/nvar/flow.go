package nvar

import (
	"errors"
	"fmt"

	"go.uber.org/zap"

	serverstate "github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type Flows struct {
	s *State
}

func (f *Flows) Create(req *serverstate.FlowsCreateReq) (*serverstate.FlowsCreateResp, *serverstate.ErrorResp) {
	f.s.flowsLock.Lock()
	defer f.s.flowsLock.Unlock()

	k := flowCompositeKey{id: req.Flow.ID, namespace: req.Flow.Namespace}

	// Check if flow already exists
	if f.s.enableCache {
		if _, ok := f.s.flowsCache[k]; ok {
			return nil, serverstate.NewErrorResp(errors.New("flow already exists"), 409)
		}
	} else {
		_, err := f.s.getVariable(flowVarPath(req.Flow.Namespace, req.Flow.ID))
		if err == nil {
			return nil, serverstate.NewErrorResp(errors.New("flow already exists"), 409)
		}
	}

	// Create the variable
	v, err := encodeToVariable(flowVarPath(req.Flow.Namespace, req.Flow.ID), req.Flow)
	if err != nil {
		f.s.logger.Error("failed to encode flow", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to encode flow: %w", err), 500)
	}

	if err := f.s.putVariable(v); err != nil {
		f.s.logger.Error("failed to store flow in Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to store flow: %w", err), 500)
	}

	// Update cache
	if f.s.enableCache {
		f.s.flowsCache[k] = req.Flow
	}

	f.s.logger.Debug("flow created",
		zap.String("namespace", req.Flow.Namespace),
		zap.String("flow_id", req.Flow.ID))

	return &serverstate.FlowsCreateResp{Flow: req.Flow}, nil
}

func (f *Flows) Delete(req *serverstate.FlowsDeleteReq) (*serverstate.FlowsDeleteResp, *serverstate.ErrorResp) {
	// Check if any triggers are using the flow
	f.s.triggersLock.Lock()
	if f.s.enableCache {
		for _, trigger := range f.s.triggersCache {
			if trigger.Flow == req.ID && trigger.Namespace == req.Namespace {
				f.s.triggersLock.Unlock()
				return nil, serverstate.NewErrorResp(errors.New("cannot delete flow with linked trigger"), 409)
			}
		}
	} else {
		// Check triggers in Nomad Variables
		triggerVars, err := f.s.listVariablesByPrefix(fmt.Sprintf("%s/%s/", triggersPathPrefix, req.Namespace))
		if err != nil && !isNotFoundError(err) {
			f.s.triggersLock.Unlock()
			f.s.logger.Error("failed to list triggers for flow", zap.Error(err))
			return nil, serverstate.NewErrorResp(fmt.Errorf("failed to check triggers: %w", err), 500)
		}

		// Check each trigger to see if it references this flow
		for _, triggerMeta := range triggerVars {
			v, err := f.s.getVariable(triggerMeta.Path)
			if err != nil {
				continue
			}

			var trigger state.Trigger
			if err := decodeFromVariable(v, &trigger); err != nil {
				continue
			}

			if trigger.Flow == req.ID {
				f.s.triggersLock.Unlock()
				return nil, serverstate.NewErrorResp(errors.New("cannot delete flow with linked trigger"), 409)
			}
		}
	}
	f.s.triggersLock.Unlock()

	k := flowCompositeKey{id: req.ID, namespace: req.Namespace}

	f.s.flowsLock.Lock()
	defer f.s.flowsLock.Unlock()

	// Check if flow exists
	if f.s.enableCache {
		if _, ok := f.s.flowsCache[k]; !ok {
			return nil, serverstate.NewErrorResp(errors.New("flow not found"), 404)
		}
	} else {
		_, err := f.s.getVariable(flowVarPath(req.Namespace, req.ID))
		if err != nil {
			return nil, serverstate.NewErrorResp(errors.New("flow not found"), 404)
		}
	}

	// Delete from Nomad Variables
	if err := f.s.deleteVariable(flowVarPath(req.Namespace, req.ID)); err != nil {
		f.s.logger.Error("failed to delete flow from Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to delete flow: %w", err), 500)
	}

	// Update cache
	if f.s.enableCache {
		delete(f.s.flowsCache, k)
	}

	f.s.logger.Debug("flow deleted",
		zap.String("namespace", req.Namespace),
		zap.String("flow_id", req.ID))

	return &serverstate.FlowsDeleteResp{}, nil
}

func (f *Flows) Get(req *serverstate.FlowsGetReq) (*serverstate.FlowsGetResp, *serverstate.ErrorResp) {
	f.s.flowsLock.RLock()
	defer f.s.flowsLock.RUnlock()

	k := flowCompositeKey{id: req.ID, namespace: req.Namespace}

	// Check cache first
	if f.s.enableCache {
		if flow, ok := f.s.flowsCache[k]; ok {
			return &serverstate.FlowsGetResp{Flow: flow}, nil
		}
		return nil, serverstate.NewErrorResp(errors.New("flow not found"), 404)
	}

	// Read from Nomad Variables
	v, err := f.s.getVariable(flowVarPath(req.Namespace, req.ID))
	if err != nil {
		if isNotFoundError(err) {
			return nil, serverstate.NewErrorResp(errors.New("flow not found"), 404)
		}
		f.s.logger.Error("failed to get flow from Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to get flow: %w", err), 500)
	}

	var flow state.Flow
	if err := decodeFromVariable(v, &flow); err != nil {
		f.s.logger.Error("failed to decode flow", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to decode flow: %w", err), 500)
	}

	return &serverstate.FlowsGetResp{Flow: &flow}, nil
}

func (f *Flows) List(req *serverstate.FlowsListReq) (*serverstate.FlowsListResp, *serverstate.ErrorResp) {
	f.s.flowsLock.RLock()
	defer f.s.flowsLock.RUnlock()

	resp := serverstate.FlowsListResp{}

	// Use cache if enabled
	if f.s.enableCache {
		for _, flow := range f.s.flowsCache {
			if req.Namespace == "*" || flow.Namespace == req.Namespace {
				resp.Flows = append(resp.Flows, flow.Stub())
			}
		}
		return &resp, nil
	}

	// List from Nomad Variables
	var prefix string
	if req.Namespace == "*" {
		prefix = flowsPathPrefix + "/"
	} else {
		prefix = fmt.Sprintf("%s/%s/", flowsPathPrefix, req.Namespace)
	}

	vars, err := f.s.listVariablesByPrefix(prefix)
	if err != nil && !isNotFoundError(err) {
		f.s.logger.Error("failed to list flows from Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to list flows: %w", err), 500)
	}

	for _, varMeta := range vars {
		v, err := f.s.getVariable(varMeta.Path)
		if err != nil {
			f.s.logger.Warn("failed to get flow", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		var flow state.Flow
		if err := decodeFromVariable(v, &flow); err != nil {
			f.s.logger.Warn("failed to decode flow", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		resp.Flows = append(resp.Flows, flow.Stub())
	}

	return &resp, nil
}

// loadFlowsCache loads all flows into the cache
func (s *State) loadFlowsCache() error {
	s.flowsLock.Lock()
	defer s.flowsLock.Unlock()

	vars, err := s.listVariablesByPrefix(flowsPathPrefix + "/")
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return err
	}

	for _, varMeta := range vars {
		v, err := s.getVariable(varMeta.Path)
		if err != nil {
			s.logger.Warn("failed to get flow for cache", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		var flow state.Flow
		if err := decodeFromVariable(v, &flow); err != nil {
			s.logger.Warn("failed to decode flow for cache", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		k := flowCompositeKey{id: flow.ID, namespace: flow.Namespace}
		s.flowsCache[k] = &flow
	}

	s.logger.Debug("loaded flows into cache", zap.Int("count", len(s.flowsCache)))
	return nil
}
