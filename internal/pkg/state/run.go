package state

import (
	"maps"
	"time"

	"github.com/oklog/ulid/v2"
)

const (
	RunStatusPending   = "pending"
	RunStatusRunning   = "running"
	RunStatusSuccess   = "success"
	RunStatusCancelled = "cancelled"
	RunStatusFailed    = "failed"
	RunStatusSkipped   = "skipped"
)

type RunNamespacedKey struct {
	ID        ulid.ULID
	Namespace string
}

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
	ID    string        `json:"id"`
	Steps []*InlineStep `json:"inline"`
}

type InlineStep struct {
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
	Trigger    string    `json:"trigger"`
	CreateTime time.Time `json:"create_time"`
}

func (r *Run) Stub() *RunStub {
	return &RunStub{
		ID:         r.ID,
		Namespace:  r.Namespace,
		FlowID:     r.FlowID,
		Status:     r.Status,
		Trigger:    r.Trigger,
		CreateTime: r.CreateTime,
	}
}

func (r *Run) MarkCancelled() {
	t := time.Now()

	r.EndTime = t
	r.Status = RunStatusCancelled

	if r.InlineRun != nil {
		for _, step := range r.InlineRun.Steps {
			if step.Status == RunStatusPending || step.Status == RunStatusRunning {
				step.Status = RunStatusCancelled
				step.EndTime = t
			}
		}
	}

	if r.SpecRun != nil {
		for _, spec := range r.SpecRun.Specs {
			if spec.Status == RunStatusPending || spec.Status == RunStatusRunning {
				spec.Status = RunStatusCancelled
				spec.EndTime = t
			}
		}
	}
}

func (r *Run) MarkFailed() {
	r.EndTime = time.Now()
	r.Status = RunStatusFailed
}

func (r *Run) Type() string {
	if r.InlineRun != nil {
		return "inline"
	}
	if r.SpecRun != nil {
		return "specification"
	}
	return "unknown"
}

func (r *Run) Copy() *Run {
	if r == nil {
		return nil
	}

	copy := &Run{
		ID:         r.ID,
		FlowID:     r.FlowID,
		Namespace:  r.Namespace,
		Status:     r.Status,
		Trigger:    r.Trigger,
		Variables:  make(map[string]any),
		CreateTime: r.CreateTime,
		StartTime:  r.StartTime,
		EndTime:    r.EndTime,
	}

	maps.Copy(copy.Variables, r.Variables)

	if r.InlineRun != nil {
		copy.InlineRun = &InlineRun{
			ID:    r.InlineRun.ID,
			Steps: make([]*InlineStep, len(r.InlineRun.Steps)),
		}
		for i, step := range r.InlineRun.Steps {
			if step != nil {
				copy.InlineRun.Steps[i] = &InlineStep{
					ID:        step.ID,
					Status:    step.Status,
					ExitCode:  step.ExitCode,
					StartTime: step.StartTime,
					EndTime:   step.EndTime,
				}
			}
		}
	}

	if r.SpecRun != nil {
		copy.SpecRun = &SpecRun{
			Specs: make([]*Spec, len(r.SpecRun.Specs)),
		}
		for i, spec := range r.SpecRun.Specs {
			if spec != nil {
				copy.SpecRun.Specs[i] = &Spec{
					ID:                spec.ID,
					NomadJobID:        spec.NomadJobID,
					NomadJobNamespace: spec.NomadJobNamespace,
					Status:            spec.Status,
					StartTime:         spec.StartTime,
					EndTime:           spec.EndTime,
				}
			}
		}
	}

	return copy
}
