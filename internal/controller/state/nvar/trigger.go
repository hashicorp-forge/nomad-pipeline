package nvar

import (
	"errors"
	"fmt"

	"go.uber.org/zap"

	serverstate "github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type Triggers struct {
	s *State
}

func (t *Triggers) Create(req *serverstate.TriggersCreateReq) (*serverstate.TriggersCreateResp, *serverstate.ErrorResp) {
	t.s.triggersLock.Lock()
	defer t.s.triggersLock.Unlock()

	k := triggerCompositeKey{id: req.Trigger.ID, namespace: req.Trigger.Namespace}

	// Check if trigger already exists
	if t.s.enableCache {
		if _, ok := t.s.triggersCache[k]; ok {
			return nil, serverstate.NewErrorResp(errors.New("trigger already exists"), 409)
		}
	} else {
		_, err := t.s.getVariable(triggerVarPath(req.Trigger.Namespace, req.Trigger.ID))
		if err == nil {
			return nil, serverstate.NewErrorResp(errors.New("trigger already exists"), 409)
		}
	}

	// Create the variable
	v, err := encodeToVariable(triggerVarPath(req.Trigger.Namespace, req.Trigger.ID), req.Trigger)
	if err != nil {
		t.s.logger.Error("failed to encode trigger", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to encode trigger: %w", err), 500)
	}

	if err := t.s.putVariable(v); err != nil {
		t.s.logger.Error("failed to store trigger in Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to store trigger: %w", err), 500)
	}

	// Update cache
	if t.s.enableCache {
		t.s.triggersCache[k] = req.Trigger
	}

	t.s.logger.Debug("trigger created",
		zap.String("namespace", req.Trigger.Namespace),
		zap.String("trigger_id", req.Trigger.ID))

	return &serverstate.TriggersCreateResp{}, nil
}

func (t *Triggers) Delete(req *serverstate.TriggersDeleteReq) (*serverstate.TriggersDeleteResp, *serverstate.ErrorResp) {
	t.s.triggersLock.Lock()
	defer t.s.triggersLock.Unlock()

	k := triggerCompositeKey{id: req.ID, namespace: req.Namespace}

	// Check if trigger exists
	if t.s.enableCache {
		if _, ok := t.s.triggersCache[k]; !ok {
			return nil, serverstate.NewErrorResp(errors.New("trigger not found"), 404)
		}
	} else {
		_, err := t.s.getVariable(triggerVarPath(req.Namespace, req.ID))
		if err != nil {
			return nil, serverstate.NewErrorResp(errors.New("trigger not found"), 404)
		}
	}

	// Delete from Nomad Variables
	if err := t.s.deleteVariable(triggerVarPath(req.Namespace, req.ID)); err != nil {
		t.s.logger.Error("failed to delete trigger from Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to delete trigger: %w", err), 500)
	}

	// Update cache
	if t.s.enableCache {
		delete(t.s.triggersCache, k)
	}

	t.s.logger.Debug("trigger deleted",
		zap.String("namespace", req.Namespace),
		zap.String("trigger_id", req.ID))

	return &serverstate.TriggersDeleteResp{}, nil
}

func (t *Triggers) Get(req *serverstate.TriggersGetReq) (*serverstate.TriggersGetResp, *serverstate.ErrorResp) {
	t.s.triggersLock.RLock()
	defer t.s.triggersLock.RUnlock()

	k := triggerCompositeKey{id: req.ID, namespace: req.Namespace}

	// Check cache first
	if t.s.enableCache {
		if trigger, ok := t.s.triggersCache[k]; ok {
			return &serverstate.TriggersGetResp{Trigger: trigger}, nil
		}
		return nil, serverstate.NewErrorResp(errors.New("trigger not found"), 404)
	}

	// Read from Nomad Variables
	v, err := t.s.getVariable(triggerVarPath(req.Namespace, req.ID))
	if err != nil {
		if isNotFoundError(err) {
			return nil, serverstate.NewErrorResp(errors.New("trigger not found"), 404)
		}
		t.s.logger.Error("failed to get trigger from Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to get trigger: %w", err), 500)
	}

	var trigger state.Trigger
	if err := decodeFromVariable(v, &trigger); err != nil {
		t.s.logger.Error("failed to decode trigger", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to decode trigger: %w", err), 500)
	}

	return &serverstate.TriggersGetResp{Trigger: &trigger}, nil
}

func (t *Triggers) List(req *serverstate.TriggersListReq) (*serverstate.TriggersListResp, *serverstate.ErrorResp) {
	t.s.triggersLock.RLock()
	defer t.s.triggersLock.RUnlock()

	resp := serverstate.TriggersListResp{}

	// Use cache if enabled
	if t.s.enableCache {
		for _, trigger := range t.s.triggersCache {
			if req.Namespace == "*" || trigger.Namespace == req.Namespace {
				resp.Triggers = append(resp.Triggers, trigger.Stub())
			}
		}
		return &resp, nil
	}

	// List from Nomad Variables
	var prefix string
	if req.Namespace == "*" {
		prefix = triggersPathPrefix + "/"
	} else {
		prefix = fmt.Sprintf("%s/%s/", triggersPathPrefix, req.Namespace)
	}

	vars, err := t.s.listVariablesByPrefix(prefix)
	if err != nil && !isNotFoundError(err) {
		t.s.logger.Error("failed to list triggers from Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to list triggers: %w", err), 500)
	}

	for _, varMeta := range vars {
		v, err := t.s.getVariable(varMeta.Path)
		if err != nil {
			t.s.logger.Warn("failed to get trigger", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		var trigger state.Trigger
		if err := decodeFromVariable(v, &trigger); err != nil {
			t.s.logger.Warn("failed to decode trigger", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		resp.Triggers = append(resp.Triggers, trigger.Stub())
	}

	return &resp, nil
}

// loadTriggersCache loads all triggers into the cache
func (s *State) loadTriggersCache() error {
	s.triggersLock.Lock()
	defer s.triggersLock.Unlock()

	vars, err := s.listVariablesByPrefix(triggersPathPrefix + "/")
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return err
	}

	for _, varMeta := range vars {
		v, err := s.getVariable(varMeta.Path)
		if err != nil {
			s.logger.Warn("failed to get trigger for cache", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		var trigger state.Trigger
		if err := decodeFromVariable(v, &trigger); err != nil {
			s.logger.Warn("failed to decode trigger for cache", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		k := triggerCompositeKey{id: trigger.ID, namespace: trigger.Namespace}
		s.triggersCache[k] = &trigger
	}

	s.logger.Debug("loaded triggers into cache", zap.Int("count", len(s.triggersCache)))
	return nil
}
