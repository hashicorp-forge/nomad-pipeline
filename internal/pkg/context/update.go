package context

import (
	"fmt"
	"time"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

func (c *Context) CancelRun() {

	t := time.Now()

	c.NomadPipeline.EndTime = t
	c.NomadPipeline.Status = state.RunStatusCancelled

	for _, specCtx := range c.Specifications {
		switch specCtx.Status {
		case state.RunStatusRunning, state.RunStatusPending:
			specCtx.Status = state.RunStatusCancelled
			specCtx.EndTime = t
		default:
		}
	}

	for _, stepCtx := range c.Inline.Steps {
		switch stepCtx.Status {
		case state.RunStatusRunning, state.RunStatusPending:
			stepCtx.Status = state.RunStatusCancelled
			stepCtx.EndTime = t
		default:
		}
	}
}

func (c *Context) EndRun(status string) {
	c.NomadPipeline.EndTime = time.Now()
	c.NomadPipeline.Status = status
}

func (c *Context) StartRun() {
	c.NomadPipeline.StartTime = time.Now()
	c.NomadPipeline.Status = state.RunStatusRunning
}

func (c *Context) StartSpecification(specID, nomadNS, nomadJobID string) {

	idx := c.specificationTracker[specID]

	c.Specifications[idx].Status = state.RunStatusRunning
	c.Specifications[idx].StartTime = time.Now()
	c.Specifications[idx].NomadJobID = nomadJobID
	c.Specifications[idx].NomadJobNamespace = nomadNS
}

func (c *Context) EndSpecification(specID, status string) {

	idx := c.specificationTracker[specID]

	// Always set the status.
	c.Specifications[idx].Status = status

	// If the specification was skipped, do not set the end time. This is
	// because it was never started and so both the create and end time should
	// be empty/zero.
	switch status {
	case state.RunStatusSkipped:
	default:
		c.Specifications[idx].EndTime = time.Now()
	}
}

func (c *Context) StartInlineStep(stepID string) {

	idx := c.Inline.stepTracker[stepID]

	c.Inline.Steps[idx].Status = state.RunStatusRunning
	c.Inline.Steps[idx].StartTime = time.Now()
}

func (c *Context) EndInlineStep(stepID, status string, exitCode int) {

	idx := c.Inline.stepTracker[stepID]

	c.Inline.Steps[idx].Status = status
	c.Inline.Steps[idx].ExitCode = exitCode

	// If the step was skipped, do not set the end time. This is because it was
	// never started and so both the create and end time should be empty/zero.
	switch status {
	case state.RunStatusSkipped:
	default:
		c.Inline.Steps[idx].EndTime = time.Now()
	}
}

func (c *Context) GetContext() *Context { return c }

func (c *Context) Run() *state.Run {
	if c == nil {
		return nil
	}

	np := c.NomadPipeline

	run := &state.Run{
		ID:         np.RunID,
		FlowID:     np.FlowID,
		Namespace:  np.Namespace,
		Status:     np.Status,
		Trigger:    np.Trigger,
		CreateTime: np.CreateTime,
		StartTime:  np.StartTime,
		EndTime:    np.EndTime,
		Variables:  map[string]any{},
	}

	variables := c.Variables["var"]

	for key, value := range variables.(map[string]any) {
		if key == "trigger" {
			for subKey, subValue := range value.(map[string]any) {
				run.Variables[fmt.Sprintf("%s.%s", key, subKey)] = subValue
			}
		} else {
			run.Variables[key] = value
		}

	}

	if len(c.Specifications) > 0 {
		run.SpecRun = &state.SpecRun{
			Specs: make([]*state.Spec, 0, len(c.Specifications)),
		}

		for _, specCtx := range c.Specifications {
			spec := &state.Spec{
				ID:                specCtx.ID,
				NomadJobID:        specCtx.NomadJobID,
				NomadJobNamespace: specCtx.NomadJobNamespace,
				Status:            specCtx.Status,
				StartTime:         specCtx.StartTime,
				EndTime:           specCtx.EndTime,
			}
			run.SpecRun.Specs = append(run.SpecRun.Specs, spec)
		}
	}

	if c.Inline != nil && len(c.Inline.Steps) > 0 {
		run.InlineRun = &state.InlineRun{
			ID:    c.Inline.ID,
			Steps: make([]*state.InlineStep, 0, len(c.Inline.Steps)),
		}

		for _, stepCtx := range c.Inline.Steps {
			step := &state.InlineStep{
				ID:        stepCtx.ID,
				Status:    stepCtx.Status,
				ExitCode:  stepCtx.ExitCode,
				StartTime: stepCtx.StartTime,
				EndTime:   stepCtx.EndTime,
			}
			run.InlineRun.Steps = append(run.InlineRun.Steps, step)
		}
	}

	return run
}
