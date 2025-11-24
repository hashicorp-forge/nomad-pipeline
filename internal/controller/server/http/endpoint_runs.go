package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/coordinator"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	sharedstate "github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type runsEndpoint struct {
	coordinator *coordinator.Coordinator
	state       state.State
}

func (re runsEndpoint) routes() chi.Router {
	router := chi.NewRouter()

	router.Route("/", func(r chi.Router) {
		r.Use(namespaceCheckMiddleware(re.state))
		r.Get("/", re.list)
	})

	router.Route("/{id}", func(r chi.Router) {
		r.Use(namespaceWildcardRejectMiddleware())
		r.Use(namespaceCheckMiddleware(re.state))
		r.Use(re.context)
		r.Delete("/", re.delete)
		r.Get("/", re.get)

		r.Route("/cancel", func(r chi.Router) {
			r.Put("/", re.cancel)
		})
		r.Route("/logs", func(r chi.Router) {
			r.Get("/", re.logs)
		})
	})

	return router
}

type RunCancelResp struct {
	internalResponseMeta `json:"-"`
}

func (re runsEndpoint) cancel(w http.ResponseWriter, r *http.Request) {
	err := re.coordinator.CancelRun(r.Context().Value("id").(ulid.ULID), getNamespaceParam(r))
	if err != nil {
		respErr := NewResponseError(err, http.StatusInternalServerError)
		httpWriteResponseError(w, respErr)
	} else {
		resp := RunCancelResp{
			internalResponseMeta: newInternalResponseMeta(http.StatusOK),
		}
		httpWriteResponse(w, &resp)
	}
}

type RunListResp struct {
	Runs                 []*sharedstate.RunStub `json:"runs"`
	internalResponseMeta `json:"-"`
}

func (re runsEndpoint) list(w http.ResponseWriter, r *http.Request) {
	stateResp, err := re.state.Runs().List(&state.RunsListReq{
		Namespace: getNamespaceParam(r),
	})
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
	} else {
		resp := RunListResp{
			Runs:                 stateResp.Runs,
			internalResponseMeta: newInternalResponseMeta(http.StatusCreated),
		}
		httpWriteResponse(w, &resp)
	}
}

type RunDeleteResp struct {
	internalResponseMeta `json:"-"`
}

func (re runsEndpoint) delete(w http.ResponseWriter, r *http.Request) {

	_, err := re.state.Runs().Delete(&state.RunsDeleteReq{
		ID:        r.Context().Value("id").(ulid.ULID),
		Namespace: getNamespaceParam(r),
	})
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
	} else {
		resp := RunDeleteResp{
			internalResponseMeta: newInternalResponseMeta(http.StatusOK),
		}
		httpWriteResponse(w, &resp)
	}
}

type RunGetResp struct {
	Run                  *sharedstate.Run `json:"run"`
	internalResponseMeta `json:"-"`
}

func (re runsEndpoint) get(w http.ResponseWriter, r *http.Request) {

	id := r.Context().Value("id").(ulid.ULID)

	stateResp, err := re.state.Runs().Get(&state.RunsGetReq{
		ID:        id,
		Namespace: getNamespaceParam(r),
	})
	if err != nil {
		respErr := NewResponseError(err.Err(), err.StatusCode())
		httpWriteResponseError(w, respErr)
	} else {
		resp := RunGetResp{
			Run:                  stateResp.Run,
			internalResponseMeta: newInternalResponseMeta(http.StatusCreated),
		}
		httpWriteResponse(w, &resp)
	}
}

type RunsLogsReq struct {
	StepID string `json:"step_id"`
	Tail   bool   `json:"tail"`
}

type RunsLogsResp struct {
	Logs                 []string `json:"logs"`
	internalResponseMeta `json:"-"`
}

func (re runsEndpoint) logs(w http.ResponseWriter, r *http.Request) {

	tail, err := strconv.ParseBool(r.URL.Query().Get("tail"))
	if err != nil {
		httpWriteResponseError(w, NewResponseError(err, http.StatusBadRequest))
		return
	}

	if tail {
		re.logsStream(w, r)
	} else {
		re.logsGet(w, r)
	}
}

func (re runsEndpoint) logsGet(w http.ResponseWriter, r *http.Request) {

	runID := r.Context().Value("id").(ulid.ULID)

	var stepID, logType string

	if stepID = r.URL.Query().Get("step_id"); stepID == "" {
		httpWriteResponseError(w, NewResponseError(errors.New("step_id not provided"), http.StatusBadRequest))
		return
	}
	if logType = r.URL.Query().Get("type"); logType == "" {
		httpWriteResponseError(w, NewResponseError(errors.New("type not provided"), http.StatusBadRequest))
		return
	}

	lines, err := re.coordinator.Getlogs(
		getNamespaceParam(r),
		runID.String(),
		stepID,
		logType,
	)
	if err != nil {
		respErr := NewResponseError(err, http.StatusInternalServerError)
		httpWriteResponseError(w, respErr)
	} else {
		resp := RunsLogsResp{
			Logs:                 lines,
			internalResponseMeta: newInternalResponseMeta(http.StatusOK),
		}
		httpWriteResponse(w, &resp)
	}
}

func (re runsEndpoint) logsStream(w http.ResponseWriter, r *http.Request) {

	flusher, ok := w.(http.Flusher)
	if !ok {
		httpWriteResponseError(w, NewResponseError(errors.New("streaming not supported"), http.StatusInternalServerError))
		return
	}

	var stepID, logType string

	if stepID = r.URL.Query().Get("step_id"); stepID == "" {
		httpWriteResponseError(w, NewResponseError(errors.New("step_id not provided"), http.StatusBadRequest))
		return
	}
	if logType = r.URL.Query().Get("type"); logType == "" {
		httpWriteResponseError(w, NewResponseError(errors.New("type not provided"), http.StatusBadRequest))
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	id := r.Context().Value("id").(ulid.ULID)

	logStreamer := re.coordinator.StreamLogs(
		getNamespaceParam(r),
		id.String(),
		stepID,
		logType,
	)

	go logStreamer.Run(context.Background())

	for {
		select {
		case <-r.Context().Done():
			return
		case err := <-logStreamer.ErrorCh():
			httpWriteResponseError(w, err)
			return
		case line := <-logStreamer.StreamCh():
			_, _ = fmt.Fprint(w, line+"\n")
			flusher.Flush()
		}
	}
}

func (re runsEndpoint) context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var runID string

		if runID = chi.URLParam(r, "id"); runID == "" {
			httpWriteResponseError(w, errors.New("run not found"))
			return
		}

		if runULID, err := ulid.Parse(runID); err != nil {
			httpWriteResponseError(w, fmt.Errorf("failed to parse ID: %w", err))
		} else {
			ctx := context.WithValue(r.Context(), "id", runULID) //nolint:staticcheck
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}
