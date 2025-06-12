package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/nomad/api"
	"github.com/oklog/ulid/v2"
)

type Flow struct {
	ID   string     `hcl:"id" json:"id"`
	Jobs []*FlowJob `hcl:"job,block" json:"jobs"`
}

type FlowJob struct {
	ID               string              `hcl:"id,label" json:"id"`
	Resource         *FlowJobResource    `hcl:"resource,block" json:"resource"`
	NomadNamespace   string              `hcl:"nomad_namespace,optional" json:"nomad_namespace,optional"`
	Artifacts        []*api.TaskArtifact `hcl:"artifact,block" json:"artifact"`
	JobSpecification *JobSpecification   `hcl:"specification,block" json:"specification"`
	Steps            []*Step             `hcl:"step,block" json:"step"`
}

type FlowJobResource struct {
	CPU    int `hcl:"cpu,optional" json:"cpu,optional"`
	Memory int `hcl:"memory,optional" json:"memory,optional"`
}

type JobSpecification struct {
	Path string `hcl:"path,optional" json:"path,optional"`
	Data string `hcl:"data,optional" json:"data,optional"`
}

type Step struct {
	ID  string `hcl:"id,label" json:"id"`
	Run string `hcl:"run" json:"run"`
}

type FlowStub struct {
	ID   string   `json:"id"`
	Jobs []string `json:"jobs"`
}

const (
	FlowJobTypeSpecification = "specification"
	FlowJobTypeInline        = "inline"
	FlowJobTypeUnknown       = "unknown"
)

func (f *FlowJob) Type() string {
	if f.isFileType() {
		return FlowJobTypeSpecification
	}
	if f.isStepType() {
		return FlowJobTypeInline
	}
	return FlowJobTypeUnknown
}

func (f *FlowJob) isFileType() bool { return f.JobSpecification != nil }

func (f *FlowJob) isStepType() bool { return len(f.Steps) > 0 }

func ParseFlowFile(path string) (*Flow, error) {

	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	fileExt := filepath.Ext(path)

	var obj Flow

	switch fileExt {
	case ".hcl":
		if err := hclsimple.DecodeFile(path, nil, &obj); err != nil {
			return nil, fmt.Errorf("failed to decode file: %w", err)
		}

		for _, job := range obj.Jobs {

			if job.JobSpecification != nil && (job.JobSpecification.Path != "" && job.JobSpecification.Data == "") {

				absPath, err := filepath.Abs(job.JobSpecification.Path)
				if err != nil {
					return nil, err
				}

				file, err := os.ReadFile(absPath)
				if err != nil {
					return nil, err
				}

				job.JobSpecification.Data = string(file)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported file extension: %q", fileExt)
	}

	return &obj, nil
}

type Flows struct {
	client *Client
}

func (c *Client) Flows() *Flows {
	return &Flows{client: c}
}

type FlowCreateReq struct {
	Flow *Flow `json:"flow"`
}

type FlowCreateResp struct {
	Flow *Flow `json:"flow"`
}

func (f *Flows) Create(ctx context.Context, req *FlowCreateReq) (*FlowCreateResp, *Response, error) {

	var resp FlowCreateResp

	httpReq, err := f.client.NewRequest(http.MethodPost, "/v1alpha1/flows", req)
	if err != nil {
		return nil, nil, err
	}

	httpResp, err := f.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, nil, err
	}

	return &resp, httpResp, nil
}

type FlowDeleteReq struct {
	ID string `json:"id"`
}

type FlowDeleteResp struct{}

func (f *Flows) Delete(ctx context.Context, req *FlowDeleteReq) (*Response, error) {

	httpReq, err := f.client.NewRequest(http.MethodDelete, "/v1alpha1/flows/"+req.ID, nil)
	if err != nil {
		return nil, err
	}

	httpResp, err := f.client.Do(ctx, httpReq, nil)
	if err != nil {
		return httpResp, err
	}

	return httpResp, nil
}

type FlowsGetReq struct {
	ID string `json:"id"`
}

type FlowsGetResp struct {
	Flow *Flow `json:"flow"`
}

func (f *Flows) Get(ctx context.Context, req *FlowsGetReq) (*FlowsGetResp, *Response, error) {

	var resp FlowsGetResp

	httpReq, err := f.client.NewRequest(http.MethodGet, "/v1alpha1/flows/"+req.ID, nil)
	if err != nil {
		return nil, nil, err
	}

	httpResp, err := f.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, httpResp, err
	}

	return &resp, httpResp, nil
}

type FlowListReq struct{}

type FlowListResp struct {
	Flows []*FlowStub `json:"flows"`
}

func (f *Flows) List(ctx context.Context, _ *FlowListReq) (*FlowListResp, *Response, error) {

	var resp FlowListResp

	httpReq, err := f.client.NewRequest(http.MethodGet, "/v1alpha1/flows", nil)
	if err != nil {
		return nil, nil, err
	}

	httpResp, err := f.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, httpResp, err
	}

	return &resp, httpResp, nil
}

type FlowRunReq struct {
	ID string `json:"id"`
}

type FlowRunResp struct {
	RunID ulid.ULID `json:"run_id"`
}

func (f *Flows) Run(ctx context.Context, req *FlowRunReq) (*FlowRunResp, *Response, error) {

	var flowRunResp FlowRunResp

	httpReq, err := f.client.NewRequest(http.MethodPost, "/v1alpha1/flows/"+req.ID+"/run", nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := f.client.Do(ctx, httpReq, &flowRunResp)
	if err != nil {
		return nil, resp, err
	}

	return &flowRunResp, resp, nil
}
