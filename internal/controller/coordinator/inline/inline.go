package inline

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/hcl"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type InlineRunnerReq struct {
	Client *api.Client
	Logger *zap.Logger

	DataDir  string
	RunID    ulid.ULID
	Flow     *state.Flow
	Vars     map[string]any
	EvalCtx  *hcl.EvalContext
	RPRCAddr string
}

type InlineRunner struct {
	cancel    chan struct{}
	req       *InlineRunnerReq
	jobSpec   *api.Job
	queryOpts *api.QueryOptions
}

func NewRunner(req *InlineRunnerReq) (*InlineRunner, error) {

	r := InlineRunner{
		cancel: make(chan struct{}),
		req:    req,
	}

	jobBuildReq := jobBuilderReq{
		runID:   req.RunID,
		flow:    req.Flow,
		vars:    req.Vars,
		evalCtx: req.EvalCtx,
		rpcAddr: req.RPRCAddr,
	}

	jobspec, err := newJobBuilder(&jobBuildReq).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build job spec: %w", err)
	}
	r.jobSpec = jobspec

	r.queryOpts = &api.QueryOptions{Namespace: *r.jobSpec.Namespace}

	if err := createDataDir(req.DataDir, req.RunID, req.Flow); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	return &r, nil
}

func (r *InlineRunner) Start(failCh chan *state.RunNamespacedKey) error {

	writeOpts := api.WriteOptions{Namespace: *r.jobSpec.Namespace}

	_, _, err := r.req.Client.Jobs().Register(r.jobSpec, &writeOpts)
	if err != nil {
		return fmt.Errorf("failed to register job: %w", err)
	}

	go func() {
		alloc, err := r.getAlloc(*r.jobSpec.ID)
		if err != nil {
			failCh <- &state.RunNamespacedKey{ID: r.req.RunID, Namespace: r.req.Flow.Namespace}
			return
		} else {
			r.req.Logger.Info("successfully started Nomad job",
				zap.String("nomad_job_id", *r.jobSpec.ID),
				zap.String("nomad_namespace", *r.jobSpec.Namespace),
				zap.String("nomad_alloc_id", alloc.ID),
			)
		}
	}()

	return nil
}

func (r *InlineRunner) getAlloc(jobID string) (*api.AllocationListStub, error) {

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.cancel:
			return nil, errors.New("cancelled")
		case <-ticker.C:
			allocs, _, err := r.req.Client.Jobs().Allocations(jobID, false, r.queryOpts)
			if err != nil {
				r.req.Logger.Error("failed to get job allocations", zap.String("nomad_job_id", jobID), zap.Error(err))
				continue
			}

			for _, alloc := range allocs {
				switch alloc.ClientStatus {
				case api.AllocClientStatusRunning:
					return alloc, nil
				case api.AllocClientStatusPending:
					continue
				case api.AllocClientStatusUnknown, api.AllocClientStatusFailed, api.AllocClientStatusLost:
					return nil, fmt.Errorf("allocation in terminal state")
				}
			}
		}
	}
}

func (r *InlineRunner) Cancel() error {
	r.req.Logger.Info("cancelling inline runner", zap.String("nomad_job_id", *r.jobSpec.ID))

	select {
	case r.cancel <- struct{}{}:
	default:
	}

	_, _, err := r.req.Client.Jobs().Deregister(*r.jobSpec.ID, false, nil)
	return err
}
