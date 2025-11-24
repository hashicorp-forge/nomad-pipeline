package job

import (
	stdcontext "context"
	"fmt"
	"net/rpc"
	"os"
	"os/exec"

	"go.uber.org/zap"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/context"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/host"
	sharedrpc "github.com/hashicorp-forge/nomad-pipeline/internal/pkg/rpc"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type stepRunner struct {
	cfg         *host.RunConfig
	context     *context.Context
	logger      *zap.Logger
	logHandlers []*LogHandler
	rpcClient   *rpc.Client
}

func (sr *stepRunner) executeStepRun(step *state.Step) (*state.InlineStep, error) {

	sr.logger.Info("executing flow job step",
		zap.String("flow_step_id", step.ID),
		zap.String("flow_step_path", "./local/"+sr.cfg.ID.String()+"/"+step.ID),
	)

	parsedExpr, err := sr.context.ParseTemplateStringExpr(step.Run)
	if err != nil {
		return nil, fmt.Errorf("could not process HCL expression for step run: %w", err)
	}

	if err := os.WriteFile("./local/"+sr.cfg.ID.String()+"/"+step.ID, []byte(parsedExpr), 0755); err != nil {
		return nil, fmt.Errorf("could not write step script: %w", err)
	}

	cmd := exec.Command("bash", "./"+step.ID)
	cmd.Dir = "./local/" + sr.cfg.ID.String()

	ctx := stdcontext.Background()
	defer ctx.Done()

	if err := sr.setupLogHandlers(cmd, step.ID); err != nil {
		return nil, fmt.Errorf("failed to setup log handlers: %w", err)
	}

	for _, logHandler := range sr.logHandlers {
		sr.logger.Debug("starting log handler", zap.String("log_type", logHandler.req.Type))
		go logHandler.Start(ctx)
	}

	sr.context.StartInlineStep(step.ID)
	req := sharedrpc.RunnerJobUpdateReq{JobID: sr.cfg.JobID, Run: sr.context.Run()}
	if err := sr.rpcClient.Call(sharedrpc.RunnerJobUpdateMethodName, req, nil); err != nil {
		sr.logger.Error("could not send job update RPC call", zap.Error(err))
	} else {
		sr.logger.Debug("sent job update RPC call for step start", zap.String("flow_step_id", step.ID))
	}

	if err := cmd.Run(); err != nil {
		fmt.Println("could not run command: ", err)
	}

	exitCode := cmd.ProcessState.ExitCode()

	sr.logger.Info("execution of flow job step finished",
		zap.String("flow_step_id", step.ID), zap.Int("exit_code", exitCode))

	res := state.InlineStep{ID: step.ID, ExitCode: exitCode}

	if exitCode != 0 {
		res.Status = state.RunStatusFailed
	} else {
		res.Status = state.RunStatusSuccess
	}

	return &res, nil
}

func (sr *stepRunner) setupLogHandlers(cmd *exec.Cmd, stepID string) error {

	stderrReq := LogHandlerReq{
		Namespace: sr.cfg.Namespace,
		RunID:     sr.cfg.ID.String(),
		StepID:    stepID,
		Type:      "stderr",
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("could not get stderr pipe: %w", err)
	}
	sr.logHandlers = append(sr.logHandlers, NewLogHandler(sr.logger, stderrPipe, sr.rpcClient, &stderrReq))

	stdoutReq := LogHandlerReq{
		RunID:     sr.cfg.ID.String(),
		Namespace: sr.cfg.Namespace,
		StepID:    stepID,
		Type:      "stdout",
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("could not get stdout pipe: %w", err)
	}
	sr.logHandlers = append(sr.logHandlers, NewLogHandler(sr.logger, stdoutPipe, sr.rpcClient, &stdoutReq))

	return nil
}
