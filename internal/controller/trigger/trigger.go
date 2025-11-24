package trigger

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	serverstate "github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/trigger/git"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/trigger/schedule"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/logger"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type coordinatorFlowRunFunc func(flowID, namespace, trigger string, vars map[string]any) error

const (
	GitWebhookProviderName = "git-webhook"
	CronProviderName       = "cron"
)

type Handler struct {
	logger  *zap.Logger
	state   serverstate.State
	runFunc coordinatorFlowRunFunc

	scheduleTrigger *schedule.Trigger
	gitTrigger      *git.Trigger
}

func NewHandler(log *zap.Logger, state serverstate.State, runFunc coordinatorFlowRunFunc) *Handler {

	h := Handler{
		logger:  log.Named(logger.ComponentNameTrigger),
		state:   state,
		runFunc: runFunc,
	}

	h.scheduleTrigger = schedule.NewTrigger(
		&schedule.TriggerConfig{
			Logger: h.logger,
			State:  h.state,
			RunFn:  h.runFunc,
		},
	)

	h.gitTrigger = git.NewTrigger(
		&git.TriggerConfig{
			Logger: h.logger,
			RunFn:  h.runFunc,
		},
	)

	return &h
}

func (h *Handler) CreateTrigger(trigger *state.Trigger) error {
	switch trigger.Source.Provider {
	case GitWebhookProviderName:
		return nil
	case CronProviderName:
		return h.scheduleTrigger.CreateTrigger(trigger)
	default:
		return fmt.Errorf("unsupported trigger provider: %s", trigger.Source.Provider)
	}
}

func (h *Handler) DeleteTrigger(trigger *state.Trigger) error {
	switch trigger.Source.Provider {
	case GitWebhookProviderName:
		return nil
	case CronProviderName:
		return h.scheduleTrigger.DeleteTrigger(trigger)
	default:
		return fmt.Errorf("unsupported trigger provider: %s", trigger.Source.Provider)
	}
}

func (h *Handler) Start() error {
	return h.scheduleTrigger.Start()
}

func (h *Handler) Stop() error {
	h.scheduleTrigger.Stop()
	return nil
}

func (h *Handler) HandleTriggerWebhook(
	w http.ResponseWriter,
	r *http.Request,
	trigger *state.Trigger,
) {
	switch trigger.Source.Provider {
	case GitWebhookProviderName:
		h.gitTrigger.HandleWebhook(w, r, trigger)
	default:
		http.Error(w, "unsupported trigger provider", http.StatusBadRequest)
	}
}
