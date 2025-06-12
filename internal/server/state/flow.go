package state

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/nomad/api"
)

type Flows interface {
	Create(*FlowsCreateReq) (*FlowsCreateResp, *ErrorResp)
	Delete(*FlowsDeleteReq) (*FlowsDeleteResp, *ErrorResp)
	Get(*FlowsGetReq) (*FlowsGetResp, *ErrorResp)
	List(*FlowsListReq) (*FlowsListResp, *ErrorResp)
}

type FlowsCreateReq struct {
	Flow *Flow
}

type FlowsCreateResp struct {
	Flow *Flow
}

type FlowsDeleteReq struct {
	ID string
}

type FlowsDeleteResp struct{}

type FlowsGetReq struct {
	ID string
}

type FlowsGetResp struct {
	Flow *Flow
}

type FlowsListReq struct{}

type FlowsListResp struct {
	Flows []*FlowStub
}

type Flow struct {
	ID   string     `json:"id"`
	Jobs []*FlowJob `json:"jobs"`
}

type FlowJob struct {
	ID               string              `json:"id"`
	Resource         *FlowJobResource    `json:"resource"`
	NomadNamespace   string              `json:"nomad_namespace"`
	Artifacts        []*api.TaskArtifact `json:"artifact"`
	JobSpecification *JobSpecification   `json:"specification"`
	Steps            []*Step             `json:"step"`
}

type FlowJobResource struct {
	CPU    int `json:"cpu,optional"`
	Memory int `json:"memory,optional"`
}

type JobSpecification struct {
	Path string `json:"path"`
	Data string `json:"data"`
}

type Step struct {
	ID  string `json:"id"`
	Run string `json:"run"`
}

type FlowStub struct {
	ID   string   `json:"id"`
	Jobs []string `json:"jobs"`
}

func (f *Flow) Stub() *FlowStub {

	jobs := make([]string, len(f.Jobs))
	for i, job := range f.Jobs {
		jobs[i] = job.ID
	}

	return &FlowStub{
		ID:   f.ID,
		Jobs: jobs,
	}
}

const (
	FlowJobTypeSpecification = "specification"
	FlowJobTypeInline        = "inline"
	FlowJobTypeUnknown       = "unknown"
)

func (f *Flow) Validate() error {

	var mErr *multierror.Error

	for _, job := range f.Jobs {
		if job.isFileType() && job.isStepType() {
			mErr = multierror.Append(mErr, fmt.Errorf(
				"job %s cannot contain steps and a specification file", job.ID))
		}
	}

	return mErr.ErrorOrNil()
}

func (f *FlowJob) isFileType() bool { return f.JobSpecification != nil }

func (f *FlowJob) isStepType() bool { return len(f.Steps) > 0 }

func (f *FlowJob) Type() string {
	if f.isFileType() {
		return FlowJobTypeSpecification
	}
	if f.isStepType() {
		return FlowJobTypeInline
	}
	return FlowJobTypeUnknown
}
