package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/coordinator"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	sharedstate "github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type triggersEndpoint struct {
	coordinator *coordinator.Coordinator
	state       state.State
}

func (t triggersEndpoint) routes() chi.Router {
	router := chi.NewRouter()

	router.Route("/", func(r chi.Router) {
		r.Use(namespaceCheckMiddleware(t.state))
		r.Post("/", t.create)
		r.Get("/", t.list)
	})

	router.Route("/{id}", func(r chi.Router) {
		r.Use(namespaceWildcardRejectMiddleware())
		r.Use(namespaceCheckMiddleware(t.state))
		r.Use(t.context)
		r.Delete("/", t.delete)
		r.Get("/", t.get)
		r.Post("/webhooks", t.webhooks)
	})

	return router
}

type TriggerCreateReq struct {
	Trigger *sharedstate.Trigger `json:"trigger"`
}

type TriggerCreateResp struct {
	Trigger              *sharedstate.Trigger `json:"trigger"`
	internalResponseMeta `json:"-"`
}

func (t triggersEndpoint) create(w http.ResponseWriter, r *http.Request) {

	var req TriggerCreateReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpWriteResponseError(w, NewResponseError(fmt.Errorf("failed to decode object: %w", err), 400))
		return
	}

	// Perform the static validation which is cheap and does not require state
	// access.
	if err := req.Trigger.Validate(); err != nil {
		httpWriteResponseError(w, NewResponseError(err, http.StatusBadRequest))
	}

	// Check that the referenced flow exists in state before creating the
	// trigger to avoid dangling references.
	if _, err := t.state.Flows().Get(
		&state.FlowsGetReq{
			ID:        req.Trigger.Flow,
			Namespace: req.Trigger.Namespace,
		},
	); err != nil {
		httpWriteResponseError(w, NewResponseError(err.Err(), err.StatusCode()))
		return
	}

	if _, err := t.state.Triggers().Create(
		&state.TriggersCreateReq{Trigger: req.Trigger},
	); err != nil {
		httpWriteResponseError(w, NewResponseError(err.Err(), err.StatusCode()))
		return
	}

	// Add trigger to coordinator
	if err := t.coordinator.CreateTrigger(req.Trigger); err != nil {
		_, delErr := t.state.Triggers().Delete(&state.TriggersDeleteReq{
			ID:        req.Trigger.ID,
			Namespace: req.Trigger.Namespace,
		})
		if delErr != nil {
			httpWriteResponseError(w, NewResponseError(
				fmt.Errorf("failed to schedule trigger: %w; additionally failed to rollback trigger creation: %v",
					err, delErr),
				http.StatusInternalServerError,
			))
			return
		}

		httpWriteResponseError(w, NewResponseError(err, http.StatusInternalServerError))
		return
	}

	resp := TriggerCreateResp{
		Trigger:              req.Trigger,
		internalResponseMeta: newInternalResponseMeta(http.StatusCreated),
	}
	httpWriteResponse(w, &resp)
}

type TriggerDeleteResp struct {
	internalResponseMeta `json:"-"`
}

func (t triggersEndpoint) delete(w http.ResponseWriter, r *http.Request) {

	triggerID := r.Context().Value("id").(string)
	namespace := getNamespaceParam(r)

	getResp, err := t.state.Triggers().Get(&state.TriggersGetReq{
		ID:        triggerID,
		Namespace: namespace,
	})
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
		return
	}

	_, err = t.state.Triggers().Delete(&state.TriggersDeleteReq{
		ID:        triggerID,
		Namespace: namespace,
	})
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
		return
	}

	if err := t.coordinator.TriggerDelete(getResp.Trigger); err != nil {
		httpWriteResponseError(
			w,
			NewResponseError(
				fmt.Errorf("trigger deleted but failed to unschedule: %w", err),
				http.StatusInternalServerError),
		)
		return
	}

	resp := FlowDeleteResp{
		internalResponseMeta: newInternalResponseMeta(http.StatusOK),
	}
	httpWriteResponse(w, &resp)
}

type TriggerGetReq struct {
	ID string `json:"id"`
}

type TriggerGetResp struct {
	Trigger              *sharedstate.Trigger `json:"trigger"`
	internalResponseMeta `json:"-"`
}

func (t triggersEndpoint) get(w http.ResponseWriter, r *http.Request) {

	stateResp, err := t.state.Triggers().Get(
		&state.TriggersGetReq{
			ID:        r.Context().Value("id").(string),
			Namespace: getNamespaceParam(r),
		},
	)
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
	} else {
		resp := TriggerGetResp{
			Trigger:              stateResp.Trigger,
			internalResponseMeta: newInternalResponseMeta(http.StatusOK),
		}
		httpWriteResponse(w, &resp)
	}
}

type TriggerListResp struct {
	Triggers             []*sharedstate.TriggerStub `json:"triggers"`
	internalResponseMeta `json:"-"`
}

func (t triggersEndpoint) list(w http.ResponseWriter, r *http.Request) {
	stateResp, err := t.state.Triggers().List(&state.TriggersListReq{
		Namespace: getNamespaceParam(r),
	})
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
	} else {
		resp := TriggerListResp{
			Triggers:             stateResp.Triggers,
			internalResponseMeta: newInternalResponseMeta(http.StatusOK),
		}
		httpWriteResponse(w, &resp)
	}
}

func (t triggersEndpoint) webhooks(w http.ResponseWriter, r *http.Request) {

	id := r.Context().Value("id").(string)
	ns := getNamespaceParam(r)

	stateResp, err := t.state.Triggers().Get(&state.TriggersGetReq{
		ID:        id,
		Namespace: ns,
	})
	if err != nil {
		httpWriteResponseError(w, NewResponseError(err.Err(), err.StatusCode()))
		return
	}

	if stateResp.Trigger.Source.Provider != "git-webhook" {
		httpWriteResponseError(w, NewResponseError(
			fmt.Errorf("trigger %s is not configured for git webhooks", stateResp.Trigger.ID),
			http.StatusBadRequest))
		return
	}

	t.coordinator.HandleWebhook(w, r, stateResp.Trigger)
}

func (t triggersEndpoint) context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var id string

		if id = chi.URLParam(r, "id"); id == "" {
			httpWriteResponseError(w, errors.New("trigger not found"))
			return
		}

		ctx := context.WithValue(r.Context(), "id", id) //nolint:staticcheck
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
