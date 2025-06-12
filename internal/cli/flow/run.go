package flow

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/cli/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api"
)

func runCommand() *cli.Command {
	return &cli.Command{
		Name:      "run",
		Category:  "flow",
		Usage:     "Run an instance of a Nomad Pipeline flow",
		Args:      true,
		UsageText: "nomad-pipeline flow run [options] [flow-id]",
		Flags:     helper.ClientFlags(),
		Action: func(cliCtx *cli.Context) error {

			if numArgs := cliCtx.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(runCommandCLIErrorMsg,
					fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cliCtx))

			req := api.FlowRunReq{ID: cliCtx.Args().First()}

			resp, _, err := client.Flows().Run(cliCtx.Context, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(runCommandCLIErrorMsg, err), 1)
			}

			_, _ = fmt.Fprint(cliCtx.App.Writer, helper.FormatKV([]string{
				fmt.Sprint("Message|Successfully triggered flow run"),
				fmt.Sprintf("Flow ID|%s", cliCtx.Args().First()),
				fmt.Sprintf("Run ID|%s", resp.RunID),
			}))
			_, _ = fmt.Fprintf(cliCtx.App.Writer, "\n")
			return nil
		},
	}
}
