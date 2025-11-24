package run

import (
	"context"
	"fmt"
	"strconv"

	"github.com/oklog/ulid/v2"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

func getCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Category:  "run",
		Usage:     "Get the detail of a Nomad Pipeline run",
		UsageText: "nomad-pipeline runs get [options] [run-id]",
		Flags:     helper.ClientFlags(true),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(getCommandCLIErrorMsg,
					fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			id, err := ulid.Parse(cmd.Args().First())
			if err != nil {
				return cli.Exit(helper.FormatError(getCommandCLIErrorMsg, err), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			req := api.RunGetReq{ID: id}

			resp, _, err := client.Runs().Get(ctx, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(getCommandCLIErrorMsg, err), 1)
			}

			outputRun(resp.Run)
			return nil
		},
	}
}

func outputRun(run *api.Run) {
	pterm.DefaultBasicText.Print(runHeader(run))
	pterm.DefaultBasicText.Print("\n")

	if runVars := runVariables(run); runVars != "" {
		pterm.DefaultSection.Print("Variables")
		pterm.DefaultBasicText.Print(runVars)
	}

	pterm.DefaultSection.Print("Steps")
	pterm.DefaultBasicText.Print(runBody(run))
}

func runHeader(run *api.Run) string {
	return helper.FormatKV([]string{
		fmt.Sprintf("ID|%s", run.ID),
		fmt.Sprintf("Namespace|%s", run.Namespace),
		fmt.Sprintf("Flow ID|%s", run.FlowID),
		fmt.Sprintf("Status|%v", colouredRunStatus(run.Status)),
		fmt.Sprintf("Trigger|%s", run.Trigger),
		fmt.Sprintf("Create Time|%v", helper.FormatTime(run.CreateTime)),
		fmt.Sprintf("Start Time|%s", helper.FormatTime(run.StartTime)),
		fmt.Sprintf("End Time|%s", helper.FormatTime(run.EndTime)),
	})
}

func runVariables(run *api.Run) string {
	var body string

	if len(run.Variables) > 0 {
		out := pterm.TableData{{"Name", "Value"}}

		for key, value := range run.Variables {
			out = append(out, []string{
				key,
				fmt.Sprintf("%v", value),
			})
		}

		body, _ = pterm.DefaultTable.WithHasHeader().WithData(out).Srender()
	}

	return body
}

func runBody(run *api.Run) string {

	var body string

	if run.InlineRun != nil {
		out := pterm.TableData{{"ID", "Status", "Exit Code", "Start Time", "End Time"}}

		for _, step := range run.InlineRun.Steps {
			out = append(out, []string{
				step.ID,
				colouredRunStatus(step.Status),
				strconv.Itoa(step.ExitCode),
				helper.FormatTime(step.StartTime),
				helper.FormatTime(step.EndTime),
			})
		}

		body, _ = pterm.DefaultTable.WithHasHeader().WithData(out).Srender()
	}

	if run.SpecRun != nil {
		out := pterm.TableData{{"ID", "Nomad ID", "Nomad Namespace", "Status", "Start Time", "End Time"}}

		for _, spec := range run.Specs {
			out = append(out, []string{
				spec.ID,
				spec.NomadJobID,
				spec.NomadJobNamespace,
				colouredRunStatus(spec.Status),
				helper.FormatTime(spec.StartTime),
				helper.FormatTime(spec.EndTime),
			})
		}

		body, _ = pterm.DefaultTable.WithHasHeader().WithData(out).Srender()
	}

	return body
}

func colouredRunStatus(status string) string {
	switch status {
	case api.RunStatusPending:
		return pterm.Yellow(status)
	case api.RunStatusRunning:
		return pterm.LightMagenta(status)
	case api.RunStatusSuccess:
		return pterm.Green(status)
	case api.RunStatusFailed:
		return pterm.Red(status)
	case api.RunStatusSkipped, api.RunStatusCancelled:
		return pterm.Gray(status)
	default:
		return pterm.White()
	}
}
