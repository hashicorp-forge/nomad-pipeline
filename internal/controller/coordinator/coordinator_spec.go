package coordinator

import (
	"fmt"

	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/coordinator/spec"
	serverstate "github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/context"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

func (c *Coordinator) triggerSpecFlow(runID ulid.ULID, flow *state.Flow, trigger string, vars map[string]any) error {

	ctx := context.New(runID, trigger, flow, vars)

	if _, err := c.state.Runs().Create(&serverstate.RunsCreateReq{Run: ctx.Run()}); err != nil {
		return fmt.Errorf("failed to create run state: %w", err)
	}

	specReq := spec.SpecRunnerReq{
		Client:   c.nomadClient,
		Logger:   c.logger.With(zap.String("flow_id", flow.ID)).With(zap.String("run_id", runID.String())),
		RunID:    runID,
		Flow:     flow,
		Trigger:  trigger,
		UpdateCh: make(chan *state.Run, 1),
		Vars:     vars,
	}

	specRunner, err := spec.NewRunner(&specReq)
	if err != nil {
		return fmt.Errorf("failed to create spec runner: %w", err)
	}

	if err := specRunner.Start(); err != nil {
		return fmt.Errorf("failed to start spec runner: %w", err)
	}

	c.specRunnersLock.Lock()
	c.specRunners[runID.String()] = specRunner
	c.specRunnersLock.Unlock()

	go c.monitorStatusUpdates(specReq.UpdateCh)

	return nil
}

func (c *Coordinator) monitorStatusUpdates(ch chan *state.Run) {
	for update := range ch {
		_, err := c.state.Runs().Update(&serverstate.RunsUpdateReq{Run: update})
		if err != nil {
			c.logger.Error("failed to update run status", zap.String("run_id", update.ID.String()), zap.Error(err))
			continue
		}
	}
}
