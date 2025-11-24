package dev

import (
	"errors"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
)

func (s *State) Flows() state.Flows {
	return &Flows{s: s}
}

type Flows struct {
	s *State
}

func (p *Flows) Create(req *state.FlowsCreateReq) (*state.FlowsCreateResp, *state.ErrorResp) {
	p.s.flowsLock.Lock()
	defer p.s.flowsLock.Unlock()

	k := flowCompositeKey{id: req.Flow.ID, namesapce: req.Flow.Namespace}

	_, ok := p.s.flows[k]
	if ok {
		return nil, state.NewErrorResp(errors.New("flow already exists"), 409)
	}

	p.s.flows[k] = req.Flow
	return &state.FlowsCreateResp{Flow: req.Flow}, nil
}

func (p *Flows) Delete(req *state.FlowsDeleteReq) (*state.FlowsDeleteResp, *state.ErrorResp) {

	// Check if any triggers are using the flow.
	p.s.triggersLock.Lock()
	for _, trigger := range p.s.triggers {
		if trigger.Flow == req.ID && trigger.Namespace == req.Namespace {
			p.s.triggersLock.Unlock()
			return nil, state.NewErrorResp(errors.New("cannot delete flow with linked trigger"), 409)
		}
	}
	p.s.triggersLock.Unlock()

	k := flowCompositeKey{id: req.ID, namesapce: req.Namespace}

	p.s.flowsLock.Lock()
	defer p.s.flowsLock.Unlock()

	if _, ok := p.s.flows[k]; !ok {
		return nil, state.NewErrorResp(errors.New("flow not found"), 404)
	} else {
		delete(p.s.flows, k)
		return &state.FlowsDeleteResp{}, nil
	}
}

func (p *Flows) Get(req *state.FlowsGetReq) (*state.FlowsGetResp, *state.ErrorResp) {
	p.s.flowsLock.RLock()
	defer p.s.flowsLock.RUnlock()

	if flow, ok := p.s.flows[flowCompositeKey{id: req.ID, namesapce: req.Namespace}]; !ok {
		return nil, state.NewErrorResp(errors.New("flow not found"), 404)
	} else {
		return &state.FlowsGetResp{Flow: flow}, nil
	}
}

func (p *Flows) List(req *state.FlowsListReq) (*state.FlowsListResp, *state.ErrorResp) {
	p.s.flowsLock.RLock()
	defer p.s.flowsLock.RUnlock()

	resp := state.FlowsListResp{}

	for _, flow := range p.s.flows {
		if req.Namespace == "*" || flow.Namespace == req.Namespace {
			resp.Flows = append(resp.Flows, flow.Stub())
		}
	}

	return &resp, nil
}
