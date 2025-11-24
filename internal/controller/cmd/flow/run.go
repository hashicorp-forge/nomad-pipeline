package flow

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/run"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

func runCommand() *cli.Command {
	return &cli.Command{
		Name:      "run",
		Category:  "flow",
		Usage:     "Run an instance of a Nomad Pipeline flow",
		UsageText: "nomad-pipeline flow run [options] [flow-id]",
		Flags:     runCommandFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(runCommandCLIErrorMsg,
					fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			cliVars := map[string]any{}

			for _, cliVar := range cmd.StringSlice("var") {

				splitVar := strings.Split(cliVar, "=")
				if len(splitVar) != 2 {
					return cli.Exit(helper.FormatError(runCommandCLIErrorMsg,
						fmt.Errorf("invalid variable format: %s, expected key=value", cliVar)), 1)
				}

				//
				variableNamespace := strings.Split(splitVar[0], ".")

				switch len(variableNamespace) {
				case 2:
					if _, ok := cliVars[variableNamespace[0]]; !ok {
						cliVars[variableNamespace[0]] = map[string]any{}
					}
					nsMap := cliVars[variableNamespace[0]].(map[string]any)
					nsMap[variableNamespace[1]] = splitVar[1]
					cliVars[variableNamespace[0]] = nsMap
				case 1:
					cliVars[splitVar[0]] = splitVar[1]
				default:
					return cli.Exit(helper.FormatError(runCommandCLIErrorMsg,
						fmt.Errorf("invalid variable key format: %s, expected key or namespace.key", splitVar[0])), 1)
				}
			}

			req := api.FlowRunReq{
				ID:        cmd.Args().First(),
				Variables: cliVars,
			}

			resp, _, err := client.Flows().Run(ctx, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(runCommandCLIErrorMsg, err), 1)
			}

			if cmd.Bool("monitor") {
				if err := run.MonitorRun(ctx, client, resp.RunID); err != nil {
					return cli.Exit(helper.FormatError(runCommandCLIErrorMsg, err), 1)
				}
			} else {
				_, _ = fmt.Fprint(cmd.Writer, helper.FormatKV([]string{
					"Message|Successfully triggered flow run",
					fmt.Sprintf("Flow ID|%s", cmd.Args().First()),
					fmt.Sprintf("Run ID|%s", resp.RunID),
				}))
				_, _ = fmt.Fprintf(cmd.Writer, "\n")
			}
			return nil
		},
	}
}

func runCommandFlags() []cli.Flag {
	return append(
		helper.ClientFlags(true),
		[]cli.Flag{
			&cli.StringSliceFlag{
				Name:    "var",
				Aliases: []string{"v"},
				Usage:   "Set a variable for the flow run (key=value)",
			},
			&cli.BoolFlag{
				Name:  "monitor",
				Usage: "Monitor the flow run until completion",
			},
		}...,
	)
}
