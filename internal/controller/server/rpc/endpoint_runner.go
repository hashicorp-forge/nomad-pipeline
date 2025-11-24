package rpc

import (
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/coordinator"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	intrpc "github.com/hashicorp-forge/nomad-pipeline/internal/pkg/rpc"
)

type RunnerEndpoint struct {
	coordinator *coordinator.Coordinator
	state       state.State
}

func (r *RunnerEndpoint) JobUpdate(
	req *intrpc.RunnerJobUpdateReq,
	reply *intrpc.RunnerJobUpdateResp,
) error {

	if err := req.Validate(); err != nil {
		return err
	}

	if _, err := r.state.Runs().Update(&state.RunsUpdateReq{Run: req.Run}); err != nil {
		return err
	}
	return nil
}

// JobLogsBatch receives a batch of log lines from a runner and writes them to disk
func (r *RunnerEndpoint) JobLogsBatch(
	req *intrpc.RunnerLogsBatchReq,
	reply *intrpc.RunnerLogsBatchResp,
) error {

	if err := req.Validate(); err != nil {
		return err
	}

	return r.coordinator.WriteLogsBatch(
		req.Namespace,
		req.RunID,
		req.StepID,
		req.Type,
		req.Logs,
	)
}
