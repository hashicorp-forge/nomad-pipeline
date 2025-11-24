package context

import (
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type Context struct {
	NomadPipeline *NomadPipelineContext

	Specifications       []*SpecificationContext
	specificationTracker map[string]int

	Inline *InlineContext

	Variables map[string]any
}

type NomadPipelineContext struct {
	FlowID     string
	FlowType   string
	Namespace  string
	RunID      ulid.ULID
	Status     string
	Trigger    string
	CreateTime time.Time
	StartTime  time.Time
	EndTime    time.Time
}

type SpecificationContext struct {
	ID                string
	Status            string
	StartTime         time.Time
	EndTime           time.Time
	NomadJobID        string
	NomadJobNamespace string
}

type InlineContext struct {
	ID    string
	Steps []*StepContext

	// stepTracker maps step IDs to their index in the Steps slice. This
	// provides fast lookup and modification access to the steps in the array
	// without needing to loop through them each time.
	stepTracker map[string]int
}

type StepContext struct {
	ID        string
	Status    string
	ExitCode  int
	StartTime time.Time
	EndTime   time.Time
}

func New(runID ulid.ULID, trigger string, flow *state.Flow, vars map[string]any) *Context {

	ctx := &Context{
		NomadPipeline: &NomadPipelineContext{
			FlowID:     flow.ID,
			FlowType:   flow.Type(),
			Namespace:  flow.Namespace,
			RunID:      runID,
			Status:     state.RunStatusPending,
			Trigger:    trigger,
			CreateTime: time.Now(),
		},
		Variables: vars,
	}

	if flow.Specification != nil {

		ctx.Specifications = make([]*SpecificationContext, 0, len(flow.Specification))
		ctx.specificationTracker = make(map[string]int, len(flow.Specification))

		for i, spec := range flow.Specification {

			ctx.specificationTracker[spec.ID] = i

			ctx.Specifications = append(
				ctx.Specifications,
				&SpecificationContext{
					ID:     spec.ID,
					Status: state.RunStatusPending,
				},
			)
		}
	}

	if flow.Inline != nil {
		ctx.Inline = &InlineContext{
			ID:          flow.Inline.ID,
			Steps:       make([]*StepContext, 0, len(flow.Inline.Steps)),
			stepTracker: make(map[string]int, len(flow.Inline.Steps)),
		}

		for i, step := range flow.Inline.Steps {

			ctx.Inline.stepTracker[step.ID] = i

			ctx.Inline.Steps = append(ctx.Inline.Steps,
				&StepContext{
					ID:       step.ID,
					Status:   state.RunStatusPending,
					ExitCode: -1,
				},
			)
		}
	}

	return ctx
}
