package api

import (
	"bufio"
	"context"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/helper"
)

const (
	RunStatusPending   = "pending"
	RunStatusRunning   = "running"
	RunStatusSuccess   = "success"
	RunStatusFailed    = "failed"
	RunStatusCancelled = "cancelled"
	RunStatusSkipped   = "skipped"
)

type Run struct {
	ID        ulid.ULID `json:"id"`
	FlowID    string    `json:"flow_id"`
	Namespace string    `json:"namespace"`
	Status    string    `json:"status"`
	Trigger   string    `json:"trigger"`

	CreateTime time.Time `json:"create_time"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`

	Variables map[string]any `json:"variables"`

	*InlineRun `json:"inline_run,omitempty"`
	*SpecRun   `json:"spec_run,omitempty"`
}

type InlineRun struct {
	JobID string          `json:"job_id"`
	Steps []*RunJobInline `json:"inline"`
}

type RunJobInline struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	ExitCode  int       `json:"exit_code"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type SpecRun struct {
	Specs []*Spec `json:"spec"`
}

type Spec struct {
	ID                string    `json:"id"`
	NomadJobID        string    `json:"nomad_job_id"`
	NomadJobNamespace string    `json:"nomad_job_namespace"`
	Status            string    `json:"status"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
}

type RunStub struct {
	ID         ulid.ULID `json:"id"`
	Namespace  string    `json:"namespace"`
	FlowID     string    `json:"flow_id"`
	Status     string    `json:"status"`
	CreateTime time.Time `json:"create_time"`
}

type Runs struct {
	client *Client
}

func (c *Client) Runs() *Runs {
	return &Runs{client: c}
}

type RunCancelReq struct {
	ID ulid.ULID `json:"id"`
}

type RunCancelResp struct{}

func (r *Runs) Cancel(ctx context.Context, req *RunCancelReq) (*RunCancelResp, *Response, error) {

	var resp RunCancelResp

	httpReq, err := r.client.NewRequest(http.MethodPut, "/v1/runs/"+req.ID.String()+"/cancel", nil)
	if err != nil {
		return nil, nil, err
	}

	httpResp, err := r.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, httpResp, err
	}

	return &RunCancelResp{}, httpResp, nil
}

type RunGetReq struct {
	ID ulid.ULID `json:"id"`
}

type RunGetResp struct {
	Run *Run `json:"run"`
}

func (r *Runs) Get(ctx context.Context, req *RunGetReq) (*RunGetResp, *Response, error) {

	var resp RunGetResp

	httpReq, err := r.client.NewRequest(http.MethodGet, "/v1/runs/"+req.ID.String(), req)
	if err != nil {
		return nil, nil, err
	}

	httpResp, err := r.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, httpResp, err
	}

	return &resp, httpResp, nil
}

type RunListReq struct{}

type RunListResp struct {
	Runs []*RunStub `json:"runs"`
}

func (r *Runs) List(ctx context.Context, req *RunListReq) (*RunListResp, *Response, error) {

	var resp RunListResp

	httpReq, err := r.client.NewRequest(http.MethodGet, "/v1/runs", req)
	if err != nil {
		return nil, nil, err
	}

	httpResp, err := r.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, httpResp, err
	}

	return &resp, httpResp, nil
}

type RunLogsGetReq struct {
	ID     ulid.ULID `json:"id"`
	JobID  string    `json:"job_id"`
	StepID string    `json:"step_id"`
	Type   string    `json:"type"`
}

type RunLogsGetResp struct {
	Logs []string `json:"logs"`
}

func (r *Runs) LogsGet(ctx context.Context, req *RunLogsGetReq) (*RunLogsGetResp, *Response, error) {
	var resp RunLogsGetResp

	httpReq, err := r.client.NewRequest(
		http.MethodGet,
		"/v1/runs/"+req.ID.String()+"/logs",
		nil,
		func(r *http.Request) {
			q := r.URL.Query()
			q.Set("job_id", req.JobID)
			q.Set("step_id", req.StepID)
			q.Set("type", req.Type)
			q.Set("tail", "false")
			r.URL.RawQuery = q.Encode()
		},
	)
	if err != nil {
		return nil, nil, err
	}

	httpResp, err := r.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, httpResp, err
	}

	return &resp, httpResp, nil
}

type RunLogsTailReq struct {
	ID     ulid.ULID `json:"id"`
	StepID string    `json:"step_id"`
	Type   string    `json:"type"`
}

type RunLogsTailResp struct {
	LogCh chan string
	ErrCh chan error
}

func (r *Runs) LogsTail(ctx context.Context, req *RunLogsTailReq) (*RunLogsTailResp, *Response, error) {
	var resp RunLogsTailResp

	httpReq, err := r.client.NewRequest(
		http.MethodGet,
		"/v1/runs/"+req.ID.String()+"/logs",
		nil,
		func(r *http.Request) {
			q := r.URL.Query()
			q.Set("step_id", req.StepID)
			q.Set("type", req.Type)
			q.Set("tail", "true")
			r.URL.RawQuery = q.Encode()
		},
	)
	if err != nil {
		return nil, nil, err
	}

	httpResp, err := r.client.bareDo(ctx, httpReq)
	if err != nil {
		return nil, httpResp, err
	}

	resp.LogCh = make(chan string, 10)
	resp.ErrCh = make(chan error)

	go func() {

		defer helper.IgnoreError(httpResp.Body.Close)

		scanner := bufio.NewScanner(httpResp.Body)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			for scanner.Scan() {
				resp.LogCh <- scanner.Text()
			}
		}
	}()

	return &resp, httpResp, nil
}
