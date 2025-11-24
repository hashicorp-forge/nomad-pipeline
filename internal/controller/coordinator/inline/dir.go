package inline

import (
	"os"
	"path/filepath"

	"github.com/oklog/ulid/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

func createDataDir(base string, runID ulid.ULID, flow *state.Flow) error {
	for _, step := range flow.Inline.Steps {
		logDir := filepath.Join(
			base,
			flow.Namespace,
			runID.String(),
			step.ID,
			"logs",
		)

		if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
