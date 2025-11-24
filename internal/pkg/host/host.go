package host

import (
	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type RunConfig struct {
	ID            ulid.ULID      `json:"id"`
	Namespace     string         `json:"namespace"`
	JobID         string         `json:"job_id"`
	Flow          *state.Flow    `json:"flow"`
	Variables     map[string]any `json:"variables"`
	JobSteps      []*state.Step  `json:"job_steps"`
	ControllerRPC string         `json:"controller_rpc"`
}
