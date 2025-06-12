package dev

import (
	"errors"

	"github.com/hashicorp-forge/nomad-pipeline/internal/server/state"
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

	_, ok := p.s.flows[req.Flow.ID]
	if ok {
		return nil, state.NewErrorResp(errors.New("flow already exists"), 409)
	}

	p.s.flows[req.Flow.ID] = req.Flow
	return &state.FlowsCreateResp{Flow: req.Flow}, nil
}

func (p *Flows) Delete(req *state.FlowsDeleteReq) (*state.FlowsDeleteResp, *state.ErrorResp) {
	p.s.flowsLock.Lock()
	defer p.s.flowsLock.Unlock()

	if _, ok := p.s.flows[req.ID]; !ok {
		return nil, state.NewErrorResp(errors.New("flow not found"), 404)
	} else {
		delete(p.s.flows, req.ID)
		return &state.FlowsDeleteResp{}, nil
	}
}

func (p *Flows) Get(req *state.FlowsGetReq) (*state.FlowsGetResp, *state.ErrorResp) {
	p.s.flowsLock.RLock()
	defer p.s.flowsLock.RUnlock()

	if flow, ok := p.s.flows[req.ID]; !ok {
		return nil, state.NewErrorResp(errors.New("flow not found"), 404)
	} else {
		return &state.FlowsGetResp{Flow: flow}, nil
	}
}

func (p *Flows) List(_ *state.FlowsListReq) (*state.FlowsListResp, *state.ErrorResp) {
	p.s.flowsLock.RLock()
	defer p.s.flowsLock.RUnlock()

	resp := state.FlowsListResp{}

	for _, flow := range p.s.flows {
		resp.Flows = append(resp.Flows, flow.Stub())
	}

	return &resp, nil
}
