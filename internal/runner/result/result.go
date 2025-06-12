package result

import (
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/server/state"
)

type Result struct {
	state  state.State
	logger hclog.Logger
	r      *state.Run

	// horrible mappings, so we can use arrays for the result but track where
	// jobs and steps are.
	jobIDs map[string]int
	inline map[string]map[string]int
}

func New(id ulid.ULID, logger hclog.Logger, flow *state.Flow, stateImpl state.State) (*Result, error) {

	res := Result{
		state:  stateImpl,
		logger: logger.Named("result"),
		r: &state.Run{
			ID:     id,
			FlowID: flow.ID,
			Status: state.RunStatusPending,
			Jobs:   make([]*state.JobRun, len(flow.Jobs)),
		},
		jobIDs: make(map[string]int),
		inline: make(map[string]map[string]int),
	}

	for i, job := range flow.Jobs {

		res.r.Jobs[i] = &state.JobRun{ID: job.ID, Status: state.RunStatusPending}
		res.jobIDs[job.ID] = i

		switch job.Type() {
		case state.FlowJobTypeSpecification:
			res.r.Jobs[i].Specification = &state.RunJobSpecification{
				ID:     job.ID,
				Status: state.RunStatusPending,
			}
		case state.FlowJobTypeInline:
			for stepIdx, step := range job.Steps {
				res.r.Jobs[i].Inline = append(res.r.Jobs[i].Inline, &state.RunJobInline{
					ID:       step.ID,
					Status:   state.RunStatusPending,
					ExitCode: -1,
				})

				if _, ok := res.inline[job.ID]; !ok {
					res.inline[job.ID] = make(map[string]int)
				}
				res.inline[job.ID][step.ID] = stepIdx
			}
		}
	}

	_, err := res.state.Runs().Create(&state.RunsCreateReq{Run: res.r})
	if err != nil {
		return nil, err.Err()
	}

	return &res, nil
}

func (r *Result) StartRun() {
	r.r.StartTime = time.Now()
	r.r.Status = state.RunStatusRunning
	r.writeStateUpdate()
}

func (r *Result) EndRun(status string) {
	r.r.EndTime = time.Now()
	r.r.Status = status
	r.writeStateUpdate()
}

func (r *Result) StartJob(id string) {
	idx := r.jobIDs[id]
	r.r.Jobs[idx].StartTime = time.Now()
	r.r.Jobs[idx].Status = state.RunStatusRunning
	r.writeStateUpdate()
}

func (r *Result) EndJob(id, status string) {
	idx := r.jobIDs[id]
	r.r.Jobs[idx].EndTime = time.Now()
	r.r.Jobs[idx].Status = status
	r.writeStateUpdate()
}

func (r *Result) StartJobInlineStep(jobID, stepID string) {
	idx := r.jobIDs[jobID]
	r.r.Jobs[idx].Inline[r.inline[jobID][stepID]].Status = state.RunStatusRunning
	r.r.Jobs[idx].Inline[r.inline[jobID][stepID]].StartTime = time.Now()
	r.writeStateUpdate()
}

func (r *Result) EndJobInlineStep(jobID, stepID, status string, exitCode int) {
	idx := r.jobIDs[jobID]
	r.r.Jobs[idx].Inline[r.inline[jobID][stepID]].Status = status
	r.r.Jobs[idx].Inline[r.inline[jobID][stepID]].ExitCode = exitCode
	r.r.Jobs[idx].Inline[r.inline[jobID][stepID]].EndTime = time.Now()
	r.writeStateUpdate()
}

func (r *Result) StartJobSpecification(jobID, nomadID, nomadNS string) {
	idx := r.jobIDs[jobID]
	r.r.Jobs[idx].Specification.Status = state.RunStatusRunning
	r.r.Jobs[idx].Specification.StartTime = time.Now()
	r.r.Jobs[idx].Specification.NomadJobID = nomadID
	r.r.Jobs[idx].Specification.NomadNamespace = nomadNS
	r.writeStateUpdate()
}

func (r *Result) EndJobJobSpecification(jobID, status string) {
	idx := r.jobIDs[jobID]
	r.r.Jobs[idx].Specification.Status = status
	r.r.Jobs[idx].Specification.EndTime = time.Now()
	r.writeStateUpdate()
}

func (r *Result) writeStateUpdate() {
	if _, err := r.state.Runs().Update(&state.RunsUpdateReq{Run: r.r}); err != nil {
		r.logger.Error("failed to write run update to state", "error", err)
	}
}
