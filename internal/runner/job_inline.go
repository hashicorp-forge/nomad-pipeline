package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/runner/jobspec"
	"github.com/hashicorp-forge/nomad-pipeline/internal/runner/result"
	"github.com/hashicorp-forge/nomad-pipeline/internal/server/state"
)

type jobInline struct {
	client  *api.Client
	logger  hclog.Logger
	dataDir string

	runID   ulid.ULID
	flowID  string
	flowJob *state.FlowJob

	queryOpts *api.QueryOptions
	res       *result.Result
}

func (j *jobInline) run() error {

	nomadSpec := jobspec.BuildStepJob(j.runID, j.flowID, j.flowJob)

	j.queryOpts = &api.QueryOptions{Namespace: *nomadSpec.Namespace}

	_, _, err := j.client.Jobs().Register(nomadSpec, nil)
	if err != nil {
		return fmt.Errorf("failed to register job: %w", err)
	}

	alloc, err := j.getAlloc(context.Background(), *nomadSpec.ID)
	if err != nil {
		return fmt.Errorf("failed to get alloc: %w", err)
	}

	j.logger.Info("successfully started Nomad job",
		"nomad_job_id", *nomadSpec.ID, "nomad_alloc_id", alloc.ID)

	defer func() {
		_, _, err := j.client.Jobs().DeregisterOpts(*nomadSpec.ID, nil, &api.WriteOptions{Namespace: *nomadSpec.Namespace})
		if err != nil {
			j.logger.Error("failed to stop job", "nomad_job_id", *nomadSpec.ID, "error", err)
		} else {
			j.logger.Info("successfully stopped job", "nomad_job_id", *nomadSpec.ID)
		}
	}()

	for _, step := range j.flowJob.Steps {

		j.res.StartJobInlineStep(j.flowJob.ID, step.ID)

		stepResult, err := j.executeStep(step, alloc.ID)
		if err != nil {
			return fmt.Errorf("failed to execute step: %w", err)
		}

		j.res.EndJobInlineStep(j.flowJob.ID, step.ID, stepResult.Status, stepResult.ExitCode)

		if stepResult.Status == state.RunStatusFailed {
			return fmt.Errorf("step failed, exit code: %d", stepResult.ExitCode)
		}
	}

	return nil
}

func (j *jobInline) getAlloc(ctx context.Context, jobID string) (*api.AllocationListStub, error) {

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			allocs, _, err := j.client.Jobs().Allocations(jobID, false, j.queryOpts)
			if err != nil {
				j.logger.Error("failed to get job allocations", "nomad_job_id", jobID, "error", err)
				continue
			}

			for _, alloc := range allocs {
				if alloc.ClientStatus == "running" {
					return alloc, nil
				}
			}
		}
	}
}

func (j *jobInline) executeStep(step *state.Step, allocID string) (*state.RunJobInline, error) {

	j.logger.Info("executing flow job step", "flow_step_id", step.ID)

	if err := os.MkdirAll(filepath.Join(j.dataDir, j.flowJob.ID, step.ID), 0755); err != nil {
		return nil, err
	}

	stdoutFile, err := createLogFile(filepath.Join(j.dataDir, j.flowJob.ID, step.ID, "stdout.log"))
	if err != nil {
		return nil, err
	}
	defer func(stdoutFile *os.File) { _ = stdoutFile.Close() }(stdoutFile)

	stderrFile, err := createLogFile(filepath.Join(j.dataDir, j.flowJob.ID, step.ID, "stderr.log"))
	if err != nil {
		return nil, err
	}
	defer func(stderrFile *os.File) { _ = stderrFile.Close() }(stderrFile)

	exitCode, err := j.client.Allocations().Exec(
		context.Background(),
		&api.Allocation{ID: allocID},
		j.flowJob.ID,
		false,
		[]string{
			"bash",
			"-c",
			fmt.Sprintf("cd local/ && /bin/bash %s", step.ID),
		},
		os.Stdin,
		stdoutFile,
		stderrFile,
		nil,
		j.queryOpts,
	)
	if err != nil {
		return nil, err
	}

	j.logger.Info("execution of flow job step finished",
		"flow_step_id", step.ID, "exit_code", exitCode)

	res := state.RunJobInline{ID: step.ID, ExitCode: exitCode}

	if exitCode != 0 {
		res.Status = state.RunStatusFailed
	} else {
		res.Status = state.RunStatusSuccess
	}

	return &res, nil
}

func createLogFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
}
