package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/runner"
	"github.com/hashicorp-forge/nomad-pipeline/internal/server/state"
)

type flowsEndpoint struct {
	runController *runner.Controller
	state         state.State
}

func (f flowsEndpoint) routes() chi.Router {
	router := chi.NewRouter()

	router.Route("/", func(r chi.Router) {
		r.Post("/", f.create)
		r.Get("/", f.list)
	})

	router.Route("/{id}", func(r chi.Router) {
		r.Use(f.context)
		r.Delete("/", f.delete)
		r.Get("/", f.get)
		r.Post("/run", f.run)
	})

	return router
}

type FlowCreateReq struct {
	Flow *state.Flow `json:"flow"`
}

type FlowCreateResp struct {
	Flow                 *state.Flow `json:"flow"`
	internalResponseMeta `json:"-"`
}

func (f flowsEndpoint) create(w http.ResponseWriter, r *http.Request) {

	var req FlowCreateReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpWriteResponseError(w, NewResponseError(fmt.Errorf("failed to decode object: %w", err), 400))
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
	Flows                []*state.FlowStub `json:"flows"`
	internalResponseMeta `json:"-"`
}

func (f flowsEndpoint) list(w http.ResponseWriter, r *http.Request) {
	stateResp, err := f.state.Flows().List(&state.FlowsListReq{})
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

	_, err := f.state.Flows().Delete(&state.FlowsDeleteReq{ID: flowID})
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
	Flow                 *state.Flow `json:"flow"`
	internalResponseMeta `json:"-"`
}

func (f flowsEndpoint) get(w http.ResponseWriter, r *http.Request) {

	flowID := r.Context().Value("id").(string)

	stateResp, err := f.state.Flows().Get(&state.FlowsGetReq{ID: flowID})
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

type FlowRunReq struct{}

type FlowRunResp struct {
	RunID                ulid.ULID `json:"run_id"`
	internalResponseMeta `json:"-"`
}

func (f flowsEndpoint) run(w http.ResponseWriter, r *http.Request) {

	runID, err := f.runController.RunFlow(r.Context().Value("id").(string))
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
