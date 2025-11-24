package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Trigger struct {
	ID        string `hcl:"id,label" json:"id"`
	Namespace string `hcl:"namespace" json:"namespace"`
	Flow      string `hcl:"flow" json:"flow"`

	Source *TriggerSource `hcl:"source,block" json:"source"`
}

type TriggerSource struct {
	ID       string `hcl:"id,label" json:"id"`
	Provider string `hcl:"provider" json:"provider"`
	Config   []byte `json:"config"`
}

type TriggerSchedule struct {
	Crons []string `hcl:"crons" json:"crons"`
}

func (t *Trigger) Validate() error {

	var errs []error

	if t.ID == "" {
		errs = append(errs, errors.New("trigger ID cannot be empty"))
	}

	if t.Namespace == "" {
		errs = append(errs, errors.New("namespace cannot be empty"))
	}

	if t.Flow == "" {
		errs = append(errs, errors.New("flow cannot be empty"))
	}

	if t.Source == nil {
		errs = append(errs, errors.New("trigger source cannot be empty"))
	}

	return errors.Join(errs...)
}

type TriggerStub struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
	Flow      string `json:"flow"`
}

func ParseTriggerFile(path string) (*Trigger, error) {

	// Read the source file
	srcData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the HCL file
	file, diags := hclsyntax.ParseConfig(srcData, path, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL: %w", diags)
	}

	// Find the trigger block
	var triggerBlock *hclsyntax.Block
	body := file.Body.(*hclsyntax.Body)
	for _, block := range body.Blocks {
		if block.Type == "trigger" {
			triggerBlock = block
			break
		}
	}

	if triggerBlock == nil {
		return nil, fmt.Errorf("no trigger block found")
	}

	// Parse basic trigger fields
	trigger := &Trigger{}

	if len(triggerBlock.Labels) > 0 {
		trigger.ID = triggerBlock.Labels[0]
	}

	bodyContent := triggerBlock.Body

	// Extract namespace and flow attributes
	if attr, exists := bodyContent.Attributes["namespace"]; exists {
		val, diags := attr.Expr.Value(nil)
		if !diags.HasErrors() {
			trigger.Namespace = val.AsString()
		}
	}

	if attr, exists := bodyContent.Attributes["flow"]; exists {
		val, diags := attr.Expr.Value(nil)
		if !diags.HasErrors() {
			trigger.Flow = val.AsString()
		}
	}

	// Find the source block
	for _, block := range bodyContent.Blocks {
		if block.Type == "source" {
			trigger.Source = &TriggerSource{}

			if len(block.Labels) > 0 {
				trigger.Source.ID = block.Labels[0]
			}

			sourceBody := block.Body

			// Extract provider attribute
			if attr, exists := sourceBody.Attributes["provider"]; exists {
				val, diags := attr.Expr.Value(nil)
				if !diags.HasErrors() {
					trigger.Source.Provider = val.AsString()
				}
			}

			// Find and extract the config block
			for _, configBlock := range sourceBody.Blocks {
				if configBlock.Type == "config" {
					configBody := configBlock.Body

					// Extract the raw bytes of the config block's content
					if len(configBody.Attributes) > 0 || len(configBody.Blocks) > 0 {
						var startByte, endByte int
						startByte = len(srcData)
						endByte = 0

						// Find earliest start and latest end of config content
						for _, attr := range configBody.Attributes {
							if attr.Range().Start.Byte < startByte {
								startByte = attr.Range().Start.Byte
							}
							if attr.Range().End.Byte > endByte {
								endByte = attr.Range().End.Byte
							}
						}
						for _, blk := range configBody.Blocks {
							if blk.Range().Start.Byte < startByte {
								startByte = blk.Range().Start.Byte
							}
							if blk.Range().End.Byte > endByte {
								endByte = blk.Range().End.Byte
							}
						}

						if startByte < endByte && endByte <= len(srcData) {
							trigger.Source.Config = srcData[startByte:endByte]
						}
					}
					break
				}
			}
			break
		}
	}

	return trigger, nil
}

type Triggers struct {
	client *Client
}

func (c *Client) Triggers() *Triggers {
	return &Triggers{client: c}
}

type TriggerCreateReq struct {
	Trigger *Trigger `json:"trigger"`
}

type TriggerCreateResp struct {
	Trigger *Trigger `json:"trigger"`
}

func (t *Triggers) Create(ctx context.Context, req *TriggerCreateReq) (*TriggerCreateResp, *Response, error) {

	var resp TriggerCreateResp

	httpReq, err := t.client.NewRequest(http.MethodPost, "/v1/triggers", req)
	if err != nil {
		return nil, nil, err
	}

	httpResp, err := t.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, nil, err
	}

	return &resp, httpResp, nil
}

type TriggerDeleteReq struct {
	ID string `json:"id"`
}

type TriggerDeleteResp struct{}

func (t *Triggers) Delete(ctx context.Context, req *TriggerDeleteReq) (*Response, error) {

	httpReq, err := t.client.NewRequest(http.MethodDelete, "/v1/triggers/"+req.ID, nil)
	if err != nil {
		return nil, err
	}

	httpResp, err := t.client.Do(ctx, httpReq, nil)
	if err != nil {
		return httpResp, err
	}

	return httpResp, nil
}

type TriggersGetReq struct {
	ID string `json:"id"`
}

type TriggersGetResp struct {
	Trigger *Trigger `json:"trigger"`
}

func (t *Triggers) Get(ctx context.Context, req *TriggersGetReq) (*TriggersGetResp, *Response, error) {

	var resp TriggersGetResp

	httpReq, err := t.client.NewRequest(http.MethodGet, "/v1/triggers/"+req.ID, nil)
	if err != nil {
		return nil, nil, err
	}

	httpResp, err := t.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, httpResp, err
	}

	return &resp, httpResp, nil
}

type TriggerListReq struct{}

type TriggerListResp struct {
	Triggers []*TriggerStub `json:"triggers"`
}

func (t *Triggers) List(ctx context.Context, _ *TriggerListReq) (*TriggerListResp, *Response, error) {

	var resp TriggerListResp

	httpReq, err := t.client.NewRequest(http.MethodGet, "/v1/triggers", nil)
	if err != nil {
		return nil, nil, err
	}

	httpResp, err := t.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, httpResp, err
	}

	return &resp, httpResp, nil
}
