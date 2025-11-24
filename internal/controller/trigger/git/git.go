package git

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/google/go-github/v79/github"
	"go.uber.org/zap"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/logger"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type Trigger struct {
	logger  *zap.Logger
	runFlow func(flowID, namespace, trigger string, vars map[string]any) error
}

type TriggerConfig struct {
	Logger *zap.Logger
	RunFn  func(flowID, namespace, trigger string, vars map[string]any) error
}

func NewTrigger(cfg *TriggerConfig) *Trigger {
	return &Trigger{
		logger:  cfg.Logger.Named(logger.ComponentNameTriggerGitWebhook),
		runFlow: cfg.RunFn,
	}
}

type webhookPayload struct {
	repo   string
	event  string
	branch string

	vars map[string]any
}

func (h *Trigger) HandleWebhook(w http.ResponseWriter, r *http.Request, trigger *state.Trigger) {

	// Decode the trigger config from the any field
	cfg, err := decodeTriggerConfig(trigger)
	if err != nil {
		h.logger.Error("failed to decode trigger config",
			zap.String("trigger_id", trigger.ID),
			zap.Error(err))
		http.Error(w, "failed to decode trigger config", http.StatusInternalServerError)
		return
	}

	var (
		payload *webhookPayload
	)

	switch cfg.Provider {
	case "github":
		payload, err = h.handleGitHubWebhook(r, cfg)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "unsupported git provider", http.StatusBadRequest)
		return
	}

	if !slices.Contains(cfg.Events, payload.event) {
		h.logger.Debug("event type not configured, ignoring",
			zap.String("event", payload.event),
			zap.Strings("configured_events", cfg.Events))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("event type not configured, ignored"))
		return
	}

	go func() {
		if err := h.runFlow(trigger.Flow, trigger.Namespace, trigger.ID, payload.vars); err != nil {
			h.logger.Error("failed to execute flow from webhook",
				zap.String("trigger_id", trigger.ID),
				zap.String("flow_id", trigger.Flow),
				zap.Error(err))
		}
	}()

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("webhook processed successfully"))
}

func (h *Trigger) handleGitHubWebhook(r *http.Request, cfg *triggerConfig) (*webhookPayload, error) {

	h.logger.Debug("processing GitHub webhook",
		zap.String("content-type", r.Header.Get("Content-Type")),
		zap.String("event-type", r.Header.Get("X-GitHub-Event")))

	payload, err := github.ValidatePayload(r, []byte(cfg.Secret))
	if err != nil {
		return nil, err
	}

	webHookType := github.WebHookType(r)

	event, err := github.ParseWebHook(webHookType, payload)
	if err != nil {
		return nil, err
	}

	resp := webhookPayload{}

	switch event := event.(type) {
	case *github.PushEvent:
		resp.repo = event.GetRepo().GetFullName()
		resp.event = webHookType
		resp.branch = event.GetRef()
		resp.vars = buildGithubVars(event)
	default:
		return nil, fmt.Errorf("unsupported GitHub event type: %s", webHookType)
	}

	return &resp, nil
}

func buildGithubVars(event *github.PushEvent) map[string]any {
	vars := map[string]any{
		"git_ref":        event.GetRef(),
		"git_sha":        event.GetAfter(),
		"git_before":     event.GetBefore(),
		"git_repository": event.GetRepo().GetFullName(),
		"git_pusher":     event.GetPusher().GetName(),
	}

	if event.GetRepo() != nil {
		vars["git_repo_name"] = event.GetRepo().GetName()
		vars["git_repo_owner"] = event.GetRepo().GetOwner().GetLogin()
		vars["git_repo_url"] = event.GetRepo().GetHTMLURL()
	}

	if event.GetHeadCommit() != nil {
		vars["git_commit_message"] = event.GetHeadCommit().GetMessage()
		vars["git_commit_author"] = event.GetHeadCommit().GetAuthor().GetName()
		vars["git_commit_url"] = event.GetHeadCommit().GetURL()
	}

	return map[string]any{"trigger": vars}
}
