package coordinator

import (
	"fmt"

	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/coordinator/inline"
	serverstate "github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/context"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/hcl"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

func (c *Coordinator) triggerInlineFlow(runID ulid.ULID, flow *state.Flow, trigger string, vars map[string]any) error {

	evalCtx, err := hcl.GenerateEvalContext(vars)
	if err != nil {
		return fmt.Errorf("failed to create HCL eval context: %w", err)
	}

	inlineReq := inline.InlineRunnerReq{
		Client:   c.nomadClient,
		DataDir:  c.dataDir,
		Logger:   c.logger.With(zap.String("flow_id", flow.ID)).With(zap.String("run_id", runID.String())),
		RunID:    runID,
		Flow:     flow,
		EvalCtx:  evalCtx,
		Vars:     vars,
		RPRCAddr: c.rpcAddr,
	}

	inlineRunner, err := inline.NewRunner(&inlineReq)
	if err != nil {
		return fmt.Errorf("failed to create inline runner: %w", err)
	}

	ctx := context.New(runID, trigger, flow, vars)

	if _, err := c.state.Runs().Create(&serverstate.RunsCreateReq{Run: ctx.Run()}); err != nil {
		return fmt.Errorf("failed to create run state: %w", err)
	}

	if err := inlineRunner.Start(c.inlineStartCh); err != nil {
		return fmt.Errorf("failed to start inline runner: %w", err)
	}

	c.inlineRunnersLock.Lock()
	c.inlineRunners[state.RunNamespacedKey{ID: runID, Namespace: flow.Namespace}] = inlineRunner
	c.inlineRunnersLock.Unlock()

	return nil
}

func (c *Coordinator) monitorInlineStart() {

	c.logger.Info("starting inline start failure monitor")

	for {
		select {
		case id := <-c.inlineStartCh:
			go c.hanldeInlineStartFailure(id)
		case <-c.shutdownCh:
			return
		}
	}
}

func (c *Coordinator) hanldeInlineStartFailure(id *state.RunNamespacedKey) {

	stateResp, err := c.state.Runs().Get(&serverstate.RunsGetReq{ID: id.ID, Namespace: id.Namespace})
	if err != nil {
		c.logger.Error("failed to query state for inline start failure",
			zap.String("run_id", id.ID.String()),
			zap.String("namespace", id.Namespace),
			zap.Error(err),
		)
		return
	}

	stateResp.Run.MarkFailed()

	if _, err := c.state.Runs().Update(&serverstate.RunsUpdateReq{Run: stateResp.Run}); err != nil {
		c.logger.Error("failed to update state for inline start failure",
			zap.String("run_id", id.ID.String()),
			zap.String("namespace", id.Namespace),
			zap.Error(err),
		)
		return
	} else {
		c.logger.Info("updated run state to failed for inline start failure",
			zap.String("run_id", id.ID.String()),
			zap.String("namespace", id.Namespace),
			zap.Error(err),
		)
	}
}
