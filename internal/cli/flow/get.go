package flow

import (
	"fmt"
	
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/cli/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api"
)

func getCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Category:  "flow",
		Usage:     "Get a Nomad Pipeline flow",
		Args:      true,
		UsageText: "nomad-pipeline flow get [options] [flow-id]",
		Flags:     helper.ClientFlags(),
		Action: func(cliCtx *cli.Context) error {

			if numArgs := cliCtx.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(getCommandCLIErrorMsg, fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cliCtx))

			req := api.FlowsGetReq{ID: cliCtx.Args().First()}

			resp, _, err := client.Flows().Get(cliCtx.Context, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(getCommandCLIErrorMsg, err), 1)
			}

			outputFlow(cliCtx, resp.Flow)
			return nil
		},
	}
}

func outputFlow(cliCtx *cli.Context, flow *api.Flow) {

	_, _ = fmt.Fprint(cliCtx.App.Writer, helper.FormatKV([]string{
		fmt.Sprintf("ID|%s", flow.ID),
	}))
	_, _ = fmt.Fprintf(cliCtx.App.Writer, "\n\n")

	bold := color.New(color.FgWhite, color.Bold)

	for _, job := range flow.Jobs {
		_, _ = bold.Fprintf(cliCtx.App.Writer, "Job %q details:\n\n", job.ID)
		_, _ = fmt.Fprint(cliCtx.App.Writer, helper.FormatKV([]string{
			fmt.Sprintf("ID|%s", job.ID),
			fmt.Sprintf("Type|%s", job.Type()),
			fmt.Sprintf("Nomad Namespace|%s", job.NomadNamespace),
		}))
		_, _ = fmt.Fprintf(cliCtx.App.Writer, "\n\n")

		if len(job.Artifacts) > 0 {
			out := make([]string, 0, len(job.Artifacts)+1)
			out = append(out, "Source|Destination")
			for _, art := range job.Artifacts {
				out = append(out, fmt.Sprintf(
					"%v|%v", *art.GetterSource, *art.RelativeDest))
			}

			_, _ = fmt.Fprintf(cliCtx.App.Writer, "Artifacts:\n")
			_, _ = fmt.Fprint(cliCtx.App.Writer, helper.FormatList(out))
			_, _ = fmt.Fprintf(cliCtx.App.Writer, "\n\n")
		}

		switch job.Type() {
		case api.FlowJobTypeSpecification:
			_, _ = fmt.Fprintf(cliCtx.App.Writer, "---\n")
			_, _ = fmt.Fprintf(cliCtx.App.Writer, job.JobSpecification.Data)
			_, _ = fmt.Fprintf(cliCtx.App.Writer, "---\n\n")
		case api.FlowJobTypeInline:
			for _, step := range job.Steps {
				_, _ = bold.Fprintf(cliCtx.App.Writer, "Inline Step %q details:\n", step.ID)
				_, _ = fmt.Fprintf(cliCtx.App.Writer, "---\n")
				_, _ = fmt.Fprintf(cliCtx.App.Writer, step.Run)
				_, _ = fmt.Fprintf(cliCtx.App.Writer, "---\n\n")
			}
		default:
		}

	}
}
