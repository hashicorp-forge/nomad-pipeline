package dev

import (
	"errors"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
)

func (s *State) Triggers() state.Triggers {
	return &Triggers{s: s}
}

type Triggers struct {
	s *State
}

func (t *Triggers) Create(req *state.TriggersCreateReq) (*state.TriggersCreateResp, *state.ErrorResp) {
	t.s.triggersLock.Lock()
	defer t.s.triggersLock.Unlock()

	key := triggerCompositeKey{id: req.Trigger.ID, namesapce: req.Trigger.Namespace}

	_, ok := t.s.triggers[key]
	if ok {
		return nil, state.NewErrorResp(errors.New("trigger already exists"), 409)
	}

	t.s.triggers[key] = req.Trigger
	return &state.TriggersCreateResp{}, nil
}

func (t *Triggers) Delete(req *state.TriggersDeleteReq) (*state.TriggersDeleteResp, *state.ErrorResp) {
	t.s.triggersLock.Lock()
	defer t.s.triggersLock.Unlock()

	key := triggerCompositeKey{id: req.ID, namesapce: req.Namespace}

	if _, ok := t.s.triggers[key]; !ok {
		return nil, state.NewErrorResp(errors.New("trigger not found"), 404)
	}

	delete(t.s.triggers, key)
	return &state.TriggersDeleteResp{}, nil
}

func (t *Triggers) Get(req *state.TriggersGetReq) (*state.TriggersGetResp, *state.ErrorResp) {
	t.s.triggersLock.RLock()
	defer t.s.triggersLock.RUnlock()

	key := triggerCompositeKey{id: req.ID, namesapce: req.Namespace}

	if trigger, ok := t.s.triggers[key]; !ok {
		return nil, state.NewErrorResp(errors.New("trigger not found"), 404)
	} else {
		return &state.TriggersGetResp{Trigger: trigger}, nil
	}
}

func (t *Triggers) List(req *state.TriggersListReq) (*state.TriggersListResp, *state.ErrorResp) {
	t.s.triggersLock.RLock()
	defer t.s.triggersLock.RUnlock()

	resp := state.TriggersListResp{}

	for _, trigger := range t.s.triggers {
		if req.Namespace == "*" || trigger.Namespace == req.Namespace {
			resp.Triggers = append(resp.Triggers, trigger.Stub())
		}
	}

	return &resp, nil
}
