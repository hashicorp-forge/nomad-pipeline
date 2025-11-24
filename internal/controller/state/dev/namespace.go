package dev

import (
	"errors"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
)

func (s *State) Namespaces() state.Namespaces {
	return &Namespaces{s: s}
}

type Namespaces struct {
	s *State
}

func (p *Namespaces) Create(req *state.NamespacesCreateReq) (*state.NamespacesCreateResp, *state.ErrorResp) {
	p.s.namesapcesLock.Lock()
	defer p.s.namesapcesLock.Unlock()

	_, ok := p.s.namespaces[req.Namespace.ID]
	if ok {
		return nil, state.NewErrorResp(errors.New("namespace already exists"), 409)
	}

	p.s.namespaces[req.Namespace.ID] = req.Namespace
	return &state.NamespacesCreateResp{}, nil
}

func (p *Namespaces) Delete(req *state.NamespacesDeleteReq) (*state.NamespacesDeleteResp, *state.ErrorResp) {

	// Check if any flows are using the namespace. There is no need to check the
	// targets as they depend on flows.
	p.s.flowsLock.Lock()
	for _, flow := range p.s.flows {
		if flow.Namespace == req.Name {
			p.s.flowsLock.Unlock()
			return nil, state.NewErrorResp(errors.New("cannot delete in-use namespace"), 409)
		}
	}
	p.s.flowsLock.Unlock()

	p.s.namesapcesLock.Lock()
	defer p.s.namesapcesLock.Unlock()

	if _, ok := p.s.namespaces[req.Name]; !ok {
		return nil, state.NewErrorResp(errors.New("namespace not found"), 404)
	} else {
		delete(p.s.namespaces, req.Name)
		return &state.NamespacesDeleteResp{}, nil
	}
}

func (p *Namespaces) Get(req *state.NamespacesGetReq) (*state.NamespacesGetResp, *state.ErrorResp) {
	p.s.namesapcesLock.RLock()
	defer p.s.namesapcesLock.RUnlock()

	if ns, ok := p.s.namespaces[req.Name]; !ok {
		return nil, state.NewErrorResp(errors.New("namespace not found"), 404)
	} else {
		return &state.NamespacesGetResp{Namespace: ns}, nil
	}
}

func (p *Namespaces) List(_ *state.NamespacesListReq) (*state.NamespacesListResp, *state.ErrorResp) {
	p.s.namesapcesLock.RLock()
	defer p.s.namesapcesLock.RUnlock()

	resp := state.NamespacesListResp{}

	for _, ns := range p.s.namespaces {
		resp.Namespaces = append(resp.Namespaces, ns.Stub())
	}

	return &resp, nil
}
