package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/coordinator"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	sharedstate "github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type flowsEndpoint struct {
	runController *coordinator.Coordinator
	state         state.State
}

func (f flowsEndpoint) routes() chi.Router {
	router := chi.NewRouter()

	router.Route("/", func(r chi.Router) {
		r.Use(namespaceCheckMiddleware(f.state))
		r.Post("/", f.create)
		r.Get("/", f.list)
	})

	router.Route("/{id}", func(r chi.Router) {
		r.Use(namespaceWildcardRejectMiddleware())
		r.Use(namespaceCheckMiddleware(f.state))
		r.Use(f.context)
		r.Delete("/", f.delete)
		r.Get("/", f.get)
		r.Post("/run", f.run)
	})

	return router
}

type FlowCreateReq struct {
	Flow *sharedstate.Flow `json:"flow"`
}

type FlowCreateResp struct {
	Flow                 *sharedstate.Flow `json:"flow"`
	internalResponseMeta `json:"-"`
}

func (f flowsEndpoint) create(w http.ResponseWriter, r *http.Request) {

	var req FlowCreateReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpWriteResponseError(w, NewResponseError(fmt.Errorf("failed to decode object: %w", err), 400))
		return
	}

	if err := req.Flow.Validate(getNamespaceParam(r)); err != nil {
		httpWriteResponseError(w, NewResponseError(err, http.StatusBadRequest))
		return
	}

	stateResp, err := f.state.Flows().Create(&state.FlowsCreateReq{Flow: req.Flow})
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
	} else {
		resp := FlowCreateResp{
			Flow:                 stateResp.Flow,
			internalResponseMeta: newInternalResponseMeta(http.StatusCreated),
		}
		httpWriteResponse(w, &resp)
	}
}

type FlowListResp struct {
	Flows                []*sharedstate.FlowStub `json:"flows"`
	internalResponseMeta `json:"-"`
}

func (f flowsEndpoint) list(w http.ResponseWriter, r *http.Request) {
	stateResp, err := f.state.Flows().List(&state.FlowsListReq{
		Namespace: r.URL.Query().Get("namespace"),
	})
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
	} else {
		resp := FlowListResp{
			Flows:                stateResp.Flows,
			internalResponseMeta: newInternalResponseMeta(http.StatusOK),
		}
		httpWriteResponse(w, &resp)
	}
}

type FlowDeleteResp struct {
	internalResponseMeta `json:"-"`
}

func (f flowsEndpoint) delete(w http.ResponseWriter, r *http.Request) {

	flowID := r.Context().Value("id").(string)

	_, err := f.state.Flows().Delete(&state.FlowsDeleteReq{
		ID:        flowID,
		Namespace: r.URL.Query().Get("namespace"),
	})
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
	} else {
		resp := FlowDeleteResp{
			internalResponseMeta: newInternalResponseMeta(http.StatusOK),
		}
		httpWriteResponse(w, &resp)
	}
}

type FlowGetReq struct {
	ID string `json:"id"`
}

type FlowGetResp struct {
	Flow                 *sharedstate.Flow `json:"flow"`
	internalResponseMeta `json:"-"`
}

func (f flowsEndpoint) get(w http.ResponseWriter, r *http.Request) {

	flowID := r.Context().Value("id").(string)

	stateResp, err := f.state.Flows().Get(&state.FlowsGetReq{
		ID:        flowID,
		Namespace: r.URL.Query().Get("namespace"),
	})
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
	} else {
		resp := FlowGetResp{
			Flow:                 stateResp.Flow,
			internalResponseMeta: newInternalResponseMeta(http.StatusOK),
		}
		httpWriteResponse(w, &resp)
	}
}

type FlowRunReq struct {
	Variables map[string]any `json:"variables"`
}

type FlowRunResp struct {
	RunID                ulid.ULID `json:"run_id"`
	internalResponseMeta `json:"-"`
}

func (f flowsEndpoint) run(w http.ResponseWriter, r *http.Request) {

	var req FlowRunReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpWriteResponseError(w, NewResponseError(fmt.Errorf("failed to decode object: %w", err), 400))
		return
	}

	runID, err := f.runController.RunFlow(
		r.Context().Value("id").(string),
		getNamespaceParam(r),
		"manual",
		req.Variables,
	)
	if err != nil {
		respErr := NewResponseError(err, http.StatusInternalServerError)
		httpWriteResponseError(w, respErr)
	} else {
		resp := FlowRunResp{
			RunID:                runID,
			internalResponseMeta: newInternalResponseMeta(http.StatusCreated),
		}
		httpWriteResponse(w, &resp)
	}
}

func (f flowsEndpoint) context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var flowID string

		if flowID = chi.URLParam(r, "id"); flowID == "" {
			httpWriteResponseError(w, errors.New("flow not found"))
			return
		}

		ctx := context.WithValue(r.Context(), "id", flowID) //nolint:staticcheck
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
