package runner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec2"
	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/runner/jobspec"
	"github.com/hashicorp-forge/nomad-pipeline/internal/runner/result"
	"github.com/hashicorp-forge/nomad-pipeline/internal/server/state"
)

type jobSpecification struct {
	client *api.Client
	logger hclog.Logger

	runID   ulid.ULID
	flowID  string
	flowJob *state.FlowJob

	queryOpts *api.QueryOptions
	res       *result.Result
}

func (j *jobSpecification) runJobFile(job *state.FlowJob) error {

	nomadSpec, err := jobspec2.Parse("", strings.NewReader(job.JobSpecification.Data))
	if err != nil {
		return err
	}

	jobID := jobspec.GenerateJobID(j.runID, j.flowID, job.ID)
	nomadSpec.ID = &jobID
	nomadSpec.Name = &jobID

	j.queryOpts = &api.QueryOptions{Namespace: api.DefaultNamespace}
	if nomadSpec.Namespace != nil {
		j.queryOpts.Namespace = *nomadSpec.Namespace
	}

	j.res.StartJobSpecification(job.ID, jobID, j.queryOpts.Namespace)

	if err := j.runJob(context.Background(), nomadSpec); err != nil {
		j.res.EndJobJobSpecification(job.ID, state.RunStatusFailed)
		return err
	}

	j.res.EndJobJobSpecification(job.ID, state.RunStatusSuccess)
	return nil
}

func (j *jobSpecification) runJob(ctx context.Context, job *api.Job) error {

	_, _, err := j.client.Jobs().Register(job, nil)
	if err != nil {
		return err
	}

	jobID := *job.ID

	if job.IsParameterized() {

		opts := api.DispatchOptions{
			JobID: jobID,
		}

		dispatchResp, _, err := j.client.Jobs().DispatchOpts(&opts, nil)
		if err != nil {
			return err
		}

		jobID = dispatchResp.DispatchedJobID
	}

	return j.monitorJob(ctx, jobID)
}

func (j *jobSpecification) monitorJob(ctx context.Context, id string) error {

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			job, _, err := j.client.Jobs().Info(id, j.queryOpts)
			if err != nil {
				return err
			}

			if *job.Status == "dead" {
				return j.collectAllocStatus(id)
			}
		}
	}
}

func (j *jobSpecification) collectAllocStatus(id string) error {

	allocs, _, err := j.client.Jobs().Allocations(id, false, j.queryOpts)
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
		return fmt.Errorf("%v failed allocations found", failedAllocs)
	}
	return nil
}
