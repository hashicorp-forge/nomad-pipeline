package state

import (
	"errors"
	"fmt"
)

type Flow struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`

	Variables []*HCLVariable `json:"variable"`

	//
	Inline        *InlineFlow          `hcl:"inline,block" json:"inline"`
	Specification []*SpecificationFlow `hcl:"specification,optional" json:"specification"`
}

type InlineFlow struct {
	ID     string      `hcl:"id,label" json:"id"`
	Runner *FlowRunner `hcl:"runner,block" json:"runner"`
	Steps  []*Step     `hcl:"step,block" json:"step"`
}

type FlowRunner struct {
	NomadOnDemand *FlowRunnerNomadOnDemand `hcl:"nomad_on_demand,block" json:"nomad_on_demand"`
}

type FlowRunnerNomadOnDemand struct {
	Namespace string                 `hcl:"namespace,optional" json:"namespace"`
	Image     string                 `hcl:"image" json:"image"`
	Artifacts []*FlowRunnerArtifact  `hcl:"artifact,block" json:"artifact"`
	Resource  *NomadOnDemandResource `hcl:"resource,block" json:"resource"`
}

type FlowRunnerArtifact struct {
	Source  string            `json:"source"`
	Dest    string            `json:"destination"`
	Options map[string]string `json:"options"`
}

type NomadOnDemandResource struct {
	CPU    int `hcl:"cpu,optional" json:"cpu"`
	Memory int `hcl:"memory,optional" json:"memory"`
}

type SpecificationFlow struct {
	ID               string            `json:"id"`
	Condition        string            `json:"condition"`
	JobSpecification *JobSpecification `json:"job"`
}

type JobSpecification struct {
	NameFormat string            `json:"name_format"`
	Path       string            `json:"path"`
	Raw        string            `json:"raw"`
	Variables  map[string]string `json:"variables"`
}

type Step struct {
	ID        string `json:"id"`
	Condition string `json:"condition"`
	Run       string `json:"run"`
}

type FlowStub struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
}

func (f *Flow) Stub() *FlowStub {
	return &FlowStub{
		ID:        f.ID,
		Namespace: f.Namespace,
		Type:      f.Type(),
	}
}

func (f *Flow) Type() string {
	if f.Inline != nil {
		return FlowTypeInline
	}
	if len(f.Specification) > 0 {
		return FlowTypeSpecification
	}
	return FlowTypeUnknown
}

const (
	FlowTypeSpecification = "specification"
	FlowTypeInline        = "inline"
	FlowTypeUnknown       = "unknown"
)

func (f *Flow) Validate(reqNamespace string) error {

	var errs []error

	if reqNamespace != f.Namespace {
		errs = append(errs, fmt.Errorf(
			"flow specification namespace %q does not match request namespace %q",
			f.Namespace, reqNamespace))
	}

	return errors.Join(errs...)
}
