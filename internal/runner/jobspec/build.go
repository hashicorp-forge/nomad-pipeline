package jobspec

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/nomad/api"
	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/helper"
	"github.com/hashicorp-forge/nomad-pipeline/internal/server/state"
)

func BuildStepJob(runID ulid.ULID, flowID string, flowJob *state.FlowJob) *api.Job {

	jobID := GenerateJobID(runID, flowID, flowJob.ID)

	j := api.Job{
		Name:      helper.PointerOf(jobID),
		ID:        helper.PointerOf(jobID),
		Type:      helper.PointerOf(api.JobTypeBatch),
		Namespace: helper.PointerOf(getNamespace(flowJob)),
		TaskGroups: []*api.TaskGroup{
			{
				Name: helper.PointerOf("runner"),
				Tasks: []*api.Task{
					{
						Name:   flowJob.ID,
						Driver: "docker",
						Config: map[string]interface{}{
							"image":   "ubuntu:24.04",
							"command": "sleep",
							"args":    []string{"infinity"},
						},
						Resources: &api.Resources{
							CPU:      helper.PointerOf(flowJob.Resource.CPU),
							MemoryMB: helper.PointerOf(flowJob.Resource.Memory),
						},
					},
				},
			},
		},
	}

	for _, artifact := range flowJob.Artifacts {
		if !strings.HasPrefix(*artifact.RelativeDest, "local") {
			*artifact.RelativeDest = filepath.Join("local", *artifact.RelativeDest)
		}
		j.TaskGroups[0].Tasks[0].Artifacts = append(j.TaskGroups[0].Tasks[0].Artifacts, artifact)
	}

	for _, step := range flowJob.Steps {
		j.TaskGroups[0].Tasks[0].Templates = append(j.TaskGroups[0].Tasks[0].Templates, &api.Template{
			DestPath:     helper.PointerOf(StepScriptPath(step.ID)),
			EmbeddedTmpl: helper.PointerOf(step.Run),
		})
	}

	return &j
}

func GenerateJobID(runID ulid.ULID, jobID, flowJobID string) string {
	return "nmd-pl-" + runID.String() + "-" + jobID + "-" + flowJobID
}

func StepScriptPath(name string) string { return fmt.Sprintf("local/%s", name) }

func getNamespace(flowJob *state.FlowJob) string {
	if flowJob.NomadNamespace != "" {
		return flowJob.NomadNamespace
	}
	return api.DefaultNamespace
}
