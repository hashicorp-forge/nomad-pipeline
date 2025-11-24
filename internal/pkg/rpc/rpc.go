package rpc

import (
	"errors"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

const (
	RunnerJobUpdateMethodName = "Runner.JobUpdate"
	RunnerLogsBatchMethodName = "Runner.JobLogsBatch"
)

type RunnerJobUpdateReq struct {
	JobID string
	Run   *state.Run
}

func (r *RunnerJobUpdateReq) Validate() error {
	if r.JobID == "" {
		return errors.New("empty job ID")
	}
	if r.Run.Namespace == "" {
		return errors.New("empty namespace ID")
	}
	if r.Run == nil {
		return errors.New("empty run object")
	}
	return nil
}

type RunnerJobUpdateResp struct{}

type RunnerLogStreamResp struct{}

type RunnerLogsBatchReq struct {
	Namespace string   `json:"namespace"`
	RunID     string   `json:"run_id"`
	StepID    string   `json:"step_id"`
	Type      string   `json:"type"`
	Logs      []string `json:"logs"`
}

type RunnerLogsBatchResp struct{}

func (r *RunnerLogsBatchReq) Validate() error {
	if r.Namespace == "" {
		return errors.New("empty namespace")
	}
	if r.RunID == "" {
		return errors.New("empty run ID")
	}
	if r.StepID == "" {
		return errors.New("empty step ID")
	}
	if r.Type == "" {
		return errors.New("empty log type")
	}
	if r.Type != "stdout" && r.Type != "stderr" {
		return errors.New("log type must be 'stdout' or 'stderr'")
	}
	if len(r.Logs) == 0 {
		return errors.New("empty logs")
	}
	return nil
}
