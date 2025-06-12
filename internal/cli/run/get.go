package run

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/oklog/ulid/v2"
	"github.com/urfave/cli/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/cli/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api"
)

func getCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Category:  "run",
		Usage:     "Get the detail of a Nomad Pipeline run",
		Args:      true,
		UsageText: "nomad-pipeline runs get [options] [run-id]",
		Flags:     helper.ClientFlags(),
		Action: func(cliCtx *cli.Context) error {

			if numArgs := cliCtx.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(getCommandCLIErrorMsg,
					fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			id, err := ulid.Parse(cliCtx.Args().First())
			if err != nil {
				return cli.Exit(helper.FormatError(getCommandCLIErrorMsg, err), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cliCtx))

			req := api.RunGetReq{ID: id}

			resp, _, err := client.Runs().Get(context.Background(), &req)
			if err != nil {
				return cli.Exit(helper.FormatError(getCommandCLIErrorMsg, err), 1)
			}

			outputRun(cliCtx, resp.Run)
			return nil
		},
	}
}

func outputRun(cliCtx *cli.Context, run *api.Run) {

	_, _ = fmt.Fprint(cliCtx.App.Writer, helper.FormatKV([]string{
		fmt.Sprintf("ID|%s", run.ID),
		fmt.Sprintf("Flow ID|%s", run.FlowID),
		fmt.Sprintf("Status|%v", run.Status),
		fmt.Sprintf("Start Time|%s", helper.FormatTime(run.StartTime)),
		fmt.Sprintf("End Time|%s", helper.FormatTime(run.EndTime)),
	}))
	_, _ = fmt.Fprintf(cliCtx.App.Writer, "\n\n")

	for _, job := range run.Jobs {

		bold := color.New(color.FgWhite, color.Bold)
		_, _ = bold.Fprintf(cliCtx.App.Writer, "Job %q has status %q\n\n", job.ID, job.Status)

		_, _ = fmt.Fprintf(cliCtx.App.Writer, "Details:\n")

		if len(job.Inline) > 0 {

			_, _ = fmt.Fprint(cliCtx.App.Writer, helper.FormatKV([]string{
				fmt.Sprintf("ID|%s", job.ID),
				fmt.Sprintf("Status|%s", job.Status),
				fmt.Sprintf("Start Time|%s", helper.FormatTime(job.StartTime)),
				fmt.Sprintf("End Time|%s", helper.FormatTime(job.EndTime)),
			}))

			_, _ = fmt.Fprintf(cliCtx.App.Writer, "\n\n")
			_, _ = fmt.Fprint(cliCtx.App.Writer, "Inline Steps:\n")

			inlineOut := make([]string, 0, len(job.Inline)+1)
			inlineOut = append(inlineOut, "ID|Status|Exit Code|Start Time|End Time")
			for _, inline := range job.Inline {
				inlineOut = append(inlineOut, fmt.Sprintf(
					"%s|%s|%v|%v|%v",
					inline.ID, inline.Status, inline.ExitCode,
					helper.FormatTime(inline.StartTime), helper.FormatTime(inline.EndTime)))
			}

			_, _ = fmt.Fprint(cliCtx.App.Writer, helper.FormatList(inlineOut))
			_, _ = fmt.Fprintf(cliCtx.App.Writer, "\n")
		}

		if job.Specification != nil {
			_, _ = fmt.Fprint(cliCtx.App.Writer, helper.FormatKV([]string{
				fmt.Sprintf("ID|%s", job.ID),
				fmt.Sprintf("Status|%s", job.Status),
				fmt.Sprintf("Start Time|%s", helper.FormatTime(job.StartTime)),
				fmt.Sprintf("End Time|%s", helper.FormatTime(job.EndTime)),
				fmt.Sprintf("Nomad Job ID|%s", job.Specification.NomadJobID),
				fmt.Sprintf("Nomad Namespace|%s", job.Specification.NomadNamespace),
			}))
			_, _ = fmt.Fprintf(cliCtx.App.Writer, "\n\n")
		}
	}
}
