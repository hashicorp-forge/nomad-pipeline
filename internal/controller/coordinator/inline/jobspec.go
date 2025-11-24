package inline

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/nomad/api"
	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/hcl"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/helper"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/host"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

type jobBuilderReq struct {
	flow    *state.Flow
	runID   ulid.ULID
	evalCtx *hcl.EvalContext
	vars    map[string]any

	rpcAddr string
}

type jobBuilder struct {
	baseDir string
	req     *jobBuilderReq
}

func newJobBuilder(req *jobBuilderReq) *jobBuilder {
	return &jobBuilder{
		baseDir: filepath.Join("local", req.runID.String()),
		req:     req,
	}
}

func (b *jobBuilder) Build() (*api.Job, error) {

	j := api.Job{
		Name:      helper.PointerOf(b.req.runID.String()),
		ID:        helper.PointerOf(b.req.runID.String()),
		Type:      helper.PointerOf(api.JobTypeBatch),
		Namespace: helper.PointerOf(b.getNamespace()),
		TaskGroups: []*api.TaskGroup{
			{
				Name: helper.PointerOf("runner"),
				ReschedulePolicy: &api.ReschedulePolicy{
					Attempts:  helper.PointerOf(0),
					Unlimited: helper.PointerOf(false),
				},
				RestartPolicy: &api.RestartPolicy{
					Attempts: helper.PointerOf(0),
					Mode:     helper.PointerOf("fail"),
				},
				Tasks: []*api.Task{
					{
						Name:   b.req.flow.ID,
						Driver: "docker",
						Config: map[string]any{
							"image":   b.req.flow.Inline.Runner.NomadOnDemand.Image,
							"command": "nomad-pipeline-runner",
							"args":    []string{"job", "run", "-config", filepath.Join(b.baseDir, "runner.json")},
						},
						Resources: b.getResources(),
					},
				},
			},
		},
	}

	for _, artifact := range b.req.flow.Inline.Runner.NomadOnDemand.Artifacts {
		evaluatedOpts, err := b.evalArtifactOptions(artifact.Options)
		if err != nil {
			return nil, err
		}

		taskArtifact := &api.TaskArtifact{
			GetterSource:  &artifact.Source,
			GetterOptions: evaluatedOpts,
		}

		if artifact.Dest != "" {
			taskArtifact.RelativeDest = helper.PointerOf(filepath.Join(b.baseDir, artifact.Dest))
		}

		j.TaskGroups[0].Tasks[0].Artifacts = append(j.TaskGroups[0].Tasks[0].Artifacts, taskArtifact)
	}

	hostCfg := host.RunConfig{
		ID:            b.req.runID,
		Namespace:     b.req.flow.Namespace,
		Flow:          b.req.flow,
		JobID:         b.req.flow.Inline.ID,
		JobSteps:      b.req.flow.Inline.Steps,
		ControllerRPC: b.req.rpcAddr,
		Variables:     b.req.vars,
	}

	data, err := json.Marshal(&hostCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal runner config: %w", err)
	}

	j.TaskGroups[0].Tasks[0].Templates = append(j.TaskGroups[0].Tasks[0].Templates, &api.Template{
		DestPath:     helper.PointerOf(filepath.Join(b.baseDir, "runner.json")),
		EmbeddedTmpl: helper.PointerOf(string(data)),
	})

	return &j, nil
}

func (b *jobBuilder) getNamespace() string {
	if b.req.flow.Inline.Runner.NomadOnDemand.Namespace != "" {
		return b.req.flow.Inline.Runner.NomadOnDemand.Namespace
	}
	return api.DefaultNamespace
}

func (b *jobBuilder) getResources() *api.Resources {
	r := api.Resources{}

	if b.req.flow.Inline.Runner.NomadOnDemand.Resource != nil {
		r.CPU = helper.PointerOf(b.req.flow.Inline.Runner.NomadOnDemand.Resource.CPU)
		r.MemoryMB = helper.PointerOf(b.req.flow.Inline.Runner.NomadOnDemand.Resource.Memory)
	}

	return &r
}

func (b *jobBuilder) evalArtifactOptions(opts map[string]string) (map[string]string, error) {

	res := make(map[string]string, len(opts))

	for k, v := range opts {
		evaled, err := hcl.EvaluateTemplateString(v, b.req.evalCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to eval artifact option %q: %w", k, err)
		}
		res[k] = evaled
	}

	return res, nil
}
