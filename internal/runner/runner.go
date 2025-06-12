package runner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/logger"
	"github.com/hashicorp-forge/nomad-pipeline/internal/runner/result"
	"github.com/hashicorp-forge/nomad-pipeline/internal/server/state"
)

type Controller struct {
	dataDir     string
	logger      hclog.Logger
	nomadClient *api.Client
	state       state.State
}

type ControllerConfig struct {
	Logger      hclog.Logger
	NomadClient *api.Client
	State       state.State
	DataDir     string
}

func NewController(cfg *ControllerConfig) *Controller {
	return &Controller{
		dataDir:     filepath.Join(cfg.DataDir, "runs"),
		logger:      cfg.Logger.Named(logger.ComponentNameRunner),
		nomadClient: cfg.NomadClient,
		state:       cfg.State,
	}
}

func (c *Controller) RunFlow(id string) (ulid.ULID, error) {

	stateResp, err := c.state.Flows().Get(&state.FlowsGetReq{ID: id})
	if err != nil {
		return ulid.ULID{}, fmt.Errorf("failed to get flow: %w", err)
	}

	runID := ulid.MustNew(ulid.Now(), nil)

	runDir := filepath.Join(c.dataDir, runID.String())

	if err := os.MkdirAll(runDir, 0755); err != nil {
		return ulid.ULID{}, fmt.Errorf("failed to create run directory: %w", err)
	}

	resultHandler, resErr := result.New(runID, c.logger, stateResp.Flow, c.state)
	if resErr != nil {
		return ulid.ULID{}, fmt.Errorf("failed to create result handler: %w", err)
	}

	go c.runFlow(runID, runDir, stateResp.Flow, resultHandler)

	return runID, nil
}

func (c *Controller) runFlow(id ulid.ULID, runDir string, flow *state.Flow, res *result.Result) {

	runLogger := c.logger.With("run_id", id.String(), "flow_id", flow.ID)

	runLogger.Info("executing run of flow")

	res.StartRun()

	endState := state.RunStatusSuccess

	for _, job := range flow.Jobs {

		runLogger.Info("executing flow job", "job_id", job.ID)
		res.StartJob(job.ID)

		var jobStatus string

		if err := c.runFlowJob(id, flow.ID, runDir, job, res); err != nil {
			endState = state.RunStatusFailed
			jobStatus = state.RunStatusFailed
			runLogger.Error("failed to run flow job", "job_id", job.ID, "error", err)
			break
		} else {
			jobStatus = state.RunStatusSuccess
			runLogger.Info("successfully ran flow job", "job_id", job.ID)
		}

		res.EndJob(job.ID, jobStatus)
	}

	res.EndRun(endState)
	runLogger.Info("flow run finished", "status", endState)
}

func (c *Controller) runFlowJob(id ulid.ULID, flowID, runDir string, job *state.FlowJob, res *result.Result) error {

	switch job.Type() {
	case state.FlowJobTypeSpecification:

		fileRun := jobSpecification{
			client:  c.nomadClient,
			logger:  c.logger.With("flow_job_id", job.ID).With("run_id", id.String()),
			runID:   id,
			flowID:  flowID,
			flowJob: job,
			res:     res,
		}
		return fileRun.runJobFile(job)

	case state.FlowJobTypeInline:

		inlineRun := jobInline{
			client:  c.nomadClient,
			dataDir: runDir,
			logger:  c.logger.With("flow_job_id", job.ID).With("run_id", id.String()),
			runID:   id,
			flowID:  flowID,
			flowJob: job,
			res:     res,
		}
		return inlineRun.run()
	default:
		return errors.New("failed to determine job type")
	}
}
