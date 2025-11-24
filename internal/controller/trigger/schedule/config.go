package schedule

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type triggerConfig struct {
	Crons []string `hcl:"crons" json:"crons"`
}

func decodeTriggerConfig(trigger *state.Trigger) (*triggerConfig, error) {

	if len(trigger.Source.Config) == 0 {
		return nil, fmt.Errorf("trigger config is empty")
	}

	file, diags := hclsyntax.ParseConfig(trigger.Source.Config, "config.hcl", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL config: %w", diags)
	}

	var cfg triggerConfig
	diags = gohcl.DecodeBody(file.Body, nil, &cfg)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to decode HCL config: %w", diags)
	}

	return &cfg, nil
}
