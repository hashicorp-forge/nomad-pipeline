package spec

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec2"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/context"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type SpecRunnerReq struct {
	Client *api.Client
	Logger *zap.Logger

	RunID    ulid.ULID
	Flow     *state.Flow
	Trigger  string
	UpdateCh chan *state.Run
	Vars     map[string]any
}

type SpecRunner struct {
	cancel     chan struct{}
	context    *context.Context
	inProgress atomic.Value
	req        *SpecRunnerReq
	queryOpts  *api.QueryOptions
}

func NewRunner(req *SpecRunnerReq) (*SpecRunner, error) {

	r := SpecRunner{
		cancel:    make(chan struct{}),
		context:   context.New(req.RunID, req.Trigger, req.Flow, req.Vars),
		req:       req,
		queryOpts: &api.QueryOptions{},
	}

	return &r, nil
}

func (s *SpecRunner) Start() error {

	_, err := s.req.Client.Status().Leader()
	if err != nil {
		return fmt.Errorf("failed to connect to Nomad cluster: %w", err)
	}

	go s.start()
	return nil
}

func (s *SpecRunner) Cancel() error {
	s.req.Logger.Info("cancelling spec runner")

	select {
	case s.cancel <- struct{}{}:
	default:
	}

	_, _, err := s.req.Client.Jobs().Deregister(s.inProgress.Load().(string), false, nil)
	return err
}

func (s *SpecRunner) start() {

	defer func() {
		close(s.req.UpdateCh)
	}()

	s.req.Logger.Info("starting spec flow run")
	s.context.StartRun()
	s.req.UpdateCh <- s.context.Run()

	var failed bool

	for _, job := range s.req.Flow.Specification {

		should := true

		if job.Condition != "" {
			eval, err := s.context.ParseBoolExpr(job.Condition)
			if err != nil {
				s.req.Logger.Error("failed to evaluate condition", zap.String("spec_id", job.ID), zap.Error(err))
				s.context.EndRun(state.RunStatusFailed)
				s.req.UpdateCh <- s.context.Run()
				return
			}
			should = eval
		}

		if !should || failed {
			s.req.Logger.Info("skipping spec due to condition evaluation", zap.String("spec_id", job.ID))
			s.context.EndSpecification(job.ID, state.RunStatusSkipped)
			s.req.UpdateCh <- s.context.Run()
			continue
		}

		if err := s.runSpec(job); err != nil {
			s.req.Logger.Error("specification run failed", zap.String("spec_id", job.ID), zap.Error(err))
			failed = true
		}
	}

	endState := state.RunStatusSuccess
	if failed {
		endState = state.RunStatusFailed
	}

	s.context.EndRun(endState)
	s.req.UpdateCh <- s.context.Run()
}

func (s *SpecRunner) runSpec(spec *state.SpecificationFlow) error {

	inputVars := []string{}

	varNS, ok := s.req.Vars["var"].(map[string]any)
	if !ok {
		varNS = s.req.Vars
	}

	for k, v := range spec.JobSpecification.Variables {

		val := varNS[v]
		if val == nil {
			return fmt.Errorf("variable %q not provided for spec %q", v, spec.ID)
		}

		inputVars = append(inputVars, fmt.Sprintf("%s=%v", k, val))
	}

	job, err := jobspec2.ParseWithConfig(&jobspec2.ParseConfig{
		ArgVars: inputVars,
		Body:    []byte(spec.JobSpecification.Raw),
	})
	if err != nil {
		return err
	}

	//
	if spec.JobSpecification.NameFormat != "" {
		name, err := s.context.ParseTemplateStringExpr(spec.JobSpecification.NameFormat)
		if err != nil {
			return fmt.Errorf("failed to parse job name format: %w", err)
		}
		job.Name = &name
		job.ID = &name
	}

	job.Canonicalize()

	s.context.StartSpecification(spec.ID, *job.Namespace, *job.ID)
	s.req.UpdateCh <- s.context.Run()

	defer func() {
		if err != nil {
			s.context.EndSpecification(spec.ID, state.RunStatusFailed)
		} else {
			s.context.EndSpecification(spec.ID, state.RunStatusSuccess)
		}
		s.req.UpdateCh <- s.context.Run()
	}()

	_, _, err = s.req.Client.Jobs().Register(job, nil)
	if err != nil {
		return err
	}

	jobID := *job.ID

	switch job.IsParameterized() {
	case true:
		var dispatchResp *api.JobDispatchResponse
		dispatchResp, _, err = s.req.Client.Jobs().DispatchOpts(&api.DispatchOptions{JobID: jobID}, nil)
		if err != nil {
			return err
		}

		jobID = dispatchResp.DispatchedJobID
	default:
	}

	s.inProgress.Store(jobID)
	return s.monitorJob(jobID)
}

func (s *SpecRunner) monitorJob(id string) error {

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.cancel:
			return errors.New("cancelled")
		case <-ticker.C:
			job, _, err := s.req.Client.Jobs().Info(id, s.queryOpts)
			if err != nil {
				return err
			}

			if *job.Status == "dead" {
				return s.collectAllocStatus(id)
			}
		}
	}
}

func (s *SpecRunner) collectAllocStatus(id string) error {

	allocs, _, err := s.req.Client.Jobs().Allocations(id, false, s.queryOpts)
	if err != nil {
		return err
	}

	var failedAllocs int

	for _, alloc := range allocs {
		if alloc.ClientStatus == api.AllocClientStatusComplete {
			continue
		}
		if alloc.ClientStatus == api.AllocClientStatusFailed && alloc.NextAllocation == "" {
			failedAllocs++
		}
	}

	if failedAllocs > 0 {
		return fmt.Errorf("%v allocations failed", failedAllocs)
	}
	return nil
}
