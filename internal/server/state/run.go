package state

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type Runs interface {
	Create(*RunsCreateReq) (*RunsCreateResp, *ErrorResp)
	Delete(*RunsDeleteReq) (*RunsDeleteResp, *ErrorResp)
	Get(*RunsGetReq) (*RunsGetResp, *ErrorResp)
	List(*RunsListReq) (*RunsListResp, *ErrorResp)
	Update(*RunsUpdateReq) (*RunsUpdateResp, *ErrorResp)
}

type RunsCreateReq struct {
	Run *Run `json:"run"`
}

type RunsCreateResp struct{}

type RunsDeleteReq struct {
	ID ulid.ULID `json:"id"`
}

type RunsDeleteResp struct{}

type RunsGetReq struct {
	ID ulid.ULID `json:"id"`
}

type RunsGetResp struct {
	Run *Run `json:"run"`
}

type RunsListReq struct{}

type RunsListResp struct {
	Runs []*RunStub `json:"runs"`
}

type RunsUpdateReq struct {
	Run *Run `json:"run"`
}

type RunsUpdateResp struct{}

const (
	RunStatusPending = "pending"
	RunStatusRunning = "running"
	RunStatusSuccess = "success"
	RunStatusFailed  = "failed"
)

type Run struct {
	ID     ulid.ULID `json:"id"`
	FlowID string    `json:"flow_id"`
	Status string    `json:"status"`
	Jobs   []*JobRun `json:"jobs"`

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type JobRun struct {
	ID            string               `json:"id"`
	Status        string               `json:"status"`
	StartTime     time.Time            `json:"start_time"`
	EndTime       time.Time            `json:"end_time"`
	Specification *RunJobSpecification `json:"specification"`
	Inline        []*RunJobInline      `json:"inline"`
}

type RunJobSpecification struct {
	ID             string    `json:"id"`
	Status         string    `json:"status"`
	NomadJobID     string    `json:"nomad_job_id"`
	NomadNamespace string    `json:"nomad_namespace"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
}

type RunJobInline struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	ExitCode  int       `json:"exit_code"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

func (r *Run) Stub() *RunStub {
	return &RunStub{
		ID:        r.ID,
		FlowID:    r.FlowID,
		Status:    r.Status,
		StartTime: r.StartTime,
		EndTime:   r.EndTime,
	}
}

type RunStub struct {
	ID        ulid.ULID `json:"id"`
	FlowID    string    `json:"flow_id"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}
