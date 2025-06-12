package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/runner/logs"
	"github.com/hashicorp-forge/nomad-pipeline/internal/server/state"
)

type runsEndpoint struct {
	dataDir string
	state   state.State
}

func (re runsEndpoint) routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", re.list)
	})

	r.Route("/{id}", func(r chi.Router) {
		r.Use(re.context)
		r.Delete("/", re.delete)
		r.Get("/", re.get)

		r.Route("/logs", func(r chi.Router) {
			r.Get("/", re.logs)
		})
	})

	return r
}

type RunListResp struct {
	Runs                 []*state.RunStub `json:"runs"`
	internalResponseMeta `json:"-"`
}

func (re runsEndpoint) list(w http.ResponseWriter, r *http.Request) {
	stateResp, err := re.state.Runs().List(&state.RunsListReq{})
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

	runID := r.Context().Value("id").(ulid.ULID)

	_, err := re.state.Runs().Delete(&state.RunsDeleteReq{ID: runID})
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
	Run                  *state.Run `json:"run"`
	internalResponseMeta `json:"-"`
}

func (re runsEndpoint) get(w http.ResponseWriter, r *http.Request) {

	runID := r.Context().Value("id").(ulid.ULID)

	stateResp, err := re.state.Runs().Get(&state.RunsGetReq{ID: runID})
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
	JobID  string `json:"job_id"`
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

	var stepID, jobID, logType string

	if stepID = r.URL.Query().Get("step_id"); stepID == "" {
		httpWriteResponseError(w, NewResponseError(errors.New("step_id not provided"), http.StatusBadRequest))
		return
	}
	if jobID = r.URL.Query().Get("job_id"); jobID == "" {
		httpWriteResponseError(w, NewResponseError(errors.New("job_id not provided"), http.StatusBadRequest))
		return
	}
	if logType = r.URL.Query().Get("type"); logType == "" {
		httpWriteResponseError(w, NewResponseError(errors.New("type not provided"), http.StatusBadRequest))
		return
	}

	path := filepath.Join(re.dataDir, "runs", runID.String(), jobID, stepID, logType+".log")

	lines, err := logs.Get(path)
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

	streamCh := make(chan string)
	errorCh := make(chan error)

	var stepID, jobID, logType string

	if stepID = r.URL.Query().Get("step_id"); stepID == "" {
		httpWriteResponseError(w, NewResponseError(errors.New("step_id not provided"), http.StatusBadRequest))
		return
	}
	if jobID = r.URL.Query().Get("job_id"); jobID == "" {
		httpWriteResponseError(w, NewResponseError(errors.New("job_id not provided"), http.StatusBadRequest))
		return
	}
	if logType = r.URL.Query().Get("type"); logType == "" {
		httpWriteResponseError(w, NewResponseError(errors.New("type not provided"), http.StatusBadRequest))
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	runID := r.Context().Value("id").(ulid.ULID)

	path := filepath.Join(re.dataDir, "runs", runID.String(), jobID, stepID, logType+".log")

	logStreamer := logs.NewStream(path, streamCh, errorCh)
	go logStreamer.Run(context.Background())

	for {
		select {
		case <-r.Context().Done():
			return
		case err := <-errorCh:
			httpWriteResponseError(w, err)
			return
		case line := <-streamCh:
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
