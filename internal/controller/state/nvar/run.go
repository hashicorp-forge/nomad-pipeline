package nvar

import (
	"errors"
	"fmt"

	"go.uber.org/zap"

	serverstate "github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type Runs struct {
	s *State
}

func (r *Runs) Create(req *serverstate.RunsCreateReq) (*serverstate.RunsCreateResp, *serverstate.ErrorResp) {
	r.s.runsLock.Lock()
	defer r.s.runsLock.Unlock()

	k := runCompositeKey{id: req.Run.ID, namespace: req.Run.Namespace}

	// Check if run already exists
	if r.s.enableCache {
		if _, ok := r.s.runsCache[k]; ok {
			return nil, serverstate.NewErrorResp(errors.New("run already exists"), 409)
		}
	} else {
		_, err := r.s.getVariable(runVarPath(req.Run.Namespace, req.Run.ID))
		if err == nil {
			return nil, serverstate.NewErrorResp(errors.New("run already exists"), 409)
		}
	}

	// Create the variable
	v, err := encodeToVariable(runVarPath(req.Run.Namespace, req.Run.ID), req.Run)
	if err != nil {
		r.s.logger.Error("failed to encode run", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to encode run: %w", err), 500)
	}

	if err := r.s.putVariable(v); err != nil {
		r.s.logger.Error("failed to store run in Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to store run: %w", err), 500)
	}

	// Update cache
	if r.s.enableCache {
		r.s.runsCache[k] = req.Run
	}

	r.s.logger.Debug("run created",
		zap.String("namespace", req.Run.Namespace),
		zap.String("run_id", req.Run.ID.String()))

	return &serverstate.RunsCreateResp{}, nil
}

func (r *Runs) Delete(req *serverstate.RunsDeleteReq) (*serverstate.RunsDeleteResp, *serverstate.ErrorResp) {
	r.s.runsLock.Lock()
	defer r.s.runsLock.Unlock()

	k := runCompositeKey{id: req.ID, namespace: req.Namespace}

	// Check if run exists
	if r.s.enableCache {
		if _, ok := r.s.runsCache[k]; !ok {
			return nil, serverstate.NewErrorResp(errors.New("run not found"), 404)
		}
	} else {
		_, err := r.s.getVariable(runVarPath(req.Namespace, req.ID))
		if err != nil {
			return nil, serverstate.NewErrorResp(errors.New("run not found"), 404)
		}
	}

	// Delete from Nomad Variables
	if err := r.s.deleteVariable(runVarPath(req.Namespace, req.ID)); err != nil {
		r.s.logger.Error("failed to delete run from Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to delete run: %w", err), 500)
	}

	// Update cache
	if r.s.enableCache {
		delete(r.s.runsCache, k)
	}

	r.s.logger.Debug("run deleted",
		zap.String("namespace", req.Namespace),
		zap.String("run_id", req.ID.String()))

	return &serverstate.RunsDeleteResp{}, nil
}

func (r *Runs) Get(req *serverstate.RunsGetReq) (*serverstate.RunsGetResp, *serverstate.ErrorResp) {
	r.s.runsLock.RLock()
	defer r.s.runsLock.RUnlock()

	k := runCompositeKey{id: req.ID, namespace: req.Namespace}

	// Check cache first
	if r.s.enableCache {
		if run, ok := r.s.runsCache[k]; ok {
			return &serverstate.RunsGetResp{Run: run.Copy()}, nil
		}
		return nil, serverstate.NewErrorResp(errors.New("run not found"), 404)
	}

	// Read from Nomad Variables
	v, err := r.s.getVariable(runVarPath(req.Namespace, req.ID))
	if err != nil {
		if isNotFoundError(err) {
			return nil, serverstate.NewErrorResp(errors.New("run not found"), 404)
		}
		r.s.logger.Error("failed to get run from Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to get run: %w", err), 500)
	}

	var run state.Run
	if err := decodeFromVariable(v, &run); err != nil {
		r.s.logger.Error("failed to decode run", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to decode run: %w", err), 500)
	}

	return &serverstate.RunsGetResp{Run: run.Copy()}, nil
}

func (r *Runs) List(req *serverstate.RunsListReq) (*serverstate.RunsListResp, *serverstate.ErrorResp) {
	r.s.runsLock.RLock()
	defer r.s.runsLock.RUnlock()

	var runs []*state.RunStub

	// Use cache if enabled
	if r.s.enableCache {
		for _, run := range r.s.runsCache {
			if req.Namespace == "*" || run.Namespace == req.Namespace {
				runs = append(runs, run.Copy().Stub())
			}
		}
		return &serverstate.RunsListResp{Runs: runs}, nil
	}

	// List from Nomad Variables
	var prefix string
	if req.Namespace == "*" {
		prefix = runsPathPrefix + "/"
	} else {
		prefix = fmt.Sprintf("%s/%s/", runsPathPrefix, req.Namespace)
	}

	vars, err := r.s.listVariablesByPrefix(prefix)
	if err != nil && !isNotFoundError(err) {
		r.s.logger.Error("failed to list runs from Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to list runs: %w", err), 500)
	}

	for _, varMeta := range vars {
		v, err := r.s.getVariable(varMeta.Path)
		if err != nil {
			r.s.logger.Warn("failed to get run", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		var run state.Run
		if err := decodeFromVariable(v, &run); err != nil {
			r.s.logger.Warn("failed to decode run", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		runs = append(runs, run.Copy().Stub())
	}

	return &serverstate.RunsListResp{Runs: runs}, nil
}

func (r *Runs) Update(req *serverstate.RunsUpdateReq) (*serverstate.RunsUpdateResp, *serverstate.ErrorResp) {
	r.s.runsLock.Lock()
	defer r.s.runsLock.Unlock()

	k := runCompositeKey{id: req.Run.ID, namespace: req.Run.Namespace}

	// Get existing run to preserve certain fields
	var stateRun *state.Run
	if r.s.enableCache {
		if existingRun, ok := r.s.runsCache[k]; ok {
			stateRun = existingRun
		}
	} else {
		v, err := r.s.getVariable(runVarPath(req.Run.Namespace, req.Run.ID))
		if err == nil {
			var run state.Run
			if err := decodeFromVariable(v, &run); err == nil {
				stateRun = &run
			}
		}
	}

	// Preserve the original create time, variables, and trigger if the run
	// already exists which are generated by the controller.
	if stateRun != nil {
		req.Run.Variables = stateRun.Variables
		req.Run.CreateTime = stateRun.CreateTime
		req.Run.Trigger = stateRun.Trigger
	}

	// Update the variable
	v, err := encodeToVariable(runVarPath(req.Run.Namespace, req.Run.ID), req.Run)
	if err != nil {
		r.s.logger.Error("failed to encode run", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to encode run: %w", err), 500)
	}

	if err := r.s.putVariable(v); err != nil {
		r.s.logger.Error("failed to update run in Nomad Variables", zap.Error(err))
		return nil, serverstate.NewErrorResp(fmt.Errorf("failed to update run: %w", err), 500)
	}

	// Update cache
	if r.s.enableCache {
		r.s.runsCache[k] = req.Run
	}

	r.s.logger.Debug("run updated",
		zap.String("namespace", req.Run.Namespace),
		zap.String("run_id", req.Run.ID.String()))

	return &serverstate.RunsUpdateResp{}, nil
}

// loadRunsCache loads all runs into the cache
func (s *State) loadRunsCache() error {
	s.runsLock.Lock()
	defer s.runsLock.Unlock()

	vars, err := s.listVariablesByPrefix(runsPathPrefix + "/")
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return err
	}

	for _, varMeta := range vars {
		v, err := s.getVariable(varMeta.Path)
		if err != nil {
			s.logger.Warn("failed to get run for cache", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		var run state.Run
		if err := decodeFromVariable(v, &run); err != nil {
			s.logger.Warn("failed to decode run for cache", zap.String("path", varMeta.Path), zap.Error(err))
			continue
		}

		k := runCompositeKey{id: run.ID, namespace: run.Namespace}
		s.runsCache[k] = &run
	}

	s.logger.Debug("loaded runs into cache", zap.Int("count", len(s.runsCache)))
	return nil
}
