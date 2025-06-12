package dev

import (
	"errors"
	"github.com/hashicorp-forge/nomad-pipeline/internal/server/state"
)

func (s *State) Runs() state.Runs {
	return &Runs{s: s}
}

type Runs struct {
	s *State
}

func (r *Runs) Create(req *state.RunsCreateReq) (*state.RunsCreateResp, *state.ErrorResp) {
	r.s.runsLock.Lock()
	defer r.s.runsLock.Unlock()

	_, ok := r.s.runs[req.Run.ID]
	if ok {
		return nil, state.NewErrorResp(errors.New("flow already exists"), 409)
	}

	r.s.runs[req.Run.ID] = req.Run
	return &state.RunsCreateResp{}, nil
}

func (r *Runs) Delete(req *state.RunsDeleteReq) (*state.RunsDeleteResp, *state.ErrorResp) {
	r.s.runsLock.Lock()
	defer r.s.runsLock.Unlock()

	if _, ok := r.s.runs[req.ID]; !ok {
		return nil, state.NewErrorResp(errors.New("run not found"), 404)
	}

	delete(r.s.runs, req.ID)
	return &state.RunsDeleteResp{}, nil
}

func (r *Runs) Get(req *state.RunsGetReq) (*state.RunsGetResp, *state.ErrorResp) {
	r.s.runsLock.RLock()
	defer r.s.runsLock.RUnlock()

	if run, ok := r.s.runs[req.ID]; !ok {
		return nil, state.NewErrorResp(errors.New("run not found"), 404)
	} else {
		return &state.RunsGetResp{Run: run}, nil
	}
}

func (r *Runs) List(req *state.RunsListReq) (*state.RunsListResp, *state.ErrorResp) {
	r.s.flowsLock.RLock()
	defer r.s.flowsLock.RUnlock()

	var runs []*state.RunStub

	for _, run := range r.s.runs {
		runs = append(runs, run.Stub())
	}

	return &state.RunsListResp{Runs: runs}, nil
}

func (r *Runs) Update(req *state.RunsUpdateReq) (*state.RunsUpdateResp, *state.ErrorResp) {
	r.s.flowsLock.Lock()
	defer r.s.flowsLock.Unlock()

	r.s.runs[req.Run.ID] = req.Run
	return &state.RunsUpdateResp{}, nil
}
