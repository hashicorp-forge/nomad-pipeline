package coordinator

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/hashicorp/nomad/api"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/coordinator/inline"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/coordinator/spec"
	serverstate "github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/trigger"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/logger"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type Coordinator struct {
	dataDir     string
	rpcAddr     string
	logger      *zap.Logger
	nomadClient *api.Client
	state       serverstate.State

	inlineRunners     map[state.RunNamespacedKey]*inline.InlineRunner
	inlineRunnersLock sync.RWMutex

	//
	inlineStartCh chan *state.RunNamespacedKey

	specRunners     map[string]*spec.SpecRunner
	specRunnersLock sync.RWMutex

	//
	trigger *trigger.Handler

	//
	shutdownCh chan struct{}
}

type CoordinatorConfig struct {
	Logger      *zap.Logger
	NomadClient *api.Client
	State       serverstate.State
	DataDir     string
	RPCAddr     string
}

func New(cfg *CoordinatorConfig) *Coordinator {
	c := &Coordinator{
		dataDir:       filepath.Join(cfg.DataDir, "runs"),
		logger:        cfg.Logger.Named(logger.ComponentNameCoordinator),
		nomadClient:   cfg.NomadClient,
		state:         cfg.State,
		inlineRunners: make(map[state.RunNamespacedKey]*inline.InlineRunner),
		specRunners:   make(map[string]*spec.SpecRunner),
		inlineStartCh: make(chan *state.RunNamespacedKey, 10),
		rpcAddr:       cfg.RPCAddr,
		shutdownCh:    make(chan struct{}),
	}

	c.trigger = trigger.NewHandler(
		c.logger,
		c.state,
		c.runFlowFromTrigger,
	)

	return c
}

// Start starts the coordinator and its trigger coordinator
func (c *Coordinator) Start() error {
	if err := c.trigger.Start(); err != nil {
		return fmt.Errorf("failed to start trigger handler: %w", err)
	}

	go c.monitorInlineStart()

	return nil
}

// Stop gracefully stops the coordinator
func (c *Coordinator) Stop() {
	_ = c.trigger.Stop()
	close(c.shutdownCh)
}

func (c *Coordinator) RunFlow(
	id, namespace, trigger string,
	vars map[string]any,
) (ulid.ULID, error) {

	stateResp, stateErr := c.state.Flows().Get(&serverstate.FlowsGetReq{ID: id, Namespace: namespace})
	if stateErr != nil {
		return ulid.ULID{}, fmt.Errorf("failed to get flow: %w", stateErr)
	}

	runVars, err := generateVariablesMap(stateResp.Flow, vars)
	if err != nil {
		return ulid.ULID{}, err
	}

	runID := ulid.Make()

	switch stateResp.Flow.Type() {
	case state.FlowTypeInline:
		err = c.triggerInlineFlow(runID, stateResp.Flow, trigger, runVars)
	case state.FlowTypeSpecification:
		err = c.triggerSpecFlow(runID, stateResp.Flow, trigger, runVars)
	default:
		err = errors.New("failed to determine flow type")
	}

	if err != nil {
		return ulid.ULID{}, fmt.Errorf("failed to trigger flow: %w", err)
	}
	return runID, nil
}

// runFlowFromTrigger is called by the trigger coordinator to run a flow
func (c *Coordinator) runFlowFromTrigger(flowID, namespace, trigger string, vars map[string]any) error {
	_, err := c.RunFlow(flowID, namespace, trigger, vars)
	return err
}

func (c *Coordinator) CancelRun(id ulid.ULID, namespace string) error {

	resp, stateErr := c.state.Runs().Get(&serverstate.RunsGetReq{ID: id, Namespace: namespace})
	if stateErr != nil {
		return stateErr
	}

	var err error

	switch resp.Run.Type() {
	case "inline":
		err = c.cancelInlineRun(id, namespace)
	case "specification":
		err = c.cancelSpecRun(id)
	default:
		return errors.New("unknown run type")
	}

	if err != nil {
		return fmt.Errorf("failed to cancel run: %w", err)
	}

	resp.Run.MarkCancelled()

	_, stateErr = c.state.Runs().Update(&serverstate.RunsUpdateReq{Run: resp.Run})
	if stateErr != nil {
		return fmt.Errorf("failed to update run status: %w", stateErr)
	}

	return nil

}

func (c *Coordinator) cancelInlineRun(id ulid.ULID, namespace string) error {
	c.inlineRunnersLock.Lock()
	defer c.inlineRunnersLock.Unlock()

	inline, ok := c.inlineRunners[state.RunNamespacedKey{ID: id, Namespace: namespace}]
	if ok {
		return inline.Cancel()
	}

	return errors.New("inline runner not found")
}

func (c *Coordinator) cancelSpecRun(id ulid.ULID) error {
	c.specRunnersLock.Lock()
	defer c.specRunnersLock.Unlock()

	spec, ok := c.specRunners[id.String()]
	if ok {
		return spec.Cancel()
	}

	return errors.New("spec runner not found")
}

func (c *Coordinator) CreateTrigger(trigger *state.Trigger) error {

	c.logger.Debug(
		"adding trigger",
		zap.String("provider", trigger.Source.Provider),
		zap.String("trigger_id", trigger.ID),
	)

	if err := c.trigger.CreateTrigger(trigger); err != nil {
		c.logger.Error(
			"failed to add trigger",
			zap.String("provider", trigger.Source.Provider),
			zap.String("trigger_id", trigger.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to add trigger: %w", err)
	}

	c.logger.Info(
		"successfully added trigger",
		zap.String("provider", trigger.Source.Provider),
		zap.String("trigger_id", trigger.ID),
	)

	return nil
}

func (c *Coordinator) TriggerDelete(trigger *state.Trigger) error {

	c.logger.Debug(
		"deleting trigger",
		zap.String("provider", trigger.Source.Provider),
		zap.String("trigger_id", trigger.ID),
	)

	if err := c.trigger.DeleteTrigger(trigger); err != nil {
		c.logger.Error(
			"failed to delete trigger",
			zap.String("provider", trigger.Source.Provider),
			zap.String("trigger_id", trigger.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to remove trigger: %w", err)
	}

	c.logger.Info(
		"successfully deleted trigger",
		zap.String("provider", trigger.Source.Provider),
		zap.String("trigger_id", trigger.ID),
	)

	return nil
}

func (c *Coordinator) HandleWebhook(w http.ResponseWriter, r *http.Request, trigger *state.Trigger) {
	c.trigger.HandleTriggerWebhook(w, r, trigger)
}
