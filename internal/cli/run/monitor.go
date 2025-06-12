package run

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/urfave/cli/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/cli/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api"
)

func monitorCommand() *cli.Command {
	return &cli.Command{
		Name:      "monitor",
		Category:  "run",
		Usage:     "Monitor a Nomad Pipeline run",
		Args:      false,
		UsageText: "nomad-pipeline run monitor [options] [run-id]",
		Flags:     helper.ClientFlags(),
		Action: func(cliCtx *cli.Context) error {

			if numArgs := cliCtx.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(monitorCommandCLIErrorMsg,
					fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			id, err := ulid.Parse(cliCtx.Args().First())
			if err != nil {
				return cli.Exit(helper.FormatError(monitorCommandCLIErrorMsg, err), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cliCtx))

			resp, _, err := client.Runs().Get(cliCtx.Context, &api.RunGetReq{ID: id})
			if err != nil {
				return cli.Exit(helper.FormatError(monitorCommandCLIErrorMsg, err), 1)
			}

			switch resp.Run.Status {
			case api.RunStatusFailed, api.RunStatusSuccess:
				return monitorCompletedRun(cliCtx, client, resp.Run)
			default:
				return monitorInProgressRun(cliCtx, client, resp.Run)
			}
		},
	}
}

func monitorCompletedRun(cliCtx *cli.Context, client *api.Client, run *api.Run) error {

	for _, job := range run.Jobs {

		if job.Specification != nil {
			if err := monitorJobSpecification(cliCtx, client, run.ID, run.FlowID, job.ID); err != nil {
				return err
			}
			continue
		}

		for _, inlineStep := range job.Inline {

			req := api.RunLogsGetReq{
				ID:     run.ID,
				JobID:  job.ID,
				StepID: inlineStep.ID,
				Type:   "stdout",
			}

			resp, _, err := client.Runs().LogsGet(cliCtx.Context, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(logsCommandCLIErrorMsg, err), 1)
			}

			for _, log := range resp.Logs {
				_, _ = fmt.Fprintf(cliCtx.App.Writer, "%s.%s.stdout: %s\n", job.ID, inlineStep.ID, log)
			}

			req.Type = "stderr"

			resp, _, err = client.Runs().LogsGet(cliCtx.Context, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(logsCommandCLIErrorMsg, err), 1)
			}

			for _, log := range resp.Logs {
				_, _ = fmt.Fprintf(cliCtx.App.Writer, "%s.%s.stderr: %s\n", job.ID, inlineStep.ID, log)
			}
		}
	}
	return nil
}

func monitorJobSpecification(cliCtx *cli.Context, client *api.Client, runID ulid.ULID, flowID, jobID string) error {

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			resp, _, err := client.Runs().Get(cliCtx.Context, &api.RunGetReq{ID: runID})
			if err != nil {
				_, _ = fmt.Fprintf(cliCtx.App.ErrWriter, "error getting run detail: %s\n", err)
				continue
			}

			for _, job := range resp.Run.Jobs {
				if job.ID != jobID {
					continue
				}
				_, _ = fmt.Fprintf(cliCtx.App.ErrWriter, "%s.%s: Nomad job %q in namespace %q has status %q \n",
					flowID, job.Specification.ID, job.Specification.NomadJobID,
					job.Specification.NomadNamespace, job.Status,
				)

				if job.Status == api.RunStatusFailed {
					return errors.New("job failed")
				}
				if job.Status == api.RunStatusSuccess {
					return nil
				}
			}

		case <-cliCtx.Done():
			return nil
		}
	}
}

func monitorInProgressRun(cliCtx *cli.Context, client *api.Client, run *api.Run) error {
	for _, job := range run.Jobs {

		if job.Specification != nil {
			if err := monitorJobSpecification(cliCtx, client, run.ID, run.FlowID, job.ID); err != nil {
				return err
			}
			continue
		}

		for _, inlineStep := range job.Inline {
			if err := monitorInProgressInline(cliCtx, client, run.ID, run.FlowID, job.ID, inlineStep); err != nil {
				return err
			}
		}
	}
	return nil
}

func monitorInProgressInline(cliCtx *cli.Context, client *api.Client, runID ulid.ULID, flowID, jobID string, inline *api.RunJobInline) error {

	logTailFn := func(client *api.Client, req *api.RunLogsTailReq, prefix string, stopCh chan struct{}) {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

	TAIL:
		resp, _, err := client.Runs().LogsTail(ctx, req)
		if err != nil {
			_, _ = fmt.Fprintf(cliCtx.App.ErrWriter, "%s: error streaming logs: %s\n", prefix, err)
			time.Sleep(1 * time.Second)
			goto TAIL
		}

		for {
			select {
			case <-stopCh:
				return
			case err := <-resp.ErrCh:
				_, _ = fmt.Fprintf(cliCtx.App.ErrWriter, "%s: error streaming logs: %s\n", prefix, err)
				return
			case line := <-resp.LogCh:
				_, _ = fmt.Fprintf(cliCtx.App.ErrWriter, "%s: %s\n", prefix, line)
			}
		}
	}

	stopCh := make(chan struct{})

	go logTailFn(
		client,
		&api.RunLogsTailReq{
			ID:     runID,
			JobID:  jobID,
			StepID: inline.ID,
			Type:   "stdout",
		},
		fmt.Sprintf("%s.%s.%s.stdout", flowID, jobID, inline.ID),
		stopCh,
	)

	go logTailFn(
		client,
		&api.RunLogsTailReq{
			ID:     runID,
			JobID:  jobID,
			StepID: inline.ID,
			Type:   "stderr",
		},
		fmt.Sprintf("%s.%s.%s.stderr", flowID, jobID, inline.ID),
		stopCh,
	)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			req := api.RunGetReq{ID: runID}
			resp, _, err := client.Runs().Get(cliCtx.Context, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(logsCommandCLIErrorMsg, err), 1)
			}

			for _, job := range resp.Run.Jobs {
				if job.ID != jobID {
					continue
				}
				for _, inlineStep := range job.Inline {
					if inlineStep.ID != inline.ID {
						continue
					}
					if inlineStep.Status == api.RunStatusFailed || inlineStep.Status == api.RunStatusSuccess {
						close(stopCh)
						return nil
					}
					continue
				}
			}
		case <-cliCtx.Done():
			close(stopCh)
		}
	}
}
