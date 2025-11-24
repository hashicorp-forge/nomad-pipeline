package job

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"os"

	"go.uber.org/zap"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/context"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/host"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/logger"
	sharedrpc "github.com/hashicorp-forge/nomad-pipeline/internal/pkg/rpc"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type Runner struct {
	cfg       *host.RunConfig
	logger    *zap.Logger
	context   *context.Context
	rpcClient *rpc.Client
}

func NewRunner(path string) (*Runner, error) {

	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cfg host.RunConfig

	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode file: %w", err)
	}

	client, err := rpc.Dial("tcp", cfg.ControllerRPC)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}

	zapLogger, err := logger.NewZap(logger.DefaultRunnerConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create zap logger: %w", err)
	}

	return &Runner{
		cfg: &cfg,
		logger: zapLogger.With(
			zap.String("job_id", cfg.JobID),
			zap.String("flow_id", cfg.Flow.ID),
			zap.String("run_id", cfg.ID.String()),
			zap.String("namespace", cfg.Namespace),
		),
		context: context.New(
			cfg.ID,
			cfg.Flow.ID,
			cfg.Flow,
			cfg.Variables,
		),
		rpcClient: client,
	}, nil
}

func (r *Runner) Run() error {

	r.startJob()

	var failed bool

	for _, step := range r.cfg.JobSteps {

		should := true

		if step.Condition != "" {
			eval, err := r.context.ParseBoolExpr(step.Condition)
			if err != nil {
				return fmt.Errorf("failed to evaluate condition for step %s: %w", step.ID, err)
			}
			should = eval
		}

		if !should || failed {
			r.logger.Info("skipping step due to condition evaluation", zap.String("step_id", step.ID))
			r.context.EndInlineStep(step.ID, state.RunStatusSkipped, -1)
			r.sendUpdateRPC()
			continue
		}

		sr := &stepRunner{
			cfg:       r.cfg,
			context:   r.context,
			logger:    r.logger,
			rpcClient: r.rpcClient,
		}

		stepResult, err := sr.executeStepRun(step)
		if err != nil {
			return fmt.Errorf("failed to execute step: %w", err)
		}

		r.context.EndInlineStep(step.ID, stepResult.Status, stepResult.ExitCode)
		r.sendUpdateRPC()

		if stepResult.Status == state.RunStatusFailed {
			failed = true
		}
	}

	endState := state.RunStatusSuccess
	if failed {
		endState = state.RunStatusFailed
	}

	r.endJob(endState)

	return nil
}

func (r *Runner) startJob() {
	r.logger.Info("starting flow job")
	r.context.StartRun()
	r.sendUpdateRPC()
}

func (r *Runner) endJob(status string) {
	r.logger.Info("ending flow job", zap.String("status", status))
	r.context.EndRun(status)
	r.sendUpdateRPC()
}

func (r *Runner) sendUpdateRPC() {
	req := sharedrpc.RunnerJobUpdateReq{JobID: r.cfg.JobID, Run: r.context.Run()}
	err := r.rpcClient.Call(sharedrpc.RunnerJobUpdateMethodName, req, nil)

	if err != nil {
		r.logger.Error("failed to send job update via RPC", zap.Error(err))
	} else {
		r.logger.Debug("successfully sent job update via RPC")
	}
}
