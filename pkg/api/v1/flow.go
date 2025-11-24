package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/oklog/ulid/v2"
	"github.com/zclconf/go-cty/cty"
)

type Flow struct {
	ID        string `hcl:"id,label" json:"id"`
	Namespace string `hcl:"namespace" json:"namespace"`

	Variables []*FlowVariable `hcl:"variable,block" json:"variable"`

	Inline        *InlineFlow          `hcl:"inline,block" json:"inline"`
	Specification []*SpecificationFlow `hcl:"specification,block" json:"specification"`
}

type InlineFlow struct {
	ID     string      `hcl:"id,label" json:"id"`
	Runner *FlowRunner `hcl:"runner,block" json:"runner"`
	Steps  []*Step     `hcl:"step,block" json:"step"`
}

type SpecificationFlow struct {
	ID        string            `hcl:"id,label" json:"id"`
	Condition string            `hcl:"condition,optional" json:"condition"`
	Job       *JobSpecification `hcl:"job,block" json:"job"`
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
	Source string `hcl:"source,optional" json:"source"`
	Dest   string `hcl:"destination,optional" json:"destination"`

	Options map[string]string `json:"options"`
	Remain  hcl.Body          `hcl:",remain"`
}

type NomadOnDemandResource struct {
	CPU    int `hcl:"cpu,optional" json:"cpu"`
	Memory int `hcl:"memory,optional" json:"memory"`
}

type JobSpecification struct {
	NameFormat     string            `json:"name_format"`
	NameFormatExpr hcl.Expression    `hcl:"name_format,optional"`
	Path           string            `hcl:"path,optional" json:"path"`
	Raw            string            `hcl:"raw,optional" json:"raw"`
	Variables      map[string]string `hcl:"variables,optional" json:"variables"`
}

type Step struct {
	ID        string         `hcl:"id,label" json:"id"`
	Condition string         `hcl:"condition,optional" json:"condition"`
	Run       string         `json:"run"`
	RunExpr   hcl.Expression `hcl:"run,optional"`
}

type FlowVariable struct {
	Name string `hcl:"name,label" json:"name"`

	Type     string         `json:"type"`
	TypeExpr hcl.Expression `hcl:"type,optional"`

	Required bool `hcl:"required,optional" json:"required"`

	Default     any            `json:"default"`
	DefaultExpr hcl.Expression `hcl:"default,optional"`
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

type FlowStub struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
}

const (
	FlowTypeSpecification = "specification"
	FlowTypeInline        = "inline"
	FlowTypeUnknown       = "unknown"
)

func ParseFlowFile(path string) (*Flow, error) {

	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Use a wrapped object to decode the flow block. This makes decoding
	// slightly easier, but keeps the flow specification struct clean and
	// flater.
	decodeObj := struct {
		Flow *Flow `hcl:"flow,block"`
	}{}

	fileExt := filepath.Ext(path)
	var srcData []byte

	switch fileExt {
	case ".hcl":
		srcData = data

		file, diags := hclsyntax.ParseConfig(data, path, hcl.InitialPos)
		if diags.HasErrors() {
			return nil, diags
		}

		diags = gohcl.DecodeBody(file.Body, nil, &decodeObj)
		if diags.HasErrors() {
			return nil, diags
		}

		// Decode flow variables
		for _, v := range decodeObj.Flow.Variables {
			if err := v.postDecodeProcessing(); err != nil {
				return nil, fmt.Errorf("failed to decode variable %q: %w", v.Name, err)
			}
		}

		// Decode artifact options from remain body
		if decodeObj.Flow.Inline != nil && decodeObj.Flow.Inline.Runner != nil && decodeObj.Flow.Inline.Runner.NomadOnDemand != nil {
			for _, artifact := range decodeObj.Flow.Inline.Runner.NomadOnDemand.Artifacts {
				if err := artifact.postDecodeProcessing(srcData); err != nil {
					return nil, fmt.Errorf("failed to decode artifact: %w", err)
				}
			}
		}

		switch decodeObj.Flow.Type() {
		case FlowTypeInline:
			for _, step := range decodeObj.Flow.Inline.Steps {
				step.postDecodeProcessing(data)
			}
		case FlowTypeSpecification:
			for _, spec := range decodeObj.Flow.Specification {
				if spec.Job.Raw == "" && spec.Job.Path != "" {
					jobData, err := os.ReadFile(spec.Job.Path)
					if err != nil {
						return nil, fmt.Errorf("failed to read job specification file %q: %w",
							spec.Job.Path, err)
					}
					spec.Job.Raw = string(jobData)
				}

				if err := spec.Job.postDecodeProcessing(srcData); err != nil {
					return nil, fmt.Errorf("failed to decode job specification for %q: %w", spec.ID, err)
				}
			}
		}
	default:
		return nil, fmt.Errorf("unsupported file extension: %q", fileExt)
	}

	return decodeObj.Flow, nil
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

	httpReq, err := f.client.NewRequest(http.MethodPost, "/v1/flows", req)
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

	httpReq, err := f.client.NewRequest(http.MethodDelete, "/v1/flows/"+req.ID, nil)
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

	httpReq, err := f.client.NewRequest(http.MethodGet, "/v1/flows/"+req.ID, nil)
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

	httpReq, err := f.client.NewRequest(http.MethodGet, "/v1/flows", nil)
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
	ID        string         `json:"id"`
	Variables map[string]any `json:"variables"`
}

type FlowRunResp struct {
	RunID ulid.ULID `json:"run_id"`
}

func (f *Flows) Run(ctx context.Context, req *FlowRunReq) (*FlowRunResp, *Response, error) {

	var flowRunResp FlowRunResp

	httpReq, err := f.client.NewRequest(http.MethodPost, "/v1/flows/"+req.ID+"/run", req)
	if err != nil {
		return nil, nil, err
	}

	resp, err := f.client.Do(ctx, httpReq, &flowRunResp)
	if err != nil {
		return nil, resp, err
	}

	return &flowRunResp, resp, nil
}

func (v *FlowVariable) postDecodeProcessing() error {
	if v.DefaultExpr != nil {
		val, diags := v.DefaultExpr.Value(nil)
		if diags.HasErrors() {
			return diags
		}

		var err error
		v.Default, err = ctyValueToGo(val)
		if err != nil {
			return fmt.Errorf("failed to convert default value: %w", err)
		}
	}

	if v.TypeExpr != nil {
		tp, diags := typeexpr.Type(v.TypeExpr)
		if diags.HasErrors() {
			return diags
		}

		v.Type = tp.FriendlyName()
	}

	return nil
}

func (j *JobSpecification) postDecodeProcessing(src []byte) error {
	if j.NameFormatExpr != nil {
		rng := j.NameFormatExpr.Range()
		if src != nil {
			rawExpr := string(extractBytes(src, rng))
			if len(rawExpr) >= 2 && rawExpr[0] == '"' && rawExpr[len(rawExpr)-1] == '"' {
				rawExpr = rawExpr[1 : len(rawExpr)-1]
			}
			j.NameFormat = rawExpr
		}
	}
	return nil
}

func (s *Step) postDecodeProcessing(src []byte) {
	if s.RunExpr != nil {
		rng := s.RunExpr.Range()
		if src != nil {
			s.Run = removeEOHMarkers(string(extractBytes(src, rng)))
		}
	}
}

func (f *FlowRunnerArtifact) postDecodeProcessing(src []byte) error {
	if f.Remain == nil {
		return nil
	}

	// Get the "options" attribute from the remaining body
	content, _, diags := f.Remain.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "options"},
		},
	})
	if diags.HasErrors() {
		return diags
	}

	optionsAttr, exists := content.Attributes["options"]
	if !exists {
		return nil
	}

	// Parse the options expression to extract map entries
	// For a map like {ref = "${github.ref}"}, we need to parse it
	expr := optionsAttr.Expr

	// Try to get the variables from the expression without evaluating
	if syntaxExpr, ok := expr.(*hclsyntax.ObjectConsExpr); ok {
		if f.Options == nil {
			f.Options = make(map[string]string)
		}

		for _, item := range syntaxExpr.Items {
			// Get the key
			keyVal, diags := item.KeyExpr.Value(nil)
			if diags.HasErrors() {
				continue
			}
			key := keyVal.AsString()

			// Get the raw source for the value expression
			valueRng := item.ValueExpr.Range()
			value := string(extractBytes(src, valueRng))

			// Remove quotes if it's a quoted string
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}

			f.Options[key] = value
		}
	}

	return nil
}

// extractBytes extracts bytes from source given an HCL range
func extractBytes(src []byte, rng hcl.Range) []byte {
	if rng.Start.Byte >= len(src) || rng.End.Byte > len(src) {
		return nil
	}
	return src[rng.Start.Byte:rng.End.Byte]
}

func removeEOHMarkers(input string) string {
	// Match <<EOH or <<-EOH (indented heredoc)
	re := regexp.MustCompile(`<<-?EOH\s*\n([\s\S]*?)\n\s*EOH\s*`)

	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return input
}

// ctyValueToGo converts a cty.Value to a native Go value
func ctyValueToGo(val cty.Value) (any, error) {
	if val.IsNull() {
		return nil, nil
	}

	ty := val.Type()

	switch {
	case ty.Equals(cty.String):
		return val.AsString(), nil
	case ty.Equals(cty.Number):
		bf := val.AsBigFloat()
		if bf.IsInt() {
			i, _ := bf.Int64()
			return i, nil
		}
		f, _ := bf.Float64()
		return f, nil
	case ty.Equals(cty.Bool):
		return val.True(), nil
	case ty.IsListType() || ty.IsTupleType():
		var result []any
		it := val.ElementIterator()
		for it.Next() {
			_, elemVal := it.Element()
			elem, err := ctyValueToGo(elemVal)
			if err != nil {
				return nil, err
			}
			result = append(result, elem)
		}
		return result, nil
	case ty.IsMapType() || ty.IsObjectType():
		result := make(map[string]any)
		it := val.ElementIterator()
		for it.Next() {
			key, elemVal := it.Element()
			elem, err := ctyValueToGo(elemVal)
			if err != nil {
				return nil, err
			}
			result[key.AsString()] = elem
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported cty type: %s", ty.FriendlyName())
	}
}
